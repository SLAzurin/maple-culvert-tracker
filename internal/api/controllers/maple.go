package controllers

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/slazurin/maple-culvert-tracker/internal/db"
)

type MapleController struct{}

type postRenameBody struct {
	NewName     string `json:"new_name"`
	CharacterID int    `json:"character_id"`
}

type linkDiscordBody struct {
	DiscordUserID string `json:"discord_user_id"`
	CharacterName string `json:"character_name"`
	Link          bool   `json:"link"`
}
type postCulvertBody struct {
	IsNew   bool   `json:"isNew"`
	Week    string `json:"week"`
	Payload []struct {
		CharacterID int64 `json:"character_id"`
		Score       int   `json:"score"`
	} `json:"payload"`
}

func (m MapleController) GETCharacters(c *gin.Context) {
	rows, err := db.DB.Query("SELECT id, maple_character_name, discord_user_id FROM characters;")
	if err != nil {
		log.Println("DB ERROR GETCharacters", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": "DB failed.",
		})
		return
	}
	defer rows.Close()
	result := []struct {
		CharacterID   int    `json:"character_id"`
		CharacterName string `json:"character_name"`
		DiscordUserID string `json:"discord_user_id"`
	}{}
	for rows.Next() {
		r := struct {
			CharacterID   int    `json:"character_id"`
			CharacterName string `json:"character_name"`
			DiscordUserID string `json:"discord_user_id"`
		}{}
		rows.Scan(&r.CharacterID, &r.CharacterName, &r.DiscordUserID)
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
	var err error
	if body.Week != "" {
		thisWeek, err = time.Parse("2006-01-02", body.Week)
		if err != nil || thisWeek.Weekday() != 0 {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "Date incorrectly formatted or isn't sunday.",
			})
			return
		}
	}
	thisWeek = thisWeek.Add(time.Hour * -24 * time.Duration(int(thisWeek.Weekday())))
	thisWeekStr := thisWeek.Format("2006-01-02")
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
	thisWeek = thisWeek.Add(time.Hour * -24 * time.Duration(int(thisWeek.Weekday())))
	lastWeek := thisWeek.Add(time.Hour * -24 * 7)
	editableDays := []string{}
	for i := 0; i < 3; i++ {
		editableDays = append(editableDays, thisWeek.Format("2006-01-02"))
		thisWeek = thisWeek.Add(time.Hour * -24 * 7)
	}
	week := c.Query("week")
	if week != "" {
		queryWeek, err := time.Parse("2006-01-02", week)
		if err != nil || queryWeek.Weekday() != 0 {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "Date format error... Probably not your fault. Also date must be Sunday.",
			})
			return
		}
		found := false
		for i := 0; i < len(editableDays); i++ {
			if week == editableDays[i] {
				found = true
				break
			}
		}
		if !found {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "Date not editable.",
			})
			return
		}
		thisWeek = queryWeek
		lastWeek = thisWeek.Add(time.Hour * -24 * 7)
	} else {
		thisWeek, _ = time.Parse("2006-01-02", editableDays[0])
		// lastWeek was not manipulated
	}

	rows, err := db.DB.Query("SELECT character_id, culvert_date, score FROM character_culvert_scores WHERE culvert_date = $1 or culvert_date = $2 ORDER BY score DESC;", thisWeek, lastWeek)
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
	c.AbortWithStatusJSON(http.StatusOK, gin.H{
		"weeks": editableDays,
		"data":  result,
	})
}

func (m MapleController) LinkDiscord(c *gin.Context) {
	body := linkDiscordBody{}
	if err := c.BindJSON(&body); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	var err error
	if body.Link {
		var rows *sql.Rows
		rows, err = db.DB.Query("SELECT discord_user_id, maple_character_name FROM characters WHERE maple_character_name ILIKE $1;", body.CharacterName)
		if err != nil {
			log.Println("DB ERROR LinkDiscord check dupe name", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "DB failed when checking for duplicate names.",
			})
			return
		}
		defer rows.Close()
		var realCharName string
		if rows.Next() {
			var discordid int64
			rows.Scan(&discordid, &realCharName)
			if discordid != 1 {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
					"error": "Found character with duplicate name. Please unlink it first.",
				})
				return
			}
		}
		if realCharName != "" && realCharName != body.CharacterName {
			// body.CharacterName = realCharName
			_, err = db.DB.Exec("UPDATE characters SET discord_user_id = $2, maple_character_name = $1 WHERE maple_character_name = $3", body.CharacterName, body.DiscordUserID, realCharName)
		} else {
			_, err = db.DB.Exec("INSERT INTO characters (maple_character_name, discord_user_id) VALUES ($1, $2) ON CONFLICT (maple_character_name) DO UPDATE SET discord_user_id = $2", body.CharacterName, body.DiscordUserID)
		}
	} else {
		body.DiscordUserID = "1"
		_, err = db.DB.Exec("UPDATE characters SET discord_user_id = $2 WHERE maple_character_name ILIKE $1", body.CharacterName, body.DiscordUserID)
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

func (m MapleController) POSTRename(c *gin.Context) {
	body := postRenameBody{}
	if err := c.BindJSON(&body); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	var err error
	var rows *sql.Rows
	rows, err = db.DB.Query("SELECT maple_character_name FROM characters WHERE maple_character_name ILIKE $1", body.NewName)
	if err != nil || rows.Next() {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": "Duplicate name found...",
		})
		return
	}
	rows.Close()
	_, err = db.DB.Exec("UPDATE characters SET maple_character_name = $1 WHERE id = $2", body.NewName, body.CharacterID)
	if err != nil {
		log.Println("DB ERROR POSTRename", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": "DB failed.",
		})
		return
	}
	c.AbortWithStatusJSON(http.StatusOK, gin.H{})
}
