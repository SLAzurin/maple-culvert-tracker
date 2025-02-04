package commands

//lint:file-ignore ST1001 Dot imports by jet
import (
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
	// . "github.com/go-jet/jet/v2/postgres"
	// . "github.com/slazurin/maple-culvert-tracker/.gen/mapleculverttrackerdb/public/table"
	cmdhelpers "github.com/slazurin/maple-culvert-tracker/internal/commands/helpers"
)

func culvertSummary(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options

	date := ""

	for _, v := range options {
		if v.Name == "date" {
			date = v.StringValue()
		}
	}

	if date == "" {
		date = cmdhelpers.GetCulvertResetDate(time.Now()).Format("2006-01-02")
	}

	d, err := time.Parse("2006-01-02", date) // YYYY-MM-DD
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
	log.Println(d)

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags:   discordgo.MessageFlagsEphemeral,
			Content: "Under construction!",
		},
	})

}
