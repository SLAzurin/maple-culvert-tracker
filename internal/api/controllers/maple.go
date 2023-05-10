package controllers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/slazurin/maple-culvert-tracker/internal/db"
)

type MapleController struct{}

type linkDiscordBody struct {
	DiscordUserID string `json:"discord_user_id"`
	CharacterName string `json:"character_name"`
	Link          bool   `json:"link"`
}
// func RetrieveCharacters

// func RetrieveCulvertScores

func (m MapleController) LinkDiscord(c *gin.Context) {
	body := linkDiscordBody{}
	if err := c.BindJSON(&body); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	var err error
	if body.Link {
		_, err = db.DB.Exec("INSERT INTO characters (maple_character_name, discord_user_id) VALUES ($1, $2) ON CONFLICT (maple_character_name) DO UPDATE SET discord_user_id = $2", body.CharacterName, body.DiscordUserID)
	} else {
		body.DiscordUserID = "1"
		_, err = db.DB.Exec("UPDATE characters SET discord_user_id = $2 WHERE maple_character_name = $1", body.CharacterName, body.DiscordUserID)
	}
	if err != nil {
		log.Println("DB ERROR", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": "DB failed.",
		})
		return
	}
	c.AbortWithStatusJSON(http.StatusOK, gin.H{})
}
