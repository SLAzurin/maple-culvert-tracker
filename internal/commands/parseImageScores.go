package commands

import (
	"fmt"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/slazurin/maple-culvert-tracker/internal/commands/helpers"
)

type imageRoster struct {
	names   []string
	parents map[string]string
	active  map[string]bool
}

func newImageRoster(active []string, secondaries map[string][]string) (imageRoster, error) {
	roster := imageRoster{parents: make(map[string]string), active: make(map[string]bool)}
	byOCRName := make(map[string][]string)
	add := func(name, parent string) {
		roster.parents[name] = parent
		roster.names = append(roster.names, name)
		key := helpers.NormalizeGPQName(name)
		byOCRName[key] = append(byOCRName[key], name)
	}
	for _, name := range active {
		roster.active[name] = true
		add(name, name)
	}
	for _, parent := range active {
		for _, secondary := range secondaries[parent] {
			for _, other := range byOCRName[helpers.NormalizeGPQName(secondary)] {
				if roster.parents[other] != parent || (roster.active[other] && strings.EqualFold(other, secondary)) {
					return imageRoster{}, fmt.Errorf("secondary character %q for %q conflicts with %q for %q. Fix the mapping before parsing images", secondary, parent, other, roster.parents[other])
				}
			}
			add(secondary, parent)
		}
	}
	return roster, nil
}

type ambiguousImageScore struct {
	row        helpers.ScoreEntry
	candidates []string
}

type duplicateImageScore struct {
	parent string
	rows   []helpers.ScoreEntry
}

type mergedImageScores struct {
	scores     []helpers.ScoreEntry
	unmatched  []helpers.ScoreEntry
	ambiguous  []ambiguousImageScore
	duplicates []duplicateImageScore
	conflict   bool
}

func (r imageRoster) mergeScores(rows []helpers.ScoreEntry) mergedImageScores {
	var result mergedImageScores
	var originals []helpers.ScoreEntry
	var positivePositions []int
	originalPositions := make(map[string]int)
	for _, row := range rows {
		matches := helpers.MatchGPQNames(row.Name, r.names)
		if len(matches) > 0 {
			chosen := matches[0]
			parent := r.parents[chosen]
			ambiguous := false
			for _, match := range matches[1:] {
				ambiguous = ambiguous || r.parents[match] != parent
				if match == row.Name {
					chosen = match
				}
			}
			if ambiguous {
				candidates := make([]string, 0, len(matches))
				for _, match := range matches {
					candidates = append(candidates, fmt.Sprintf("%s (%s)", match, r.parents[match]))
				}
				result.ambiguous = append(result.ambiguous, ambiguousImageScore{row, candidates})
				continue
			}
			row.Name = chosen
		}
		pos, found := originalPositions[row.Name]
		if !found {
			pos = len(originals)
			originalPositions[row.Name] = pos
			originals = append(originals, helpers.ScoreEntry{Name: row.Name})
		}
		// Preserve the first positive row's position and the last positive score.
		if row.Score > 0 {
			if originals[pos].Score == 0 {
				positivePositions = append(positivePositions, pos)
			}
			originals[pos].Score = row.Score
		}
	}

	var groups []duplicateImageScore
	groupPositions := make(map[string]int)
	for _, row := range originals {
		parent := r.parentName(row.Name)
		if pos, found := groupPositions[parent]; found {
			groups[pos].rows = append(groups[pos].rows, row)
		} else {
			groupPositions[parent] = len(groups)
			groups = append(groups, duplicateImageScore{parent, []helpers.ScoreEntry{row}})
		}
	}
	for _, group := range groups {
		score := group.rows[0].Score
		if len(group.rows) > 1 {
			result.duplicates = append(result.duplicates, group)
			for _, row := range group.rows[1:] {
				result.conflict = result.conflict || row.Score != score
			}
		}
	}
	seen := make(map[string]bool)
	for _, pos := range positivePositions {
		row := originals[pos]
		parent := r.parentName(row.Name)
		if !seen[parent] {
			seen[parent] = true
			result.scores = append(result.scores, helpers.ScoreEntry{Name: parent, Score: row.Score})
		}
	}
	result.unmatched = collectUnmatchedScores(result.scores, r.active)
	return result
}

func (r imageRoster) parentName(name string) string {
	if parent := r.parents[name]; parent != "" {
		return parent
	}
	return name
}

func formatDuplicateImageScores(groups []duplicateImageScore) string {
	t := table.NewWriter()
	t.AppendHeader(table.Row{"Active Character", "Matched Name", "Score"})
	for _, group := range groups {
		for _, row := range group.rows {
			t.AppendRow(table.Row{group.parent, row.Name, row.Score})
		}
	}
	return t.Render()
}

func formatAmbiguousImageScores(scores []ambiguousImageScore) string {
	t := table.NewWriter()
	t.AppendHeader(table.Row{"Decoded Name", "Score", "Matches (Active Character)"})
	for _, score := range scores {
		t.AppendRow(table.Row{score.row.Name, score.row.Score, strings.Join(score.candidates, ", ")})
	}
	return t.Render()
}
