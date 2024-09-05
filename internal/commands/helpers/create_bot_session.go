package helpers

import (
	"log"
	"os"

	"github.com/bwmarrin/discordgo"
)

func CreateBotSessionWithCommands(commands []*discordgo.ApplicationCommand, commandHandlers map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate)) (*discordgo.Session, error) {
	s, err := discordgo.New("Bot " + os.Getenv("DISCORD_TOKEN"))
	if err != nil {
		return nil, err
	}
	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok && os.Getenv("DISCORD_GUILD_ID") == i.GuildID {
			log.Printf("Got discord command %v from %v\n", i.ApplicationCommandData().Name, i.Member.User.Username)
			h(s, i)
			log.Printf("Done discord command %v from %v\n", i.ApplicationCommandData().Name, i.Member.User.Username)
		}
	})

	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
		err := UpdateCommands(s, commands)
		if err != nil {
			log.Println("Failed UpdateCommands")
			return
		}
		log.Println("Done UpdateCommands Successfully")
	})
	return s, nil
}
