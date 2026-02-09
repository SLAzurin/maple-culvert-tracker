package api

import (
	"net/http"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/gin-gonic/gin"
	"github.com/slazurin/maple-culvert-tracker/internal/api/controllers"
	"github.com/slazurin/maple-culvert-tracker/internal/api/middlewares"
)

var DiscordSession *discordgo.Session
var startHealthcheckFail *time.Time
var startHealthcheckFailMutex = sync.Mutex{}

func NewRouter() *gin.Engine {
	controllers.DiscordSession = DiscordSession
	router := gin.Default()
	router.GET("health", func(c *gin.Context) {
		// no need to check http, this is self validated
		d := time.Now().UTC()
		c.Writer.Header().Set("Date", d.Format(time.RFC1123))
		if DiscordSession == nil || !DiscordSession.DataReady {
			startHealthcheckFailMutex.Lock()
			defer startHealthcheckFailMutex.Unlock()
			if startHealthcheckFail == nil {
				startHealthcheckFail = &d
			}
			status := http.StatusServiceUnavailable
			if d.After(startHealthcheckFail.Add(30 * time.Second)) {
				status = http.StatusInternalServerError
			}
			if status == http.StatusServiceUnavailable {
				c.Writer.Header().Set("Retry-After", d.Add(time.Second*5).Format(time.RFC1123))
			}
			c.AbortWithStatus(status)
			return
		} else if startHealthcheckFail != nil {
			startHealthcheckFail = nil
		}
	})
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
