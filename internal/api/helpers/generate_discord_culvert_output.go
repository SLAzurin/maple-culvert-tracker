package helpers

import (
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func GenerateDiscordCulvertOutput(chartImageBinData io.ReadCloser, charName string, date string, otherStatsStruct any) *discordgo.InteractionResponseData {
	// date is possibly empty
	// otherStatsStruct to be implemented when we get more stats

	content := strings.Trim(charName+" "+date, " ")

	embeddedData := &discordgo.MessageEmbed{
		Title: content,
		// Convert hex to int here
		// https://www.rapidtables.com/convert/number/hex-to-decimal.html?x=36A2EB
		Color: 3580651,
		Image: &discordgo.MessageEmbedImage{
			URL: "attachment://" + content + ".png",
		},
	}

	charData, err := FetchCharacterData(charName, os.Getenv("MAPLE_REGION"))
	if err == nil {
		embeddedData.Title = strings.Trim(charData.CharacterName+" "+date, " ")
		embeddedData.Thumbnail = &discordgo.MessageEmbedThumbnail{URL: charData.CharacterImgURL}
		embeddedData.Description = charName + " is a Level " + strconv.Itoa(charData.Level) + " " + charData.JobName

		// For more embed examples, visit: https://discordjs.guide/popular-topics/embeds.html#using-the-embed-constructor
	}

	return &discordgo.InteractionResponseData{
		Embeds: []*discordgo.MessageEmbed{embeddedData},
		Files:  []*discordgo.File{{Name: content + ".png", Reader: chartImageBinData}},
	}
}
