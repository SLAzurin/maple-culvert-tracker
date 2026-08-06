package helpers

import (
	"database/sql"
	"strings"
	"testing"
	"time"
)

func TestSortZeroScoreStreaks(t *testing.T) {
	latest := time.Date(2026, time.July, 26, 0, 0, 0, 0, time.UTC)
	previous := latest.AddDate(0, 0, -7)
	streaks := []ZeroScoreStreak{
		{MapleCharacterName: "Zeta", ZeroStreak: 2, LastNonZeroDate: sql.NullTime{Time: latest, Valid: true}},
		{MapleCharacterName: "Alpha", ZeroStreak: 2, LastNonZeroDate: sql.NullTime{Time: latest, Valid: true}},
		{MapleCharacterName: "Older", ZeroStreak: 2, LastNonZeroDate: sql.NullTime{Time: previous, Valid: true}},
		{MapleCharacterName: "Longest", ZeroStreak: 3, LastNonZeroDate: sql.NullTime{Time: latest, Valid: true}},
	}

	SortZeroScoreStreaks(streaks)

	want := []string{"Longest", "Alpha", "Older", "Zeta"}
	for i, name := range want {
		if streaks[i].MapleCharacterName != name {
			t.Fatalf("row %d name = %q, want %q", i, streaks[i].MapleCharacterName, name)
		}
	}
}

func TestFormatZeroScoreStreaks(t *testing.T) {
	date := time.Date(2026, time.July, 26, 0, 0, 0, 0, time.UTC)
	report := FormatZeroScoreStreaks([]ZeroScoreStreak{{
		MapleCharacterName: "Mapler",
		ZeroStreak:         4,
		LastNonZeroDate:    sql.NullTime{Time: date, Valid: true},
	}})

	for _, expected := range []string{"CHARACTER", "ZERO STREAK", "LAST NON-ZERO DATE", "Mapler", "4", "2026-07-26"} {
		if !strings.Contains(report, expected) {
			t.Errorf("report missing %q:\n%s", expected, report)
		}
	}
}

func TestFormatZeroScoreStreaksNeverScored(t *testing.T) {
	report := FormatZeroScoreStreaks([]ZeroScoreStreak{{
		MapleCharacterName: "NewMapler",
		ZeroStreak:         3,
	}})

	if !strings.Contains(report, "Never") {
		t.Errorf("report should identify a character with no non-zero score:\n%s", report)
	}
}
