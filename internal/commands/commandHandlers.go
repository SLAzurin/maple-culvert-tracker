package commands

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/slazurin/maple-culvert-tracker/internal/data"
)

var CommandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
	"ping": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "I am alive. You are <@" + i.Member.User.ID + ">, who joined on " + i.Member.JoinedAt.Format(time.RFC822) + ".",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	},
	"culvert": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		// TODO: Fill this with Database query.
		// TODO: this is stupid but this is just sample for now. Remove this double json-ing later
		chartData := []data.ChartMakerPoints{}
		jsonSample := `[{"label":"2/26","score":0},{"label":"3/5","score":1233},{"label":"3/12","score":8000},{"label":"3/19","score":8100},{"label":"3/26","score":5600},{"label":"4/2","score":5500},{"label":"4/9","score":25000}]`
		json.Unmarshal([]byte(jsonSample), &chartData)

		r, err := http.Post("http://"+os.Getenv("CHARTMAKER_HOST")+"/chartmaker", "application/json", bytes.NewBuffer([]byte(jsonSample)))
		if err != nil {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Looks like my `chartmaker` component is broken... ",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
		} else {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "// TODO: character name",
					Files: []*discordgo.File{{Name: i.ID + ".png", Reader: r.Body}},
					// TODO: revisit and see if ephemeral is needed FOR CHARTMAKER NOT PING
					// Flags: discordgo.MessageFlagsEphemeral,
				},
			})
		}
	},
}
