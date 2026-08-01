package helpers

import (
	"database/sql"
	"slices"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
)

// ZeroScoreStreak is a character's consecutive zero-score run ending on the
// requested culvert date.
type ZeroScoreStreak struct {
	MapleCharacterName string
	ZeroStreak         int64
	LatestCulvertDate  time.Time
}

// GetCurrentZeroScoreStreaks returns characters that have a zero score on
// latestDate and counts their consecutive stored zero-score records backwards.
func GetCurrentZeroScoreStreaks(db *sql.DB, latestDate time.Time) ([]ZeroScoreStreak, error) {
	const query = `WITH ranked AS (
		SELECT
			c.id AS character_id,
			c.maple_character_name,
			sc.culvert_date,
			sc.score,
			ROW_NUMBER() OVER (
				PARTITION BY c.id
				ORDER BY sc.culvert_date DESC
			) AS rn
		FROM characters c
		INNER JOIN character_culvert_scores sc ON sc.character_id = c.id
	),
	first_non_zero AS (
		SELECT
			character_id,
			MIN(rn) FILTER (WHERE score <> 0) AS first_non_zero_rn
		FROM ranked
		GROUP BY character_id
	),
	current_zero_streaks AS (
		SELECT
			r.character_id,
			r.maple_character_name,
			COUNT(*) AS zero_streak,
			MAX(r.culvert_date) AS latest_culvert_date
		FROM ranked r
		LEFT JOIN first_non_zero f ON f.character_id = r.character_id
		WHERE r.score = 0
			AND (f.first_non_zero_rn IS NULL OR r.rn < f.first_non_zero_rn)
		GROUP BY r.character_id, r.maple_character_name
		HAVING COUNT(*) > 0
	),
	latest_characters AS (
		SELECT DISTINCT r.character_id
		FROM ranked r
		WHERE r.culvert_date = $1
	)
	SELECT
		cs.maple_character_name,
		cs.zero_streak,
		cs.latest_culvert_date
	FROM current_zero_streaks cs
	INNER JOIN latest_characters lc ON lc.character_id = cs.character_id
	ORDER BY cs.zero_streak DESC, cs.latest_culvert_date DESC, cs.maple_character_name`

	rows, err := db.Query(query, latestDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	streaks := []ZeroScoreStreak{}
	for rows.Next() {
		var streak ZeroScoreStreak
		if err := rows.Scan(&streak.MapleCharacterName, &streak.ZeroStreak, &streak.LatestCulvertDate); err != nil {
			return nil, err
		}
		streaks = append(streaks, streak)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	SortZeroScoreStreaks(streaks)
	return streaks, nil
}

// SortZeroScoreStreaks applies the report's deterministic display ordering.
func SortZeroScoreStreaks(streaks []ZeroScoreStreak) {
	slices.SortStableFunc(streaks, func(a, b ZeroScoreStreak) int {
		if a.ZeroStreak > b.ZeroStreak {
			return -1
		}
		if a.ZeroStreak < b.ZeroStreak {
			return 1
		}
		if !a.LatestCulvertDate.Equal(b.LatestCulvertDate) {
			if a.LatestCulvertDate.After(b.LatestCulvertDate) {
				return -1
			}
			return 1
		}
		if a.MapleCharacterName < b.MapleCharacterName {
			return -1
		}
		if a.MapleCharacterName > b.MapleCharacterName {
			return 1
		}
		return 0
	})
}

// FormatZeroScoreStreaks renders the shared inactive-player report table.
func FormatZeroScoreStreaks(streaks []ZeroScoreStreak) string {
	t := table.NewWriter()
	t.AppendHeader(table.Row{"Character", "Zero Streak", "Latest Culvert Date"})
	for _, streak := range streaks {
		t.AppendRow(table.Row{streak.MapleCharacterName, streak.ZeroStreak, streak.LatestCulvertDate.Format(time.DateOnly)})
	}
	return t.Render()
}
