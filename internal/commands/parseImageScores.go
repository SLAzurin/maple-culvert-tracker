package commands

import (
	"fmt"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/slazurin/maple-culvert-tracker/.gen/mapleculverttrackerdb/public/model"
	"github.com/slazurin/maple-culvert-tracker/internal/commands/helpers"
)

type imageRoster struct {
	names      []string
	parents    map[string]model.Characters
	exactNames map[string][]string
}

func newImageRoster(active []model.Characters, secondaries map[string][]string) (imageRoster, error) {
	roster := imageRoster{parents: make(map[string]model.Characters), exactNames: make(map[string][]string)}
	add := func(name string, parent model.Characters) {
		roster.parents[name] = parent
		roster.names = append(roster.names, name)
		key := helpers.NormalizeGPQName(name)
		roster.exactNames[key] = append(roster.exactNames[key], name)
	}
	for _, parent := range active {
		add(parent.MapleCharacterName, parent)
	}
	for _, parent := range active {
		for _, secondary := range secondaries[parent.MapleCharacterName] {
			for _, other := range roster.exactNames[helpers.NormalizeGPQName(secondary)] {
				otherParent := roster.parents[other]
				if otherParent.ID != parent.ID || (other == otherParent.MapleCharacterName && strings.EqualFold(other, secondary)) {
					return imageRoster{}, fmt.Errorf("secondary character %q for %q conflicts with %q for %q. Fix the mapping before parsing images", secondary, parent.MapleCharacterName, other, otherParent.MapleCharacterName)
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
}

func (r imageRoster) mergeScores(rows []helpers.ScoreEntry) mergedImageScores {
	var result mergedImageScores
	var groups []duplicateImageScore
	groupPositions := make(map[int64]int)
	unmatchedPositions := make(map[string]int)
	for _, row := range rows {
		matches := r.exactNames[helpers.NormalizeGPQName(row.Name)]
		// The table can include untracked members with zero scores. Fuzzy guesses
		// for those rows must not block an otherwise valid score export.
		if len(matches) == 0 && row.Score > 0 {
			matches = helpers.MatchGPQNames(row.Name, r.names)
		}
		if len(matches) > 0 {
			chosen := matches[0]
			parent := r.parents[chosen]
			ambiguous := false
			for _, match := range matches[1:] {
				ambiguous = ambiguous || r.parents[match].ID != parent.ID
				if match == row.Name {
					chosen = match
				}
			}
			if ambiguous {
				candidates := make([]string, 0, len(matches))
				for _, match := range matches {
					candidates = append(candidates, fmt.Sprintf("%s (%s)", match, r.parents[match].MapleCharacterName))
				}
				result.ambiguous = append(result.ambiguous, ambiguousImageScore{row, candidates})
				continue
			}
			row.Name = chosen
		}
		parent, known := r.parents[row.Name]
		if !known {
			if row.Score > 0 {
				if pos, found := unmatchedPositions[row.Name]; found {
					result.unmatched[pos].Score = row.Score
				} else {
					unmatchedPositions[row.Name] = len(result.unmatched)
					result.unmatched = append(result.unmatched, row)
				}
			}
			continue
		}
		// Keep every row for the parent ID, even when the matched name repeats.
		if pos, found := groupPositions[parent.ID]; found {
			groups[pos].rows = append(groups[pos].rows, row)
		} else {
			groupPositions[parent.ID] = len(groups)
			groups = append(groups, duplicateImageScore{parent.MapleCharacterName, []helpers.ScoreEntry{row}})
		}
	}
	for _, group := range groups {
		if len(group.rows) > 1 {
			result.duplicates = append(result.duplicates, group)
		} else if score := group.rows[0].Score; score > 0 {
			result.scores = append(result.scores, helpers.ScoreEntry{Name: group.parent, Score: score})
		}
	}
	return result
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
