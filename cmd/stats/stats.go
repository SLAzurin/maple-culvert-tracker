package main

import (
	"log"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"github.com/slazurin/maple-culvert-tracker/internal/db"
)

func main() {
	now := time.Now()
	for now.Weekday() != time.Sunday {
		now = now.Add(time.Hour * -24)
	}
	last12WeeksCulvertRaw := []time.Time{}
	last12WeeksCulvertDate := []string{}
	for i := 0; i < 12; i++ {
		last12WeeksCulvertRaw = append(last12WeeksCulvertRaw, now)
		now = now.Add(time.Hour * -24 * 7)
	}
	for _, v := range last12WeeksCulvertRaw {
		last12WeeksCulvertDate = append(last12WeeksCulvertDate, v.Format("2006-01-02"))
	}

	log.Println(last12WeeksCulvertDate)

	// Fetch active characters
	stmt, _ := db.DB.Prepare(`select character_culvert_scores.character_id as id, maple_character_name
    from character_culvert_scores
	inner join characters on characters.id = character_culvert_scores.character_id
    where culvert_date = $1`)

	rows, _ := stmt.Query(last12WeeksCulvertDate[0])

	chars := []struct {
		Name string
		ID   int64
	}{}

	for rows.Next() {
		v := struct {
			Name string
			ID   int64
		}{}
		rows.Scan(&v.ID, &v.Name)
		chars = append(chars, v)
	}
	rows.Close()
	stmt.Close()

	for _, v := range chars {
		stmt, _ := db.DB.Prepare(`select culvert_date, score from character_culvert_scores where character_id = $1 order by culvert_date desc`)
		rows, _ := stmt.Query(v.ID)

		scores := []struct {
			Date  string
			Score int
		}{}
		for rows.Next() {
			v := struct {
				Date  string
				Score int
			}{}
			rows.Scan(&v.Date, &v.Score)
			scores = append(scores, v)
		}
		log.Println(v.Name, scores)
		rows.Close()
		stmt.Close()
	}

	// sandbaggers for the last 12 weeks

	// sandbaggers this week
	// select maple_character_name from character_culvert_scores inner join characters on characters.id = character_culvert_scores.character_id where culvert_date = '2024-07-21' and score = 0;
}
