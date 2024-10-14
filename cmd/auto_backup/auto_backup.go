package main

import (
	"bytes"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	_ "github.com/joho/godotenv/autoload"
	"github.com/slazurin/maple-culvert-tracker/internal/apiredis"
)

var s *discordgo.Session

func startBackup(s *discordgo.Session, stopChan chan struct{}) {
	// cmd := exec.Command("/usr/local/bin/pg_dump", "-h", "db16", "-U", os.Getenv("POSTGRES_USER"), "-d", os.Getenv("POSTGRES_DB"), "-p", "5432") // Only use this line outside of docker container
	cmd := exec.Command("/usr/local/bin/pg_dump", "-h", os.Getenv("CLIENT_POSTGRES_HOST"), "-U", os.Getenv("POSTGRES_USER"), "-d", os.Getenv("POSTGRES_DB"), "-p", os.Getenv("CLIENT_POSTGRES_PORT"))
	cmd.Env = append(cmd.Env, "PGPASSWORD="+os.Getenv("POSTGRES_PASSWORD"))
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		log.Println(stdout.String())
		log.Println(stderr.String())
		panic(err)
	}

	_, err = s.ChannelMessageSendComplex(apiredis.CONF_DISCORD_ADMIN_CHANNEL_ID.GetWithDefault(apiredis.RedisDB, os.Getenv("DISCORD_REMINDER_CHANNEL_ID")), &discordgo.MessageSend{
		Content: "Automatic Database backup " + time.Now().Format("2006-01-02"),
		Files:   []*discordgo.File{{Name: "dump_" + time.Now().Format("2006-01-02") + ".sql", Reader: strings.NewReader(stdout.String())}},
	})
	if err != nil {
		panic(err)
	}
	stopChan <- struct{}{}
}

func main() {
	var err error
	s, err = discordgo.New("Bot " + os.Getenv("DISCORD_TOKEN"))
	if err != nil {
		log.Fatalf("Invalid bot parameters: %v", err)
	}
	stop := make(chan struct{}, 1)
	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
		go startBackup(s, stop)
	})
	err = s.Open()
	if err != nil {
		log.Fatalf("Cannot open the session: %v", err)
	}
	defer s.Close()
	<-stop
}
