package controllers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/slazurin/maple-culvert-tracker/internal/api/helpers"
	"github.com/slazurin/maple-culvert-tracker/internal/apiredis"
	"github.com/slazurin/maple-culvert-tracker/internal/data"
)

type DiscordServerController struct{}

func (d DiscordServerController) RetrieveMembers(c *gin.Context) {
	result := []data.WebGuildMember{}
	val, err := apiredis.DATA_DISCORD_MEMBERS.Get(apiredis.RedisDB)
	if err == redis.Nil {
		c.AbortWithStatusJSON(http.StatusOK, result)
		return
	}
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": "failed to retrieve members",
		})
		return
	}
	json.Unmarshal([]byte(val), &result)
	c.JSON(http.StatusOK, result)
	c.Abort()
}

func (d DiscordServerController) RetrieveMembersForce(c *gin.Context) {
	result, err := helpers.FetchMembers(c.GetString("discord_server_id"), DiscordSession)
	if err != nil {
		log.Println("Failed to fetch members during force", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": "failed to force fetch members",
		})
		return
	}

	resultData, _ := json.Marshal(result)
	apiredis.DATA_DISCORD_MEMBERS.Set(apiredis.RedisDB, string(resultData))

	c.JSON(http.StatusOK, result)
	c.Abort()
}
