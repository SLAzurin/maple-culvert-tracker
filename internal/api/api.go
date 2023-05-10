package api

import (
	"github.com/bwmarrin/discordgo"
	"github.com/gin-gonic/gin"
	"github.com/slazurin/maple-culvert-tracker/internal/api/controllers"
	"github.com/slazurin/maple-culvert-tracker/internal/api/middlewares"
)

var DiscordSession *discordgo.Session

func NewRouter() *gin.Engine {
	controllers.DiscordSession = DiscordSession
	router := gin.Default()
	apiGroup := router.Group("/api")
	{
		apiGroup.Use(middlewares.AuthMiddleware())
		discordGroup := apiGroup.Group("/discord")
		{
			discordServer := controllers.DiscordServerController{}
			discordServerMembers := discordGroup.Group("/members")
			{
				discordServerMembers.GET("/fetch", discordServer.RetrieveMembers)
				discordServerMembers.GET("/force", discordServer.RetrieveMembersForce)
			}
		}
		mapleGroup := apiGroup.Group("/maple")
		{
			maple := controllers.MapleController{}
			mapleGroup.POST("/link", maple.LinkDiscord)

		}
	}
	return router
}
