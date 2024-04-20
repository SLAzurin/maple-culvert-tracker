package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"github.com/slazurin/maple-culvert-tracker/internal/db"
)

func main() {
	// run this at UTC time +1
	now := time.Now()
	for now.Weekday() != time.Sunday {
		now = now.Add(time.Hour * -24)
	}
	date := now.Format("2006-01-02")
	stmt, err := db.DB.Prepare("SELECT COUNT(*) as count FROM character_culvert_scores WHERE culvert_date = $1")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()
	log.Println("Query", date)
	rows, err := stmt.Query(date)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		rows.Scan(&count)
	}

	if count == 0 {
		log.Println("reminding...")
		postBody, _ := json.Marshal(map[string]string{
			"content": "Reminder to input culvert scores " + date + " :meow:",
		})
		responseBody := bytes.NewBuffer(postBody)
		_, err := http.Post(os.Getenv("DISCORD_REMINDER_WEBHOOK"), "application/json", responseBody)
		if err != nil {
			panic(err)
		}
	}
}
