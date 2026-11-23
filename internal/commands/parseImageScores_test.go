package commands

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/slazurin/maple-culvert-tracker/.gen/mapleculverttrackerdb/public/model"
	"github.com/slazurin/maple-culvert-tracker/internal/commands/helpers"
)

func newTestImageRoster(names []string, secondaries map[string][]string) (imageRoster, error) {
	active := make([]model.Characters, 0, len(names))
	for idx, name := range names {
		active = append(active, model.Characters{ID: int64(idx + 1), MapleCharacterName: name})
	}
	return newImageRoster(active, secondaries)
}

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
	checkImageScoreScreenshot(t, "case-2", "Geki", "CanonicalMain")
}

func TestImageScoreResponseIgnoresFuzzyZeroScoreMatches(t *testing.T) {
	checkImageScoreScreenshot(t, "case-3", "", "")
}

func checkImageScoreScreenshot(t *testing.T, fixture, secondary, main string) {
	t.Helper()
	dir := filepath.Join("../../provided/gpq-tests", fixture)
	rosterJSON, err := os.ReadFile(filepath.Join(dir, "members.json"))
	if err != nil {
		t.Fatal(err)
	}
	var active []string
	if err := json.Unmarshal(rosterJSON, &active); err != nil {
		t.Fatal(err)
	}
	var secondaries map[string][]string
	if secondary != "" {
		for idx, name := range active {
			if name == secondary {
				active[idx] = main
				secondaries = map[string][]string{main: {secondary}}
				break
			}
		}
		if secondaries == nil {
			t.Fatalf("fixture roster is missing secondary %q", secondary)
		}
	}
	roster, err := newTestImageRoster(active, secondaries)
	if err != nil {
		t.Fatal(err)
	}
	font, err := helpers.LoadGPQFont()
	if err != nil {
		t.Fatal(err)
	}
	screenshot, err := os.ReadFile(filepath.Join(dir, "1.png"))
	if err != nil {
		t.Fatal(err)
	}
	rows, err := helpers.ParseSmallImageRows(screenshot, font)
	if err != nil {
		t.Fatal(err)
	}
	msg, files := imageScoreResponse(roster, rows, 1)
	got := readImageScoreFiles(t, files)
	expected, err := os.ReadFile(filepath.Join(dir, "1.json"))
	if err != nil {
		t.Fatal(err)
	}
	want := strings.TrimSpace(strings.ReplaceAll(string(expected), "\r\n", "\n"))
	if secondary != "" {
		want = strings.Replace(want, strconv.Quote(secondary), strconv.Quote(main), 1)
	}
	if got["gpq_scores.json"] != want || len(got) != 1 {
		t.Fatalf("response %q, files: %v; want only score JSON %s", msg, got, want)
	}
}

func TestImageScoreResponseReportsDuplicateParent(t *testing.T) {
	roster, err := newTestImageRoster([]string{"MainChar", "Other"}, map[string][]string{"MainChar": {"AltChar", "ThirdChar"}})
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name       string
		last       helpers.ScoreEntry
		imageCount int
	}{
		{"equal main score in one image", helpers.ScoreEntry{Name: "MainChar", Score: 100}, 1},
		{"different main score across images", helpers.ScoreEntry{Name: "MainChar", Score: 90}, 2},
		{"zero main score", helpers.ScoreEntry{Name: "MainChar", Score: 0}, 1},
		{"two secondaries with the same score", helpers.ScoreEntry{Name: "ThirdChar", Score: 100}, 2},
		{"two secondaries with different scores", helpers.ScoreEntry{Name: "ThirdChar", Score: 90}, 2},
		{"same secondary twice", helpers.ScoreEntry{Name: "AltChar", Score: 90}, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			first := []helpers.ScoreEntry{{Name: "AltChar", Score: 100}, {Name: "Other", Score: 80}}
			rows := append(first, tc.last)
			msg, files := imageScoreResponse(roster, rows, tc.imageCount)
			got := readImageScoreFiles(t, files)
			report := got["secondary_character_scores.txt"]
			for _, text := range []string{"MainChar", "AltChar", "100", tc.last.Name, strconv.Itoa(tc.last.Score)} {
				if !strings.Contains(report, text) {
					t.Errorf("duplicate report %q is missing %q", report, text)
				}
			}
			if !strings.Contains(msg, "appears more than once") || !strings.Contains(msg, "No score JSON was generated") {
				t.Errorf("reply did not notify the issuer: %q", msg)
			}
			if _, exported := got["gpq_scores.json"]; exported {
				t.Errorf("duplicate parent returned a score export: %v", got)
			}
		})
	}
}

func TestImageScoreResponseReportsRepeatedMain(t *testing.T) {
	roster, err := newTestImageRoster([]string{"First", "Other"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	rows := []helpers.ScoreEntry{
		{Name: "First", Score: 200}, {Name: "Other", Score: 100}, {Name: "First", Score: 150},
	}
	msg, files := imageScoreResponse(roster, rows, 2)
	got := readImageScoreFiles(t, files)
	if len(got) != 1 || !strings.Contains(got["secondary_character_scores.txt"], "200") || !strings.Contains(got["secondary_character_scores.txt"], "150") {
		t.Errorf("response = %q, files = %v, want both rows for the repeated main and no export", msg, got)
	}
}

func TestImageScoreResponseReportsCrossParentOCRTie(t *testing.T) {
	roster, err := newTestImageRoster([]string{"MainChar", "PrefixBeta"}, map[string][]string{"MainChar": {"PrefixAlpha"}})
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
	_, err := newTestImageRoster([]string{"MainChar", "Líra"}, map[string][]string{"MainChar": {"Iira"}})
	if err == nil || !strings.Contains(err.Error(), "Iira") || !strings.Contains(err.Error(), "Líra") {
		t.Fatalf("expected OCR-equivalent roster names in the error, got %v", err)
	}
}
