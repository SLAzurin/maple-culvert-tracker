package helpers

import (
	"io"
	"math/rand/v2"
	"os"

	"github.com/bwmarrin/discordgo"
	"github.com/slazurin/maple-culvert-tracker/internal/apiredis"
)

var randomFluffDuelText = []string{
	"It's just a hands diff",
	"Are you sure you popped everything?",
	"Skill issue",
	"🤏 Close",
	"Gears in #flex but scores at #fails",
	"Too much grass touching will do that to your score",
	"This your bossing mule?",
	"How long does your party wait for you to blue dot",
	"Too much janus. Should've sent it on 2nd mastery",
}

func getRandomFluffDuelText(yourWin bool, yourChar string, theirChar string) string {
	randomNum := rand.IntN(len(randomFluffDuelText) - 1)
	if yourWin {
		return randomFluffDuelText[randomNum] + " " + theirChar
	}
	return randomFluffDuelText[randomNum] + " " + yourChar
}

func GenerateDiscordCulvertDuelOutput(chartImageBinData io.ReadCloser, yourWin bool, yourCharName string, theirCharName string, otherStatsStruct any) (*discordgo.InteractionResponseData, *os.File) {
	// date is possibly empty
	// otherStatsStruct to be implemented when we get more stats
	title := yourCharName + " wins against " + theirCharName
	// thumbnail := "attachment://outcome.webp"
	thumbnail := apiredis.OPTIONAL_CONF_CULVERT_DUEL_THUMBNAIL_URL.GetWithDefault(apiredis.RedisDB, os.Getenv("CULVERT_DUEL_THUMBNAIL_URL"))
	// filename := "./backend_static/victory.webp"
	if !yourWin {
		// filename = "./backend_static/defeat.webp"
		title = theirCharName + " wins against " + yourCharName
	}

	embeddedData := &discordgo.MessageEmbed{
		Title:       title,
		Description: getRandomFluffDuelText(yourWin, yourCharName, theirCharName),
		// Convert hex to int here
		// https://www.rapidtables.com/convert/number/hex-to-decimal.html?x=36A2EB
		Color: 3580651,
		Image: &discordgo.MessageEmbedImage{
			URL: "attachment://image.png",
		},
	}

	if thumbnail != "" {
		embeddedData.Thumbnail = &discordgo.MessageEmbedThumbnail{
			URL: thumbnail,
		}
	}

	// charData, err := FetchCharacterData(charName, apiredis.OPTIONAL_CONF_MAPLE_REGION.GetWithDefault(apiredis.RedisDB, os.Getenv("MAPLE_REGION")))
	// if err == nil {
	// 	embeddedData.Title = strings.Trim(charData.CharacterName+" "+date, " ")
	// 	embeddedData.Thumbnail = &discordgo.MessageEmbedThumbnail{URL: charData.CharacterImgURL}
	// 	embeddedData.Description = charName + " is a Level " + strconv.Itoa(charData.Level) + " " + charData.JobName

	// 	// For more embed examples, visit: https://discordjs.guide/popular-topics/embeds.html#using-the-embed-constructor
	// }

	response := &discordgo.InteractionResponseData{
		Embeds: []*discordgo.MessageEmbed{embeddedData},
		Files:  []*discordgo.File{{Name: "image.png", Reader: chartImageBinData}},
	}

	// f, err := os.Open(filename)
	// if err == nil {
	// outcomeFile := &discordgo.File{Name: "outcome.webp", Reader: f, ContentType: "image/webp"}
	// response.Files = append(response.Files, outcomeFile)
	// }

	return response, nil
}
