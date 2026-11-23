package commands

import (
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/slazurin/maple-culvert-tracker/internal/commands/helpers"
)

func readImageScoreFiles(t *testing.T, files []*discordgo.File) map[string]string {
	t.Helper()
	contents := make(map[string]string)
	for _, file := range files {
		data, err := io.ReadAll(file.Reader)
		if err != nil {
			t.Fatal(err)
		}
		contents[file.Name] = string(data)
	}
	return contents
}

func TestImageScoreResponseMapsSecondaryFromScreenshot(t *testing.T) {
	rosterJSON, err := os.ReadFile("../../provided/gpq-tests/case-2/members.json")
	if err != nil {
		t.Fatal(err)
	}
	var active []string
	if err := json.Unmarshal(rosterJSON, &active); err != nil {
		t.Fatal(err)
	}
	for idx, name := range active {
		if name == "Geki" {
			active[idx] = "CanonicalMain"
		}
	}
	roster, err := newImageRoster(active, map[string][]string{"CanonicalMain": {"Geki"}})
	if err != nil {
		t.Fatal(err)
	}
	font, err := helpers.LoadGPQFont()
	if err != nil {
		t.Fatal(err)
	}
	screenshot, err := os.ReadFile("../../provided/gpq-tests/case-2/1.png")
	if err != nil {
		t.Fatal(err)
	}
	rows, err := helpers.ParseSmallImageRows(screenshot, font)
	if err != nil {
		t.Fatal(err)
	}
	msg, files := imageScoreResponse(roster, rows, 1)
	got := readImageScoreFiles(t, files)
	expected, err := os.ReadFile("../../provided/gpq-tests/case-2/1.json")
	if err != nil {
		t.Fatal(err)
	}
	if want := strings.Replace(string(expected), `"Geki"`, `"CanonicalMain"`, 1); strings.TrimSpace(strings.ReplaceAll(want, "\r\n", "\n")) != got["gpq_scores.json"] || len(got) != 1 {
		t.Fatalf("response %q, files: %v; want score JSON with Geki replaced by CanonicalMain", msg, got)
	}
}

func TestImageScoreResponseReportsDuplicateParent(t *testing.T) {
	roster, err := newImageRoster([]string{"MainChar", "Other"}, map[string][]string{"MainChar": {"AltChar", "ThirdChar"}})
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name     string
		last     helpers.ScoreEntry
		wantJSON bool
	}{
		{"equal main score", helpers.ScoreEntry{Name: "MainChar", Score: 100}, true},
		{"different main score", helpers.ScoreEntry{Name: "MainChar", Score: 90}, false},
		{"zero main score", helpers.ScoreEntry{Name: "MainChar", Score: 0}, false},
		{"distinct secondaries", helpers.ScoreEntry{Name: "ThirdChar", Score: 100}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			first := []helpers.ScoreEntry{{Name: "AltChar", Score: 100}, {Name: "Other", Score: 80}}
			rows := append(first, tc.last)
			msg, files := imageScoreResponse(roster, rows, 2)
			got := readImageScoreFiles(t, files)
			report := got["secondary_character_scores.txt"]
			for _, text := range []string{"MainChar", "AltChar", "100", tc.last.Name} {
				if !strings.Contains(report, text) {
					t.Errorf("duplicate report %q is missing %q", report, text)
				}
			}
			if !strings.Contains(msg, "Multiple character names") {
				t.Errorf("reply did not notify the issuer: %q", msg)
			}
			if tc.wantJSON {
				if want := "{\n    \"MainChar\": 100,\n    \"Other\": 80\n}"; got["gpq_scores.json"] != want {
					t.Errorf("score JSON = %q, want %q", got["gpq_scores.json"], want)
				}
			} else if _, exported := got["gpq_scores.json"]; exported || !strings.Contains(msg, "Conflicting scores") {
				t.Errorf("conflicting scores returned %q and %v", msg, got)
			}
		})
	}
}

func TestImageScoreResponsePreservesRepeatedNameBehavior(t *testing.T) {
	roster, err := newImageRoster([]string{"First", "Other"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	rows := []helpers.ScoreEntry{
		{Name: "Other", Score: 0}, {Name: "First", Score: 200},
		{Name: "Other", Score: 100}, {Name: "First", Score: 150}, {Name: "First", Score: 0},
	}
	_, files := imageScoreResponse(roster, rows, 2)
	got := readImageScoreFiles(t, files)
	if want := "{\n    \"First\": 150,\n    \"Other\": 100\n}"; len(got) != 1 || got["gpq_scores.json"] != want {
		t.Errorf("score JSON and attachments = %v, want only %s", got, want)
	}
}

func TestImageScoreResponseReportsCrossParentOCRTie(t *testing.T) {
	roster, err := newImageRoster([]string{"MainChar", "PrefixBeta"}, map[string][]string{"MainChar": {"PrefixAlpha"}})
	if err != nil {
		t.Fatal(err)
	}
	msg, files := imageScoreResponse(roster, []helpers.ScoreEntry{{Name: "Prefix", Score: 100}}, 1)
	got := readImageScoreFiles(t, files)
	report := got["ambiguous_scores.txt"]
	if len(got) != 1 || !strings.Contains(msg, "more than one active character") || !strings.Contains(report, "PrefixAlpha (MainChar)") || !strings.Contains(report, "PrefixBeta") {
		t.Errorf("ambiguous OCR return = %q, files %v", msg, got)
	}
}

func TestImageRosterRejectsSecondaryConflictingWithTrackedName(t *testing.T) {
	_, err := newImageRoster([]string{"MainChar", "Líra"}, map[string][]string{"MainChar": {"Iira"}})
	if err == nil || !strings.Contains(err.Error(), "Iira") || !strings.Contains(err.Error(), "Líra") {
		t.Fatalf("expected OCR-equivalent roster names in the error, got %v", err)
	}
}
