package helpers

import (
	"strings"
	"testing"
	"time"
)

func TestSortZeroScoreStreaks(t *testing.T) {
	latest := time.Date(2026, time.July, 26, 0, 0, 0, 0, time.UTC)
	previous := latest.AddDate(0, 0, -7)
	streaks := []ZeroScoreStreak{
		{MapleCharacterName: "Zeta", ZeroStreak: 2, LatestCulvertDate: latest},
		{MapleCharacterName: "Alpha", ZeroStreak: 2, LatestCulvertDate: latest},
		{MapleCharacterName: "Older", ZeroStreak: 2, LatestCulvertDate: previous},
		{MapleCharacterName: "Longest", ZeroStreak: 3, LatestCulvertDate: latest},
	}

	SortZeroScoreStreaks(streaks)

	want := []string{"Longest", "Alpha", "Zeta", "Older"}
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
		LatestCulvertDate:  date,
	}})

	for _, expected := range []string{"CHARACTER", "ZERO STREAK", "LATEST CULVERT DATE", "Mapler", "4", "2026-07-26"} {
		if !strings.Contains(report, expected) {
			t.Errorf("report missing %q:\n%s", expected, report)
		}
	}
}
