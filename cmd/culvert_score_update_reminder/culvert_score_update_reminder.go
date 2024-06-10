package main

import (
	"log"
	"os"
	"time"

	"github.com/bwmarrin/discordgo"
	_ "github.com/joho/godotenv/autoload"
	"github.com/slazurin/maple-culvert-tracker/internal/db"
)

var s *discordgo.Session

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

	if count > 0 {
		return
	}

	log.Println("reminding...")

	s, err = discordgo.New("Bot " + os.Getenv("DISCORD_TOKEN"))
	if err != nil {
		log.Printf("Invalid bot parameters: %v", err)
		return
	}
	sendMsgCh := make(chan struct{})
	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)

		content := "Reminder to input culvert scores " + date + " " + os.Getenv("DISCORD_REMINDER_SUFFIX")
		s.ChannelMessageSend(os.Getenv("DISCORD_REMINDER_CHANNEL_ID"), content)
		sendMsgCh <- struct{}{}
	})
	err = s.Open()
	if err != nil {
		log.Printf("Cannot open the session: %v", err)
		return
	}
	defer s.Close()
	ticker := time.NewTicker(5 * time.Second)
	done := make(chan struct{})

	// Either ticker done or send Message done for return statement
	select {
	case <-done:
		return
	case <-sendMsgCh:
		ticker.Stop()
		return
	case <-ticker.C:
		done <- struct{}{}
	}

}
