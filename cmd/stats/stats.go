package main

import (
	"log"
	"time"

	_ "github.com/joho/godotenv/autoload"
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

	// Participation for the past 5 weeks
	db.DB.Prepare()

	// sandbaggers for the last 12 weeks
	
	
	// sandbaggers this week
	// select maple_character_name from character_culvert_scores inner join characters on characters.id = character_culvert_scores.character_id where culvert_date = '2024-07-21' and score = 0;
}
