package api

import (
	"github.com/gin-gonic/gin"
	"github.com/slazurin/maple-culvert-tracker/internal/api/controllers"
	"github.com/slazurin/maple-culvert-tracker/internal/api/middlewares"
)

func NewRouter() *gin.Engine {
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
