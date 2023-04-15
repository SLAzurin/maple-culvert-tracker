package commands

import (
	"bytes"
	"net/http"
	"time"

	"github.com/bwmarrin/discordgo"
)

var CommandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
	"ping": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		// TODO Fill this with Database query.
		jsonSample := `[{"label":"2/26","score":0},{"label":"3/5","score":1233},{"label":"3/12","score":8000},{"label":"3/19","score":8100},{"label":"3/26","score":5600},{"label":"4/2","score":5500},{"label":"4/9","score":25000}]`

		r, _ := http.Post("http://localhost:3005/chartmaker", "application/json", bytes.NewBuffer([]byte(jsonSample)))

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Files:   []*discordgo.File{{Name: i.ID + ".png", Reader: r.Body}},
				Content: "This command was run by <@" + i.Member.User.ID + ">, who joined on " + i.Member.JoinedAt.Format(time.RFC822) + ".",
				// TODO: revisit and see if ephemeral is needed
				// Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	},
}
