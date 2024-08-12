package commands

import (
	"fmt"
	"os"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/golang-jwt/jwt/v5"
	"github.com/slazurin/maple-culvert-tracker/internal/data"
)

var CommandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
	"ping": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "I am alive. You are <@" + i.Member.User.ID + ">, who joined on " + i.Member.JoinedAt.Format(time.RFC822) + ".\nThis bot was created by Azuri. (AzurinDayo on Twitch, SLAzurin on Github)",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	},
	"culvert":        culvertBase,
	"culvert-anyone": culvertBase,
	"login": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		displayName := i.Member.Nick
		if i.Member.Nick == "" {
			displayName = i.Member.User.Username
		}
		claims := &data.MCTClaims{
			DiscordUsername: displayName,
			DiscordServerID: i.GuildID,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(4 * time.Hour)),
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, _ := token.SignedString([]byte(os.Getenv("JWT_SECRET")))

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("This is your temporary login (4 hours): `%v`\n\n%v", tokenString, os.Getenv("FRONTEND_URL")),
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		s.ChannelMessageSend(i.ChannelID, "<@"+i.Member.User.ID+"> is logging in. Please try to not double edit and mess something up :)")
	},
	"sandbaggers": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		s.InteractionRespond(i.Interaction, getSandbaggers())
	},
	"culvert-duel":        culvertDuel(false),
	"culvert-duel-anyone": culvertDuel(true),
	"export-csv":          exportcsv,
}
