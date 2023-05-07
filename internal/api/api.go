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
			discordServerGroup := discordGroup.Group("/:serverid")
			{
				discordServer := controllers.DiscordServerController{}
				discordServerMembers := discordServerGroup.Group("/members")
				{
					discordServerMembers.GET("/", discordServer.RetrieveMembers)
				}
			}
		}
	}
	return router
}
