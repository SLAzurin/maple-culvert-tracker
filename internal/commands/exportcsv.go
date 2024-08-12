package commands

import (
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
)

func exportcsv(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Parse discord param character-name
	date := ""
	weeks := int64(8)
	options := i.ApplicationCommandData().Options
	for _, v := range options {
		if v.Name == "date" {
			date = v.StringValue()
		}
		if v.Name == "weeks" {
			weeks = v.IntValue()
		}
	}

	// Validate date format
	if date != "" {
		_, err := time.Parse("2006-01-02", date) // YYYY-MM-DD
		if err != nil {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Invalid date format, should be YYYY-MM-DD",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
			return
		}
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "weeks " + strconv.FormatInt(weeks, 10),
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}
