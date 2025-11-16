package commands

import (
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/slazurin/maple-culvert-tracker/internal/apiredis"
	"github.com/slazurin/maple-culvert-tracker/internal/commands/helpers"
	"github.com/slazurin/maple-culvert-tracker/internal/db"
)

// I was overcaffeinated when writing this

func sniffOutRats(s *discordgo.Session, i *discordgo.InteractionCreate) {
	var err error

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})

	options := i.ApplicationCommandData().Options

	weeks := 8
	threshold := float64(3) / float64(10)
	typeshit := "zero" // or threshold

	for _, v := range options {
		if v.Name == "weeks" {
			weeks = int(v.IntValue())
		}
		if v.Name == "weeks-percentage-threshold" {
			percentage := v.IntValue()
			threshold = float64(percentage) / float64(100)
		}
		if v.Name == "value-as-offense" {
			typeshit = v.StringValue()
		}
	}

	date := helpers.GetCulvertResetDate(time.Now()).Format("2006-01-02")

	characters, err := helpers.GetActiveCharacters(apiredis.RedisDB, db.DB)
	if err != nil {
		content := "Internal error: Failed GetActiveCharacters, see server logs"
		log.Println(err)
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		return
	}

	foundYouYoureFucked, err := helpers.GetStinkyRats(db.DB, *characters, date, weeks, threshold, typeshit)
	if err != nil {
		content := "Error finding rats, see server logs"
		log.Println("sniffOutRats.go", err)
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
	}

	// Done all chars
	if len(foundYouYoureFucked) > 0 {
		content := "Here is the list of rats"
		contentInner := ""
		for _, v := range foundYouYoureFucked {
			contentInner += v.SixSeven + " has a high amount roller coaster pattern count! " + strconv.Itoa(v.LWeeks) + "/" + strconv.Itoa(v.WWeeks) + "\n"
		}
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &content,
			Files: []*discordgo.File{
				{
					Name:   "message.txt",
					Reader: strings.NewReader(contentInner),
				},
			},
		})
		return
	}
	content := "Squeaky clean! No rats!"
	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &content,
	})
}
