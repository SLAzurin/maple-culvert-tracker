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
			mapleCharacters := mapleGroup.Group("/characters")
			{
				mapleCharacters.GET("/culvert", maple.GETCulvert)
				mapleCharacters.POST("/culvert", maple.POSTCulvert)
				mapleCharacters.GET("/fetch", maple.GETCharacters)
				mapleCharacters.POST("/rename", maple.POSTRename)
			}

		}
		settingsGroup := apiGroup.Group("/editable-settings")
		{
			settings := controllers.EditableSettingsController{}
			settingsGroup.GET("", settings.GETEditable(DiscordSession))
			settingsGroup.PATCH("", settings.PatchEditable(DiscordSession))
		}
	}
	return router
}
