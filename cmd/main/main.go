package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/bwmarrin/discordgo"
	_ "github.com/joho/godotenv/autoload"
	"github.com/slazurin/maple-culvert-tracker/internal/api"
	"github.com/slazurin/maple-culvert-tracker/internal/commands"
	"github.com/slazurin/maple-culvert-tracker/internal/commands/helpers"
	"github.com/slazurin/maple-culvert-tracker/internal/db"
)

var s *discordgo.Session

func init() {
	var err error
	s, err = discordgo.New("Bot " + os.Getenv("DISCORD_TOKEN"))
	if err != nil {
		log.Fatalf("Invalid bot parameters: %v", err)
	}
	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commands.CommandHandlers[i.ApplicationCommandData().Name]; ok && os.Getenv("DISCORD_GUILD_ID") == i.GuildID {
			log.Printf("Got discord command %v from %v\n", i.ApplicationCommandData().Name, i.Member.User.Username)
			h(s, i)
			log.Printf("Done discord command %v from %v\n", i.ApplicationCommandData().Name, i.Member.User.Username)
		}
	})
}

func main() {
	stop := make(chan os.Signal, 1)
	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
		err := helpers.UpdateCommands(s, commands.Commands)
		if err != nil {
			stop <- os.Interrupt
		}
		log.Println("Done UpdateCommands Successfully")
	})
	err := s.Open()
	if err != nil {
		log.Fatalf("Cannot open the session: %v", err)
	}
	defer s.Close()

	go func() {
		api.DiscordSession = s
		r := api.NewRouter()
		port := os.Getenv("BACKEND_HTTP_PORT")
		if port == "" {
			port = "8080"
		}
		r.Run("0.0.0.0:" + port)
	}()

	go helpers.RunSunToWedFixes(db.DB)

	signal.Notify(stop, os.Interrupt)
	log.Println("Press Ctrl+C to exit")
	<-stop
}
