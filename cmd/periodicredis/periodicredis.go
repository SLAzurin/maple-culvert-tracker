package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/bwmarrin/discordgo"
	_ "github.com/joho/godotenv/autoload"
	"github.com/slazurin/maple-culvert-tracker/internal/api/helpers"
	"github.com/slazurin/maple-culvert-tracker/internal/apiredis"
)

var s *discordgo.Session

func main() {
	log.Println("env", os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT"))

	var err error
	s, err = discordgo.New("Bot " + os.Getenv("DISCORD_TOKEN"))
	if err != nil {
		log.Fatalln("Cannot init discord session", err)
	}
	err = s.Open()
	if err != nil {
		log.Fatalln("Cannot open discord session", err)
	}
	defer s.Close()
	go func() {
		result, err := helpers.FetchMembers(os.Getenv("DISCORD_GUILD_ID"), s)
		if err != nil {
			log.Println("Failed to fetch members periodically")
		} else {
			resultArr, _ := json.Marshal(result)
			cmd := apiredis.RedisDB.Set(context.Background(), "discord_members_"+os.Getenv("DISCORD_GUILD_ID"), string(resultArr), 0)
			log.Println("Set", "discord_members_"+os.Getenv("DISCORD_GUILD_ID"), cmd.Err())
		}
		time.Sleep(time.Minute * 30)
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	log.Println("Press Ctrl+C to exit")
	<-stop
}
