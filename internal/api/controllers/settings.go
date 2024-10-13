package controllers

import (
	"net/http"
	"os"

	"github.com/bwmarrin/discordgo"
	"github.com/gin-gonic/gin"
	"github.com/slazurin/maple-culvert-tracker/internal/apiredis"
)

type SettingsController struct{}

type EditableSetting struct {
	HumanReadableDescription *apiredis.HumanReadableDescriptions `json:"human_readable_description"`
	Value                    string                              `json:"value"`
	Key                      string                              `json:"key"`
	EditableType             editableType                        `json:"editable_type"`
	Multiple                 bool                                `json:"multiple"`
	AvailableRoles           []*discordgo.Role                   `json:"available_roles,omitempty"`
	AvailableChannels        []*discordgo.Channel                `json:"available_channels,omitempty"`
	AvailableSelections      []string                            `json:"available_selections,omitempty"`
}

type editableType string

const (
	editableTypeString         editableType = "string"
	editableTypeUInt           editableType = "uint"
	editableTypeDiscordRole    editableType = "discord_role"
	editableTypeDiscordChannel editableType = "discord_channel"
	editableTypeSelection      editableType = "selection"
)

func (SettingsController) GETEditable(s *discordgo.Session) func(c *gin.Context) {
	return func(c *gin.Context) {
		roles, err := s.GuildRoles(os.Getenv("DISCORD_GUILD_ID"))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "failed to get guild roles",
			})
			return
		}
		channels, err := s.GuildChannels(os.Getenv("DISCORD_GUILD_ID"))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "failed to get guild channels",
			})
			return
		}
		// TODO: filter channels by type 0 (text)
		c.JSON(http.StatusOK, gin.H{
			apiredis.CONF_DISCORD_ADMIN_CHANNEL_ID.ToString(): EditableSetting{
				HumanReadableDescription: apiredis.GetHumanReadableDescriptions(apiredis.CONF_DISCORD_ADMIN_CHANNEL_ID),
				Value:                    apiredis.CONF_DISCORD_ADMIN_CHANNEL_ID.GetWithDefault(apiredis.RedisDB, ""),
				Key:                      apiredis.CONF_DISCORD_ADMIN_CHANNEL_ID.ToString(),
				EditableType:             editableTypeDiscordChannel,
				AvailableChannels:        channels,
				Multiple:                 false,
			},
			apiredis.CONF_DISCORD_GUILD_ROLE_IDS.ToString(): EditableSetting{
				HumanReadableDescription: apiredis.GetHumanReadableDescriptions(apiredis.CONF_DISCORD_GUILD_ROLE_IDS),
				Value:                    apiredis.CONF_DISCORD_GUILD_ROLE_IDS.GetWithDefault(apiredis.RedisDB, ""),
				Key:                      apiredis.CONF_DISCORD_GUILD_ROLE_IDS.ToString(),
				EditableType:             editableTypeDiscordRole,
				AvailableRoles:           roles,
				Multiple:                 true,
			},
			apiredis.CONF_DISCORD_MEMBERS_MAIN_CHANNEL_ID.ToString(): EditableSetting{
				HumanReadableDescription: apiredis.GetHumanReadableDescriptions(apiredis.CONF_DISCORD_MEMBERS_MAIN_CHANNEL_ID),
				Value:                    apiredis.CONF_DISCORD_MEMBERS_MAIN_CHANNEL_ID.GetWithDefault(apiredis.RedisDB, ""),
				Key:                      apiredis.CONF_DISCORD_MEMBERS_MAIN_CHANNEL_ID.ToString(),
				EditableType:             editableTypeDiscordChannel,
				AvailableChannels:        channels,
			},
			apiredis.OPTIONAL_CONF_CULVERT_DUEL_THUMBNAIL_URL.ToString(): EditableSetting{
				HumanReadableDescription: apiredis.GetHumanReadableDescriptions(apiredis.OPTIONAL_CONF_CULVERT_DUEL_THUMBNAIL_URL),
				Value:                    apiredis.OPTIONAL_CONF_CULVERT_DUEL_THUMBNAIL_URL.GetWithDefault(apiredis.RedisDB, ""),
				Key:                      apiredis.OPTIONAL_CONF_CULVERT_DUEL_THUMBNAIL_URL.ToString(),
				EditableType:             editableTypeString,
			},
			apiredis.OPTIONAL_CONF_DISCORD_REMINDER_SUFFIX.ToString(): EditableSetting{
				HumanReadableDescription: apiredis.GetHumanReadableDescriptions(apiredis.OPTIONAL_CONF_DISCORD_REMINDER_SUFFIX),
				Value:                    apiredis.OPTIONAL_CONF_DISCORD_REMINDER_SUFFIX.GetWithDefault(apiredis.RedisDB, ""),
				Key:                      apiredis.OPTIONAL_CONF_DISCORD_REMINDER_SUFFIX.ToString(),
				EditableType:             editableTypeString,
			},
			apiredis.OPTIONAL_CONF_MAPLE_REGION.ToString(): EditableSetting{
				HumanReadableDescription: apiredis.GetHumanReadableDescriptions(apiredis.OPTIONAL_CONF_MAPLE_REGION),
				Value:                    apiredis.OPTIONAL_CONF_MAPLE_REGION.GetWithDefault(apiredis.RedisDB, ""),
				Key:                      apiredis.OPTIONAL_CONF_MAPLE_REGION.ToString(),
				EditableType:             editableTypeSelection,
				AvailableSelections:      []string{"na", "eu"},
			},
		})
	}
}
