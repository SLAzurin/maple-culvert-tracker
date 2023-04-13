package commands

import (
	"time"

	"github.com/bwmarrin/discordgo"
)

var CommandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
	"ping": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "This command was run by <@" + i.Member.User.ID + ">, who joined on " + i.Member.JoinedAt.Format(time.RFC822) + ".",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	},
}
