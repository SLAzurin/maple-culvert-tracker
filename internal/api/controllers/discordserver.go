package controllers

import (
	"log"
	"net/http"
	"os"

	"github.com/bwmarrin/discordgo"
	"github.com/gin-gonic/gin"
	"github.com/slazurin/maple-culvert-tracker/internal/data"
)

type DiscordServerController struct{}

func (d DiscordServerController) RetrieveMembers(c *gin.Context) {
	result := []data.WebGuildMember{}

	if DiscordSession == nil {
		log.Println("Discord dead in RetrieveMembers, no session")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": "Broken discord connection",
		})
		return
	}
	serverid := c.Param("serverid")

	// Get all members
	allMembers := []*discordgo.Member{}
	afterMember := ""
	for len(allMembers) == 0 || afterMember != "" {
		members, err := DiscordSession.GuildMembers(serverid, afterMember, 1000)
		if err != nil {
			log.Println("Failed to get members", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "Broken discord connection when getting guild members",
			})
			return
		}
		allMembers = append(allMembers, members...)
		if len(members) == 1000 {
			afterMember = members[999].User.ID
		} else {
			afterMember = ""
		}
	}

	// Get members that are member
	for _, m := range allMembers {
		for _, r := range m.Roles {
			if r == os.Getenv("DISCORD_GUILD_ROLE_ID") {
				wm := data.WebGuildMember{
					DiscordUsername: m.User.Username,
					DiscordUserID:   m.User.ID,
				}
				if m.Nick != "" {
					wm.DiscordUsername = m.Nick
				}
				result = append(result, wm)
			}
		}
	}

	c.JSON(http.StatusOK, result)
	c.Abort()
}
