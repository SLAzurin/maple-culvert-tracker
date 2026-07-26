package helpers

import (
	"testing"
	"time"
)

func d(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

func score(name string, date time.Time, value int32) CharacterScoreRow {
	return CharacterScoreRow{CharacterName: name, CulvertDate: date, Score: value}
}

func metricByName(metrics []PersonalBestRankMetric, name string) (PersonalBestRankMetric, bool) {
	for _, m := range metrics {
		if m.MapleCharacterName == name {
			return m, true
		}
	}
	return PersonalBestRankMetric{}, false
}

func TestComputePersonalBestRankMetrics_NewPBMovesRanks(t *testing.T) {
	w1 := d(2025, 10, 1)
	w2 := d(2025, 10, 8)
	scores := []CharacterScoreRow{
		score("Alice", w1, 1000),
		score("Bob", w1, 900),
		score("Alice", w2, 1000),
		score("Bob", w2, 1100),
	}

	got := ComputePersonalBestRankMetrics(scores, PersonalBestRankWeeks)
	if len(got) != 2 {
		t.Fatalf("expected 2 metrics, got %d", len(got))
	}
	if got[0].MapleCharacterName != "Bob" || got[0].Pos != 1 {
		t.Fatalf("expected Bob at pos 1, got %+v", got[0])
	}
	if got[1].MapleCharacterName != "Alice" || got[1].Pos != 2 {
		t.Fatalf("expected Alice at pos 2, got %+v", got[1])
	}

	bob, _ := metricByName(got, "Bob")
	if !bob.HasPrevRank || bob.DeltaPos != 1 {
		t.Fatalf("Bob should move up +1, got %+v", bob)
	}
	alice, _ := metricByName(got, "Alice")
	if !alice.HasPrevRank || alice.DeltaPos != -1 {
		t.Fatalf("Alice should move down -1, got %+v", alice)
	}
}

func TestComputePersonalBestRankMetrics_StreakHoldAndReset(t *testing.T) {
	w1 := d(2025, 10, 1)
	w2 := d(2025, 10, 8)
	w3 := d(2025, 10, 15)
	scores := []CharacterScoreRow{
		score("Alice", w1, 1000),
		score("Bob", w1, 900),
		score("Alice", w2, 1000),
		score("Bob", w2, 900),
		score("Alice", w3, 1000),
		score("Bob", w3, 1200),
	}

	got := ComputePersonalBestRankMetrics(scores, PersonalBestRankWeeks)
	alice, ok := metricByName(got, "Alice")
	if !ok {
		t.Fatal("missing Alice")
	}
	if alice.Streak != 1 || alice.Pos != 2 {
		t.Fatalf("Alice streak should reset to 1 at pos 2, got %+v", alice)
	}
	bob, ok := metricByName(got, "Bob")
	if !ok {
		t.Fatal("missing Bob")
	}
	if bob.Streak != 1 || bob.Pos != 1 {
		t.Fatalf("Bob should be new #1 with streak 1, got %+v", bob)
	}

	// Held #1 for 3 weeks
	holdScores := []CharacterScoreRow{
		score("Alice", w1, 1000),
		score("Bob", w1, 900),
		score("Alice", w2, 1000),
		score("Bob", w2, 900),
		score("Alice", w3, 1000),
		score("Bob", w3, 900),
	}
	held := ComputePersonalBestRankMetrics(holdScores, PersonalBestRankWeeks)
	alice, _ = metricByName(held, "Alice")
	if alice.Streak != 3 || alice.Pos != 1 {
		t.Fatalf("Alice should hold #1 for 3 weeks, got %+v", alice)
	}
	if alice.DeltaPos != 0 {
		t.Fatalf("Alice Δpos should be 0, got %+v", alice)
	}
}

func TestComputePersonalBestRankMetrics_NewEntrant(t *testing.T) {
	w1 := d(2025, 10, 1)
	w2 := d(2025, 10, 8)
	scores := []CharacterScoreRow{
		score("Alice", w1, 1000),
		score("Alice", w2, 1000),
		score("Carol", w2, 800),
	}

	got := ComputePersonalBestRankMetrics(scores, PersonalBestRankWeeks)
	carol, ok := metricByName(got, "Carol")
	if !ok {
		t.Fatal("missing Carol")
	}
	if carol.HasPrevRank {
		t.Fatalf("Carol should have no previous rank, got %+v", carol)
	}
	if FormatPersonalBestDeltaPos(carol.DeltaPos, carol.HasPrevRank) != "—" {
		t.Fatalf("new entrant Δpos display should be —")
	}
}

func TestComputePersonalBestRankMetrics_NameTieBreak(t *testing.T) {
	w1 := d(2025, 10, 1)
	scores := []CharacterScoreRow{
		score("Zoe", w1, 1000),
		score("Amy", w1, 1000),
	}

	got := ComputePersonalBestRankMetrics(scores, PersonalBestRankWeeks)
	if len(got) != 2 {
		t.Fatalf("expected 2, got %d", len(got))
	}
	if got[0].MapleCharacterName != "Amy" || got[0].Pos != 1 {
		t.Fatalf("Amy should rank above Zoe on name tie-break, got %+v", got[0])
	}
	if got[1].MapleCharacterName != "Zoe" || got[1].Pos != 2 {
		t.Fatalf("Zoe should be pos 2, got %+v", got[1])
	}
}

func TestComputePersonalBestRankMetrics_OnlyLastMaxWeeks(t *testing.T) {
	var scores []CharacterScoreRow
	start := d(2025, 1, 1)
	for i := 0; i < 60; i++ {
		week := start.AddDate(0, 0, 7*i)
		// First 8 weeks (outside last-52 of 60): Alice dominates.
		// Remaining weeks: Bob dominates with lower absolute scores than Alice's old PB.
		if i < 8 {
			scores = append(scores, score("Alice", week, 2000), score("Bob", week, 100))
		} else {
			scores = append(scores, score("Alice", week, 500), score("Bob", week, 1500))
		}
	}

	got := ComputePersonalBestRankMetrics(scores, PersonalBestRankWeeks)
	bob, ok := metricByName(got, "Bob")
	if !ok {
		t.Fatal("missing Bob")
	}
	alice, ok := metricByName(got, "Alice")
	if !ok {
		t.Fatal("missing Alice")
	}

	// Within last 52 weeks, Alice's early 2000 PBs are outside the window,
	// so Bob's 1500 should rank first.
	if bob.Pos != 1 || bob.PersonalBest != 1500 {
		t.Fatalf("Bob should be #1 with PB 1500 in 52-week window, got %+v", bob)
	}
	if alice.Pos != 2 || alice.PersonalBest != 500 {
		t.Fatalf("Alice should be #2 with PB 500 in 52-week window, got %+v", alice)
	}
}

func TestFormatPersonalBestStreak_OverOneYear(t *testing.T) {
	if FormatPersonalBestStreak(51) != "51" {
		t.Fatalf("expected 51, got %q", FormatPersonalBestStreak(51))
	}
	if FormatPersonalBestStreak(52) != "over 1 year" {
		t.Fatalf("expected over 1 year for 52, got %q", FormatPersonalBestStreak(52))
	}
	if FormatPersonalBestStreak(53) != "over 1 year" {
		t.Fatalf("expected over 1 year for 53, got %q", FormatPersonalBestStreak(53))
	}
}

func TestComputePersonalBestRankMetrics_FullWindowStreakIsOverOneYear(t *testing.T) {
	var scores []CharacterScoreRow
	start := d(2025, 1, 1)
	for i := 0; i < PersonalBestRankWeeks; i++ {
		week := start.AddDate(0, 0, 7*i)
		scores = append(scores, score("Alice", week, 1000), score("Bob", week, 900))
	}

	got := ComputePersonalBestRankMetrics(scores, PersonalBestRankWeeks)
	alice, ok := metricByName(got, "Alice")
	if !ok {
		t.Fatal("missing Alice")
	}
	if alice.Streak != PersonalBestRankWeeks {
		t.Fatalf("expected streak %d, got %d", PersonalBestRankWeeks, alice.Streak)
	}
	if FormatPersonalBestStreak(alice.Streak) != "over 1 year" {
		t.Fatalf("expected over 1 year display, got %q", FormatPersonalBestStreak(alice.Streak))
	}
}

func TestFormatPersonalBestDeltaPos(t *testing.T) {
	if FormatPersonalBestDeltaPos(2, true) != "+2" {
		t.Fatalf("expected +2")
	}
	if FormatPersonalBestDeltaPos(-3, true) != "-3" {
		t.Fatalf("expected -3")
	}
	if FormatPersonalBestDeltaPos(0, true) != "0" {
		t.Fatalf("expected 0")
	}
	if FormatPersonalBestDeltaPos(0, false) != "—" {
		t.Fatalf("expected — for new entrant")
	}
}

func TestFormatPersonalBestPos(t *testing.T) {
	if got := FormatPersonalBestPos(1, 0, true, 1); got != "1" {
		t.Fatalf("expected 1 when delta 0 and streak 1, got %q", got)
	}
	if got := FormatPersonalBestPos(2, 3, true, 1); got != "2 (+3)" {
		t.Fatalf("expected 2 (+3), got %q", got)
	}
	if got := FormatPersonalBestPos(4, -1, true, 1); got != "4 (-1)" {
		t.Fatalf("expected 4 (-1), got %q", got)
	}
	if got := FormatPersonalBestPos(5, 0, false, 1); got != "5 (—)" {
		t.Fatalf("expected 5 (—) for new entrant, got %q", got)
	}
	if got := FormatPersonalBestPos(1, 0, true, 3); got != "1 ×3" {
		t.Fatalf("expected 1 ×3, got %q", got)
	}
	if got := FormatPersonalBestPos(2, 1, true, 4); got != "2 (+1) ×4" {
		t.Fatalf("expected 2 (+1) ×4, got %q", got)
	}
	if got := FormatPersonalBestPos(1, 0, true, 52); got != "1 ×over 1 year" {
		t.Fatalf("expected 1 ×over 1 year, got %q", got)
	}
}
