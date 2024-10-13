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
	EditableType             string                              `json:"editable_type"`
	Multiple                 bool                                `json:"multiple"`
	AvailableRoles           []*discordgo.Role                   `json:"available_roles,omitempty"`
	AvailableChannels        []*discordgo.Channel                `json:"available_channels,omitempty"`
	AvailableSelections      []string                            `json:"available_selections,omitempty"`
}

func (SettingsController) PatchEditable(s *discordgo.Session) func(c *gin.Context) {
	return func(c *gin.Context) {

	}
}
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
				EditableType:             string(apiredis.CONF_DISCORD_ADMIN_CHANNEL_ID.EditableType),
				AvailableChannels:        channels,
				Multiple:                 false,
			},
			apiredis.CONF_DISCORD_GUILD_ROLE_IDS.ToString(): EditableSetting{
				HumanReadableDescription: apiredis.GetHumanReadableDescriptions(apiredis.CONF_DISCORD_GUILD_ROLE_IDS),
				Value:                    apiredis.CONF_DISCORD_GUILD_ROLE_IDS.GetWithDefault(apiredis.RedisDB, ""),
				Key:                      apiredis.CONF_DISCORD_GUILD_ROLE_IDS.ToString(),
				EditableType:             string(apiredis.CONF_DISCORD_GUILD_ROLE_IDS.EditableType),
				AvailableRoles:           roles,
				Multiple:                 true,
			},
			apiredis.CONF_DISCORD_MEMBERS_MAIN_CHANNEL_ID.ToString(): EditableSetting{
				HumanReadableDescription: apiredis.GetHumanReadableDescriptions(apiredis.CONF_DISCORD_MEMBERS_MAIN_CHANNEL_ID),
				Value:                    apiredis.CONF_DISCORD_MEMBERS_MAIN_CHANNEL_ID.GetWithDefault(apiredis.RedisDB, ""),
				Key:                      apiredis.CONF_DISCORD_MEMBERS_MAIN_CHANNEL_ID.ToString(),
				EditableType:             string(apiredis.CONF_DISCORD_MEMBERS_MAIN_CHANNEL_ID.EditableType),
				AvailableChannels:        channels,
			},
			apiredis.OPTIONAL_CONF_CULVERT_DUEL_THUMBNAIL_URL.ToString(): EditableSetting{
				HumanReadableDescription: apiredis.GetHumanReadableDescriptions(apiredis.OPTIONAL_CONF_CULVERT_DUEL_THUMBNAIL_URL),
				Value:                    apiredis.OPTIONAL_CONF_CULVERT_DUEL_THUMBNAIL_URL.GetWithDefault(apiredis.RedisDB, ""),
				Key:                      apiredis.OPTIONAL_CONF_CULVERT_DUEL_THUMBNAIL_URL.ToString(),
				EditableType:             string(apiredis.OPTIONAL_CONF_CULVERT_DUEL_THUMBNAIL_URL.EditableType),
			},
			apiredis.OPTIONAL_CONF_DISCORD_REMINDER_SUFFIX.ToString(): EditableSetting{
				HumanReadableDescription: apiredis.GetHumanReadableDescriptions(apiredis.OPTIONAL_CONF_DISCORD_REMINDER_SUFFIX),
				Value:                    apiredis.OPTIONAL_CONF_DISCORD_REMINDER_SUFFIX.GetWithDefault(apiredis.RedisDB, ""),
				Key:                      apiredis.OPTIONAL_CONF_DISCORD_REMINDER_SUFFIX.ToString(),
				EditableType:             string(apiredis.OPTIONAL_CONF_DISCORD_REMINDER_SUFFIX.EditableType),
			},
			apiredis.OPTIONAL_CONF_MAPLE_REGION.ToString(): EditableSetting{
				HumanReadableDescription: apiredis.GetHumanReadableDescriptions(apiredis.OPTIONAL_CONF_MAPLE_REGION),
				Value:                    apiredis.OPTIONAL_CONF_MAPLE_REGION.GetWithDefault(apiredis.RedisDB, ""),
				Key:                      apiredis.OPTIONAL_CONF_MAPLE_REGION.ToString(),
				EditableType:             string(apiredis.OPTIONAL_CONF_MAPLE_REGION.EditableType),
				AvailableSelections:      []string{"", "na", "eu"},
			},
		})
	}
}
