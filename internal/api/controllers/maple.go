package controllers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/gin-gonic/gin"
	"github.com/slazurin/maple-culvert-tracker/internal/api/helpers"
	"github.com/slazurin/maple-culvert-tracker/internal/apiredis"
	cmdhelpers "github.com/slazurin/maple-culvert-tracker/internal/commands/helpers"
	"github.com/slazurin/maple-culvert-tracker/internal/data"
	"github.com/slazurin/maple-culvert-tracker/internal/db"
)

type MapleController struct{}

type postRenameBody struct {
	NewName         string `json:"new_name"`
	CharacterID     int    `json:"character_id"`
	BypassNameCheck bool   `json:"bypass_name_check"`
}

type linkDiscordBody struct {
	DiscordUserID   string `json:"discord_user_id"`
	CharacterName   string `json:"character_name"`
	Link            bool   `json:"link"`
	BypassNameCheck bool   `json:"bypass_name_check"`
}
type postCulvertBody struct {
	IsNew   bool   `json:"isNew"`
	Week    string `json:"week"`
	Payload []struct {
		CharacterID int64 `json:"character_id"`
		Score       int   `json:"score"`
	} `json:"payload"`
}

func (MapleController) GETCharacters(c *gin.Context) {
	discordIDs, err := apiredis.DATA_DISCORD_MEMBERS.Get(apiredis.RedisDB)
	if err != nil {
		log.Println("Valkey ERROR GETCharacters", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": "Valkey failed.",
		})
		return
	}

	var discordIDsSlice []data.WebGuildMember
	err = json.Unmarshal([]byte(discordIDs), &discordIDsSlice)
	if err != nil {
		log.Println("JSON ERROR GETCharacters", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": "JSON failed.",
		})
		return
	}

	// map discordIDsSlice to map based on discord id as key
	discordIDMap := make(map[string]struct{})
	for _, v := range discordIDsSlice {
		discordIDMap[v.DiscordUserID] = struct{}{}
	}
	discordIDMap["2"] = struct{}{}

	rows, err := db.DB.Query("SELECT id, maple_character_name, discord_user_id FROM characters WHERE discord_user_id != '1';")
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
		if _, ok := discordIDMap[r.DiscordUserID]; ok {
			result = append(result, r)
		}
	}
	c.AbortWithStatusJSON(http.StatusOK, result)
}

func (MapleController) POSTCulvert(c *gin.Context) {
	body := postCulvertBody{}
	if err := c.BindJSON(&body); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	thisWeek := time.Now()
	thisReset := cmdhelpers.GetCulvertResetDate(thisWeek)
	var err error
	if body.Week != "" {
		thisWeek, err = time.Parse("2006-01-02", body.Week)
		if err != nil || thisWeek.Weekday() != cmdhelpers.GetCulvertResetDay(thisWeek) {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "Date incorrectly formatted or isn't a culvert reset day.",
			})
			return
		}
	}
	thisWeek = cmdhelpers.GetCulvertResetDate(thisWeek)
	thisWeekStr := thisWeek.Format("2006-01-02")
	shouldNotifyScoreUpdated := false
	if body.IsNew {
		// Check if we should notify the involved channels if the insert is successful
		if thisReset.Format("2006-01-02") == thisWeek.Format("2006-01-02") {
			shouldNotifyScoreUpdated = true
			rows, err := db.DB.Query("SELECT id FROM character_culvert_scores WHERE culvert_date = $1 AND score > 0 ORDER BY score DESC LIMIT 1", thisReset.Format("2006-01-02"))
			if err != nil {
				log.Println("DB ERROR ShouldNotifyScoreUpdated", err)
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error": "DB failed.",
				})
				return
			}
			if rows.Next() {
				shouldNotifyScoreUpdated = false
			}
			rows.Close()
		}
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

	if shouldNotifyScoreUpdated {
		if DiscordSession != nil {
			serverID := c.GetString("discord_server_id")
			discordUsername := c.GetString("discord_username")
			members, _ := DiscordSession.GuildMembersSearch(serverID, discordUsername, 1)
			// If the user isn't found, error is not fatal.
			var member *discordgo.Member
			if len(members) > 0 {
				member = members[0]
			}

			for _, key := range apiredis.MakeKeysSlice(apiredis.CONF_DISCORD_MEMBERS_MAIN_CHANNEL_ID) {
				if channelID := key.GetWithDefault(apiredis.RedisDB, ""); channelID != "" {
					message := "DID YOU MAKE ANY GAINS SINCE LAST WEEK?"
					if member != nil && member.User != nil && member.User.ID != "" {
						message = fmt.Sprintf("Thank you <@%s> for entering the culvert scores! %s", member.User.ID, message)
					} else {
						message = "The culvert scores have been updated! " + message
					}
					DiscordSession.ChannelMessageSend(channelID, message)
				}
			}
		} else {
			log.Fatalln("DiscordSession is nil when trying to notify score update completed!")
		}
	}

	go helpers.SendWeeklyDifferences(DiscordSession, db.DB, apiredis.RedisDB, thisWeek, apiredis.CONF_DISCORD_ADMIN_CHANNEL_ID.GetWithDefault(apiredis.RedisDB, ""))

	c.AbortWithStatusJSON(http.StatusOK, gin.H{})
}

func (MapleController) GETCulvert(c *gin.Context) {
	thisWeek := time.Now()
	thisWeek = cmdhelpers.GetCulvertResetDate(thisWeek)
	lastWeek := cmdhelpers.GetCulvertPreviousDate(thisWeek)
	editableDays := []string{}
	for i := 0; i < 3; i++ {
		editableDays = append(editableDays, thisWeek.Format("2006-01-02"))
		thisWeek = cmdhelpers.GetCulvertPreviousDate(thisWeek)
	}
	week := c.Query("week")
	if week != "" {
		queryWeek, err := time.Parse("2006-01-02", week)
		if err != nil || queryWeek.Weekday() != cmdhelpers.GetCulvertResetDay(queryWeek) {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "Date format error... Probably not your fault. Also date must be a reset date.",
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
		lastWeek = cmdhelpers.GetCulvertPreviousDate(thisWeek)
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

func (MapleController) LinkDiscord(c *gin.Context) {
	body := linkDiscordBody{}
	if err := c.BindJSON(&body); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	charData, err := helpers.FetchCharacterData(body.CharacterName, apiredis.OPTIONAL_CONF_MAPLE_REGION.GetWithDefault(apiredis.RedisDB, "na"))
	if err != nil && !body.BypassNameCheck {
		c.AbortWithStatusJSON(http.StatusFailedDependency, gin.H{
			"error": err.Error(),
		})
		return
	}

	if charData == nil && !body.BypassNameCheck {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
			"error": "Character not found on official rankings",
		})
		return
	}

	if body.BypassNameCheck {
		charData = &data.PlayerRank{
			CharacterName: body.CharacterName,
		}
	}

	if body.Link {
		var rows *sql.Rows
		rows, err = db.DB.Query("SELECT discord_user_id, maple_character_name FROM characters WHERE maple_character_name = $1;", charData.CharacterName)
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
			if discordid != 1 && body.DiscordUserID != "2" {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
					"error": "Found character with duplicate name. Please unlink it first.",
				})
				return
			}
		}
		if realCharName != "" && realCharName != charData.CharacterName {
			// charData.CharacterName = realCharName
			_, err = db.DB.Exec("UPDATE characters SET discord_user_id = $2, maple_character_name = $1 WHERE maple_character_name = $3", charData.CharacterName, body.DiscordUserID, realCharName)
		} else {
			_, err := db.DB.Exec("INSERT INTO characters (maple_character_name, discord_user_id) VALUES ($1, $2) ON CONFLICT (maple_character_name) DO UPDATE SET discord_user_id = $2", charData.CharacterName, body.DiscordUserID)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
					"error": "Failed to create new character in the database...",
				})
				return
			}
		}
	} else {
		body.DiscordUserID = "1"
		_, err = db.DB.Exec("UPDATE characters SET discord_user_id = $2 WHERE maple_character_name = $1", charData.CharacterName, body.DiscordUserID)
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

func (MapleController) POSTRename(c *gin.Context) {
	body := postRenameBody{}
	if err := c.BindJSON(&body); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	var err error
	var rows *sql.Rows
	rows, err = db.DB.Query("SELECT maple_character_name FROM characters WHERE maple_character_name = $1", body.NewName)
	if err != nil || rows.Next() {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": "Duplicate name found...",
		})
		return
	}
	rows.Close()

	charData, err := helpers.FetchCharacterData(body.NewName, apiredis.OPTIONAL_CONF_MAPLE_REGION.GetWithDefault(apiredis.RedisDB, "na"))
	if err != nil && !body.BypassNameCheck {
		c.AbortWithStatusJSON(http.StatusFailedDependency, gin.H{
			"error": err.Error(),
		})
		return
	}

	if charData == nil && !body.BypassNameCheck {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
			"error": "Character not found on official rankings",
		})
		return
	}

	if body.BypassNameCheck {
		charData = &data.PlayerRank{
			CharacterName: body.NewName,
		}
	}

	_, err = db.DB.Exec("UPDATE characters SET maple_character_name = $1 WHERE id = $2", charData.CharacterName, body.CharacterID)
	if err != nil {
		log.Println("DB ERROR POSTRename", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": "DB failed.",
		})
		return
	}
	c.AbortWithStatusJSON(http.StatusOK, gin.H{})
}
