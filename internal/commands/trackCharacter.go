package commands

import (
	"os"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/slazurin/maple-culvert-tracker/internal/api/helpers"
	"github.com/slazurin/maple-culvert-tracker/internal/apiredis"
	"github.com/slazurin/maple-culvert-tracker/internal/data"
)

func trackCharacter(s *discordgo.Session, i *discordgo.InteractionCreate) {

	// Parse discord param character-name
	discordUserID := ""
	characterName := ""
	skipNameCheck := false
	options := i.ApplicationCommandData().Options
	for _, v := range options {
		if v.Name == "character-name" {
			characterName = strings.ToLower(v.StringValue())
		}
		if v.Name == "discord-user-id" {
			discordUserID = v.StringValue()
		}
		if v.Name == "weeks" {
			skipNameCheck = v.BoolValue()
		}
	}

	//  Validate maple character name
	if !skipNameCheck {
		charData, err := helpers.FetchCharacterData(characterName, apiredis.OPTIONAL_CONF_MAPLE_REGION.GetWithDefault(apiredis.RedisDB, "na"))
		if err != nil {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Data: &discordgo.InteractionResponseData{
					Content: "Failed to find character name in maple  rankings",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
			return
		}
		characterName = charData.CharacterName
	}

	// Validate discord user id
	result, err := helpers.FetchMembers(os.Getenv("DISCORD_SERVER_ID"), s)
	if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Failed to fetch members",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	var discordUser *data.WebGuildMember
	for _, v := range result {
		if v.DiscordGlobalName == discordUserID || v.DiscordUserID == discordUserID {
			discordUser = &v
			break
		}
	}

	if discordUser == nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Failed to find discord user",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	// Link character

	// Check for existing character

	// INSERT or UPDATE character with discord user id

}
