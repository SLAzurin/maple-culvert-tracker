package commands

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/slazurin/maple-culvert-tracker/internal/commands/helpers"
	"github.com/slazurin/maple-culvert-tracker/internal/db"
)

func inactivePlayers(s *discordgo.Session, i *discordgo.InteractionCreate) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})

	latestDate, err := helpers.GetLatestResetDate(db.DB)
	if err != nil {
		log.Println("inactivePlayers: failed to get latest reset date", err)
		content := "Failed to retrieve latest culvert date. See server logs."
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: &content})
		return
	}

	query := `WITH ranked AS (
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

	type resultRow struct {
		MapleCharacterName string
		ZeroStreak         int64
		LatestCulvertDate  time.Time
	}

	rows := []resultRow{}

	result, err := db.DB.Query(query, latestDate)
	if err != nil {
		log.Println("inactivePlayers: query failed", err)
		content := "Failed to retrieve zero streak data from database. See server logs."
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: &content})
		return
	}
	defer result.Close()

	for result.Next() {
		var row resultRow
		if err := result.Scan(&row.MapleCharacterName, &row.ZeroStreak, &row.LatestCulvertDate); err != nil {
			log.Println("inactivePlayers: scan failed", err)
			content := "Failed to read zero streak data from database. See server logs."
			s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: &content})
			return
		}
		rows = append(rows, row)
	}

	if err := result.Err(); err != nil {
		log.Println("inactivePlayers: rows iteration failed", err)
		content := "Failed to iterate zero streak data from database. See server logs."
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: &content})
		return
	}

	if len(rows) == 0 {
		content := "No characters currently have a zero streak on the latest available culvert date."
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: &content})
		return
	}

	latestDateStr := latestDate.Format(time.DateOnly)
	t := table.NewWriter()
	t.AppendHeader(table.Row{"Character", "Zero Streak", "Latest Culvert Date"})
	for _, row := range rows {
		t.AppendRow(table.Row{row.MapleCharacterName, row.ZeroStreak, row.LatestCulvertDate.Format(time.DateOnly)})
	}

	content := fmt.Sprintf("Current zero streaks for the latest available culvert date (%s)", latestDateStr)
	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &content,
		Files: []*discordgo.File{{
			Name:   "zero-streaks.txt",
			Reader: strings.NewReader(t.Render()),
		}},
	})
}
