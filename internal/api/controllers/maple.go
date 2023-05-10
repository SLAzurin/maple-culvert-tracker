package controllers

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/slazurin/maple-culvert-tracker/internal/db"
)

type MapleController struct{}

type linkDiscordBody struct {
	DiscordUserID string `json:"discord_user_id"`
	CharacterName string `json:"character_name"`
	Link          bool   `json:"link"`
}

func (m MapleController) GETCharacters(c *gin.Context) {

	//return []{character_id: character_name}
}

func (m MapleController) POSTCulvert(c *gin.Context) {
	// hardcode past sunday
	// body []{ character_id, score }
}
func (m MapleController) GETCulvert(c *gin.Context) {
	thisWeek := time.Now()
	sub := int(thisWeek.Weekday())
	thisWeek = thisWeek.Add(time.Hour * -24 * time.Duration(sub))
	thisWeekStr := thisWeek.Format("2006-01-02")
	lastWeek := thisWeek.Add(time.Hour * -24 * 7)
	lastWeekStr := lastWeek.Format("2006-01-02")
	rows, err := db.DB.Query("SELECT character_id, culvert_date, score FROM character_culvert_scores WHERE culvert_date = $1 or culvert_date = $2 ORDER BY score DESC;", thisWeekStr, lastWeekStr)
	if err != nil {
		log.Println("DB ERROR GETCulvert", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": "DB failed.",
		})
		return
	}
	defer rows.Close()
	result := []gin.H{}
	for rows.Next() {
		var charID int64
		var culvertDate string
		var score int
		rows.Scan(&charID, &culvertDate, &score)
		result = append(result, gin.H{
			"character_id": charID,
			"culvert_date": culvertDate,
			"score":        score,
		})
	}
	c.AbortWithStatusJSON(http.StatusOK, result)
	//return []{character_id: { PreviousWeek, CurrentWeek }}
}

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
		log.Println("DB ERROR LinkDiscord", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": "DB failed.",
		})
		return
	}
	c.AbortWithStatusJSON(http.StatusOK, gin.H{})
}
