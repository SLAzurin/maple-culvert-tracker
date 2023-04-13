package main

import (
	"log"
	"os"

	"github.com/bwmarrin/discordgo"
	_ "github.com/joho/godotenv/autoload"
	"github.com/slazurin/maple-culvert-tracker/internal/commands"
)

var s *discordgo.Session

func init() {
	var err error
	s, err = discordgo.New("Bot " + os.Getenv("DISCORD_TOKEN"))
	if err != nil {
		log.Fatalf("Invalid bot parameters: %v", err)
	}
}

func main() {
	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
	})
	err := s.Open()
	if err != nil {
		log.Fatalf("Cannot open the session: %v", err)
	}
	for _, v := range commands.Commands {
		_, err := s.ApplicationCommandCreate(s.State.User.ID, os.Getenv("DISCORD_GUILD_ID"), v)
		if err != nil {
			log.Panicf("Cannot create '%v' command: %v", v.Name, err)
		}
		log.Println("Added command " + v.Name)
	}
	s.Close()
}
