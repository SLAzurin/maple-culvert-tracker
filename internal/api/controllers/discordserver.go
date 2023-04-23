package controllers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type DiscordServerController struct{}

func (d DiscordServerController) RetrieveMembers(c *gin.Context) {
	discordUsername, exists := c.Get("discord_username")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Failed to parse claims"})
		c.Abort()
		return
	}
	log.Println(discordUsername)
	// c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprintf("will retrieve members from %v and you are %v", c.Param("serverid"), discordUsername)})
	c.JSON(http.StatusOK, []struct{}{})
	c.Abort()
}
