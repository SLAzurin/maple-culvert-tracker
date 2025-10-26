package commands

import (
	"log"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

func submitScores(s *discordgo.Session, i *discordgo.InteractionCreate) {
	var err error
	options := i.ApplicationCommandData().Options
	if len(options) != 3 {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Invalid number of options provided!",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	culvertDateStr := ""
	var culvertDate time.Time
	overwriteExisting := false

	for _, v := range options {
		if v.Name == "culvert-date" {
			culvertDateStr = strings.Trim(v.StringValue(), " ")
			culvertDate, err = time.Parse(culvertDateStr, "2006-01-02")
			if err != nil {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "Invalid date format provided! Please use YYYY-MM-DD.",
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				})
				return
			}
		}
		if v.Name == "scores-attachment" {
			log.Println("TODO: scores attachment")
		}
		if v.Name == "overwrite-existing" {
			overwriteExisting = v.BoolValue()
		}
	}

	log.Println(culvertDate, overwriteExisting)

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Scores submission feature is not yet implemented.",
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}
