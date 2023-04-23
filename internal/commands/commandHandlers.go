package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/golang-jwt/jwt/v5"
	"github.com/slazurin/maple-culvert-tracker/internal/data"
	"github.com/slazurin/maple-culvert-tracker/internal/db"
)

var CommandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
	"ping": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "I am alive. You are <@" + i.Member.User.ID + ">, who joined on " + i.Member.JoinedAt.Format(time.RFC822) + ".",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	},
	"culvert": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		// TODO: Fill this with Database query.
		// TODO: this is stupid but this is just sample for now. Remove this double json-ing later
		sql := `SELECT character_culvert_scores.culvert_date, character_culvert_scores.score FROM discord_servers INNER JOIN guild_characters ON guild_characters.discord_server_id = discord_servers.id INNER JOIN character_culvert_scores ON character_culvert_scores.maple_character_name = guild_characters.maple_character_name WHERE discord_servers.discord_server_native_id = $1 AND guild_characters.discord_user_id = $2 ORDER BY character_culvert_scores.culvert_date LIMIT 52`
		stmt, err := db.DB.Prepare(sql)
		if err != nil {
			log.Println("Failed 1st prepare at culvert command", err)
			return
		}
		rows, err := stmt.Query(i.GuildID, i.Member.User.ID)
		if err != nil {
			log.Println("Query at culvert command", err)
			return
		}
		defer rows.Close()
		chartData := []data.ChartMakerPoints{}
		for rows.Next() {
			pt := data.ChartMakerPoints{}
			rows.Scan(&pt.Label, &pt.Score)
			chartData = append(chartData, pt)
		}

		if len(chartData) == 0 {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "No data on your character...",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
			return
		}

		jsonData, err := json.Marshal(chartData)
		if err != nil {
			log.Println("json at culvert command failed?", err)
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Something and something broko...",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
			return
		}

		// Sample below
		// jsonData := []byte(`[{"label":"2/26","score":0},{"label":"3/5","score":1233},{"label":"3/12","score":8000},{"label":"3/19","score":8100},{"label":"3/26","score":5600},{"label":"4/2","score":5500},{"label":"4/9","score":25000}]`)
		r, err := http.Post("http://"+os.Getenv("CHARTMAKER_HOST")+"/chartmaker", "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Looks like my `chartmaker` component is broken... ",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
		} else {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "// TODO: character name",
					Files:   []*discordgo.File{{Name: i.ID + ".png", Reader: r.Body}},
					// TODO: revisit and see if ephemeral is needed FOR CHARTMAKER NOT PING
					// Flags: discordgo.MessageFlagsEphemeral,
				},
			})
		}
	},
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
				Content: fmt.Sprintf("This is your temporary login: `%v`", tokenString),
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	},
}
