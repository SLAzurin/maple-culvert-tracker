package commands

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
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

	rows, err := helpers.GetCurrentZeroScoreStreaks(db.DB, latestDate)
	if err != nil {
		log.Println("inactivePlayers: query failed", err)
		content := "Failed to retrieve zero streak data from database. See server logs."
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: &content})
		return
	}
	if len(rows) == 0 {
		content := "No characters currently have a zero streak on the latest available culvert date."
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: &content})
		return
	}

	latestDateStr := latestDate.Format(time.DateOnly)
	content := fmt.Sprintf("Current zero streaks for the latest available culvert date (%s)", latestDateStr)
	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &content,
		Files: []*discordgo.File{{
			Name:   "zero-streaks.txt",
			Reader: strings.NewReader(helpers.FormatZeroScoreStreaks(rows)),
		}},
	})
}
