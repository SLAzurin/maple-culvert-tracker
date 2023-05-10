package controllers

import (
	"context"
	"fmt"
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
type postCulvertBody struct {
	IsNew   bool `json:"isNew"`
	Payload []struct {
		CharacterID int64 `json:"character_id"`
		Score       int   `json:"score"`
	} `json:"payload"`
}

func (m MapleController) GETCharacters(c *gin.Context) {
	rows, err := db.DB.Query("SELECT id, maple_character_name FROM characters;")
	if err != nil {
		log.Println("DB ERROR GETCharacters", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": "DB failed.",
		})
		return
	}
	result := []struct {
		CharacterID   int    `json:"character_id"`
		CharacterName string `json:"character_name"`
	}{}
	for rows.Next() {
		r := struct {
			CharacterID   int    `json:"character_id"`
			CharacterName string `json:"character_name"`
		}{}
		rows.Scan(&r.CharacterID, &r.CharacterName)
		result = append(result, r)
	}
	c.AbortWithStatusJSON(http.StatusOK, result)
}

func (m MapleController) POSTCulvert(c *gin.Context) {
	body := postCulvertBody{}
	if err := c.BindJSON(&body); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	thisWeek := time.Now()
	sub := int(thisWeek.Weekday())
	thisWeek = thisWeek.Add(time.Hour * -24 * time.Duration(sub))
	thisWeekStr := thisWeek.Format("2006-01-02")
	var err error
	if body.IsNew {
		query := ""
		args := []interface{}{}
		d := 1
		for _, v := range body.Payload {
			query += fmt.Sprintf("($%d,'%s',$%d),", d, thisWeekStr, d+1)
			d += 2
			args = append(args, v.CharacterID, v.Score)
		}
		query = "INSERT INTO character_culvert_scores (character_id, culvert_date, score) VALUES " + query[:len(query)-1]
		_, err = db.DB.Exec(query, args...)
	} else { //typically this should be a patch request
		tx, errtx := db.DB.BeginTx(context.Background(), nil)
		if errtx != nil {
			log.Println("DB ERROR tx POSTCulvert", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "DB failed.",
			})
			return
		}
		for _, v := range body.Payload {
			_, err = tx.Exec("UPDATE character_culvert_scores SET score = $1 WHERE character_id = $2 AND culvert_date = $3", v.Score, v.CharacterID, thisWeekStr)
			if err != nil {
				log.Println("DB ERROR POSTCulvert", err)
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error": "DB failed.",
				})
				tx.Rollback()
				return
			}
		}
		err = tx.Commit()
	}
	if err != nil {
		log.Println("DB ERROR POSTCulvert", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": "DB failed.",
		})
		return
	}
	c.AbortWithStatusJSON(http.StatusOK, gin.H{})
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
		var culvertDate time.Time
		var score int
		rows.Scan(&charID, &culvertDate, &score)
		result = append(result, gin.H{
			"character_id": charID,
			"culvert_date": culvertDate.Format("2006-01-02"),
			"score":        score,
		})
	}
	c.AbortWithStatusJSON(http.StatusOK, result)
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
