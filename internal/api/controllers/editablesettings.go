package controllers

import (
	"net/http"
	"os"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/slazurin/maple-culvert-tracker/internal/apiredis"
	"github.com/slazurin/maple-culvert-tracker/internal/data"
)

type EditableSettingsController struct{}

type EditableSetting struct {
	HumanReadableDescription *apiredis.HumanReadableDescriptions `json:"human_readable_description"`
	Value                    string                              `json:"value"`
	EditableType             string                              `json:"editable_type"`
	Multiple                 bool                                `json:"multiple"`
	AvailableRoles           []*discordgo.Role                   `json:"available_roles,omitempty"`
	AvailableChannels        []*discordgo.Channel                `json:"available_channels,omitempty"`
	AvailableSelections      []string                            `json:"available_selections,omitempty"`
}

type patchEditableBody struct {
	Value string `json:"value"`
	Key   string `json:"key"`
}

func (EditableSettingsController) PatchEditable(s *discordgo.Session) func(c *gin.Context) {
	return func(c *gin.Context) {
		body := patchEditableBody{}
		if err := c.BindJSON(&body); err != nil {
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}
		if body.Value == "" && !strings.HasPrefix(body.Key, "OPTIONAL_") {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "value cannot be empty",
			})
			return
		}
		if rdbKey, ok := apiredis.KeysMap[body.Key]; !ok {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "invalid key",
			})
			return
		} else {
			rdbValue, err := rdbKey.Get(apiredis.RedisDB)
			if err == redis.Nil {
				c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
					"error": "key not found",
				})
				return
			}
			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error": "failed to get key",
				})
				return
			}
			if rdbValue == body.Value {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
					"error": "value not changed",
				})
				return
			}
			switch rdbKey.EditableType {
			case apiredis.EditableTypeDiscordChannel:
				if rdbKey.Multiple {
					// TODO: long-term add support for multiple channels later as it is not needed atm
					c.JSON(http.StatusNotImplemented, gin.H{
						"error": "not implemented",
					})
					return
				} else {
					dChan, err := s.GuildChannels(os.Getenv(data.EnvVarDiscordGuildID))
					if err != nil {
						c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
							"error": "failed to get channel",
						})
						return
					}
					found := false
					for _, dChan := range dChan {
						if dChan.ID == body.Value {
							found = true
							break
						}
					}
					if !found {
						c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
							"error": "invalid channel",
						})
						return
					}

					err = rdbKey.Set(apiredis.RedisDB, body.Value)
					if err != nil {
						c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
							"error": "failed to set key, internal valkey connection error",
						})
					}
					c.JSON(http.StatusOK, gin.H{})
					return
				}
			case apiredis.EditableTypeDiscordRole:
				if rdbKey.Multiple {
					roleIDs := strings.Split(body.Value, ",")
					allRoles, err := s.GuildRoles(os.Getenv(data.EnvVarDiscordGuildID))
					if err != nil {
						c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
							"error": "failed to get existing guild roles from discord, unable to validate roles",
						})
						return
					}
					for _, roleID := range roleIDs {
						found := false
						for _, role := range allRoles {
							if roleID == role.ID {
								found = true
								break
							}
						}
						if !found {
							c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
								"error": "invalid role",
							})
							return
						}
					}
					err = rdbKey.Set(apiredis.RedisDB, body.Value)
					if err != nil {
						c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
							"error": "failed to set key, internal valkey connection error",
						})
						return
					}

					c.JSON(http.StatusOK, gin.H{})
					return
				} else {
					// TODO: long-term add support for single roles later as it is not needed atm
					c.JSON(http.StatusNotImplemented, gin.H{
						"error": "not implemented",
					})
					return
				}
			case apiredis.EditableTypeSelection:
				if availableSelections, ok := apiredis.EditableSelectionsMap[rdbKey]; !ok {
					c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
						"error": "failed to get available selections",
					})
					return
				} else {
					valid := false
					for _, selection := range availableSelections {
						if selection == body.Value {
							valid = true
							break
						}
					}
					if !valid {
						c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
							"error": "invalid selection",
						})
						return
					}
					err = rdbKey.Set(apiredis.RedisDB, body.Value)
					if err != nil {
						c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
							"error": "failed to set key, internal valkey connection error",
						})
					}
					c.JSON(http.StatusOK, gin.H{})
					return
				}
			case apiredis.EditableTypeString:
				err = rdbKey.Set(apiredis.RedisDB, body.Value)
				if err != nil {
					c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
						"error": "failed to set key, internal valkey connection error",
					})
					return
				}
				c.JSON(http.StatusOK, gin.H{})
				return
			default:
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
					"error": "Uneditable type detected",
				})
			}
		}
	}
}
func (EditableSettingsController) GETEditable(s *discordgo.Session) func(c *gin.Context) {
	return func(c *gin.Context) {
		roles, err := s.GuildRoles(os.Getenv(data.EnvVarDiscordGuildID))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "failed to get guild roles",
			})
			return
		}
		channels, err := s.GuildChannels(os.Getenv(data.EnvVarDiscordGuildID))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "failed to get guild channels",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			apiredis.CONF_DISCORD_ADMIN_CHANNEL_ID.ToString(): EditableSetting{
				HumanReadableDescription: apiredis.GetHumanReadableDescriptions(apiredis.CONF_DISCORD_ADMIN_CHANNEL_ID),
				Value:                    apiredis.CONF_DISCORD_ADMIN_CHANNEL_ID.GetWithDefault(apiredis.RedisDB, ""),
				EditableType:             string(apiredis.CONF_DISCORD_ADMIN_CHANNEL_ID.EditableType),
				AvailableChannels:        channels,
				Multiple:                 apiredis.CONF_DISCORD_ADMIN_CHANNEL_ID.Multiple,
			},
			apiredis.CONF_DISCORD_GUILD_ROLE_IDS.ToString(): EditableSetting{
				HumanReadableDescription: apiredis.GetHumanReadableDescriptions(apiredis.CONF_DISCORD_GUILD_ROLE_IDS),
				Value:                    apiredis.CONF_DISCORD_GUILD_ROLE_IDS.GetWithDefault(apiredis.RedisDB, ""),
				EditableType:             string(apiredis.CONF_DISCORD_GUILD_ROLE_IDS.EditableType),
				AvailableRoles:           roles,
				Multiple:                 apiredis.CONF_DISCORD_GUILD_ROLE_IDS.Multiple,
			},
			apiredis.CONF_DISCORD_MEMBERS_MAIN_CHANNEL_ID.ToString(): EditableSetting{
				HumanReadableDescription: apiredis.GetHumanReadableDescriptions(apiredis.CONF_DISCORD_MEMBERS_MAIN_CHANNEL_ID),
				Value:                    apiredis.CONF_DISCORD_MEMBERS_MAIN_CHANNEL_ID.GetWithDefault(apiredis.RedisDB, ""),
				EditableType:             string(apiredis.CONF_DISCORD_MEMBERS_MAIN_CHANNEL_ID.EditableType),
				AvailableChannels:        channels,
				Multiple:                 apiredis.CONF_DISCORD_MEMBERS_MAIN_CHANNEL_ID.Multiple,
			},
			apiredis.OPTIONAL_CONF_CULVERT_DUEL_THUMBNAIL_URL.ToString(): EditableSetting{
				HumanReadableDescription: apiredis.GetHumanReadableDescriptions(apiredis.OPTIONAL_CONF_CULVERT_DUEL_THUMBNAIL_URL),
				Value:                    apiredis.OPTIONAL_CONF_CULVERT_DUEL_THUMBNAIL_URL.GetWithDefault(apiredis.RedisDB, ""),
				EditableType:             string(apiredis.OPTIONAL_CONF_CULVERT_DUEL_THUMBNAIL_URL.EditableType),
				Multiple:                 apiredis.OPTIONAL_CONF_CULVERT_DUEL_THUMBNAIL_URL.Multiple,
			},
			apiredis.OPTIONAL_CONF_DISCORD_REMINDER_SUFFIX.ToString(): EditableSetting{
				HumanReadableDescription: apiredis.GetHumanReadableDescriptions(apiredis.OPTIONAL_CONF_DISCORD_REMINDER_SUFFIX),
				Value:                    apiredis.OPTIONAL_CONF_DISCORD_REMINDER_SUFFIX.GetWithDefault(apiredis.RedisDB, ""),
				EditableType:             string(apiredis.OPTIONAL_CONF_DISCORD_REMINDER_SUFFIX.EditableType),
				Multiple:                 apiredis.OPTIONAL_CONF_DISCORD_REMINDER_SUFFIX.Multiple,
			},
			apiredis.OPTIONAL_CONF_MAPLE_REGION.ToString(): EditableSetting{
				HumanReadableDescription: apiredis.GetHumanReadableDescriptions(apiredis.OPTIONAL_CONF_MAPLE_REGION),
				Value:                    apiredis.OPTIONAL_CONF_MAPLE_REGION.GetWithDefault(apiredis.RedisDB, ""),
				EditableType:             string(apiredis.OPTIONAL_CONF_MAPLE_REGION.EditableType),
				AvailableSelections:      apiredis.EditableSelectionsMap[apiredis.OPTIONAL_CONF_MAPLE_REGION],
				Multiple:                 apiredis.OPTIONAL_CONF_MAPLE_REGION.Multiple,
			},
		})
	}
}
