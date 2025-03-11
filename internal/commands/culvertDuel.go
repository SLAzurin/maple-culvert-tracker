package commands

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"slices"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/go-jet/jet/v2/postgres"
	"github.com/slazurin/maple-culvert-tracker/.gen/mapleculverttrackerdb/public/model"
	"github.com/slazurin/maple-culvert-tracker/.gen/mapleculverttrackerdb/public/table"
	apihelpers "github.com/slazurin/maple-culvert-tracker/internal/api/helpers"
	"github.com/slazurin/maple-culvert-tracker/internal/apiredis"
	"github.com/slazurin/maple-culvert-tracker/internal/commands/helpers"
	"github.com/slazurin/maple-culvert-tracker/internal/data"
	"github.com/slazurin/maple-culvert-tracker/internal/db"
)

func culvertDuel(anyone bool) func(*discordgo.Session, *discordgo.InteractionCreate) {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		// Parse discord param character-name
		yourChar := ""
		theirChar := ""
		options := i.ApplicationCommandData().Options
		for _, v := range options {
			if v.Name == "your-character" {
				yourChar = strings.Trim(strings.ToLower(v.StringValue()), " ")
			}
			if v.Name == "their-character" {
				theirChar = strings.Trim(strings.ToLower(v.StringValue()), " ")
			}
		}
		yourWin := false

		// sql := `SELECT id, maple_character_name FROM characters WHERE characters.discord_user_id = $1 AND LOWER(characters.maple_character_name) = $2 ORDER BY maple_character_name ASC`
		jetsql := postgres.SELECT(table.Characters.ID, table.Characters.MapleCharacterName).
			FROM(table.Characters).
			WHERE(table.Characters.DiscordUserID.EQ(postgres.String(i.Member.User.ID)).AND(postgres.LOWER(table.Characters.MapleCharacterName).EQ(postgres.String(yourChar))))
		if anyone {
			jetsql = postgres.SELECT(table.Characters.ID, table.Characters.MapleCharacterName).
				FROM(table.Characters).
				WHERE(postgres.LOWER(table.Characters.MapleCharacterName).EQ(postgres.String(yourChar)))

		}
		sqlCharacters := []model.Characters{}

		characters := map[string]struct {
			name string
			id   int64
		}{}

		err := jetsql.Query(db.DB, &sqlCharacters)
		if err != nil {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Database connection failed. Failed to get your available characters. Please report this to a guild jr.",
				},
			})
			return
		}
		for _, v := range sqlCharacters {
			characters[yourChar] = struct {
				name string
				id   int64
			}{name: v.MapleCharacterName, id: v.ID}

		}
		if len(characters) != 1 {
			resp := &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Your character was not found. Your available characters:",
				},
			}
			yourChars, err := helpers.GetCharactersByDiscordID(db.DB, i.Member.User.ID)
			if err == nil {
				resp.Data.Flags = discordgo.MessageFlagsEphemeral

				s := ""
				for _, v := range *yourChars {
					s += v.MapleCharacterName + "\n"
				}
				if s != "" {
					resp.Data.Files = []*discordgo.File{{Name: "message.txt", Reader: strings.NewReader(s)}}
				}
			} else {
				resp.Data.Content = "Your character was not found. Database connection failed."
			}
			s.InteractionRespond(i.Interaction, resp)
			return
		}

		// query theirs
		activeChars, err := helpers.GetAcviveCharacters(apiredis.RedisDB, db.DB)

		// Count # of chars
		if err != nil {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Database connection failed. Failed to get all active characters. Please report this to a guild jr.",
				},
			})
			return
		}
		for _, v := range *activeChars {
			if theirChar == strings.ToLower(v.MapleCharacterName) {
				characters[theirChar] = struct {
					name string
					id   int64
				}{name: v.MapleCharacterName, id: v.ID}
				break
			}
		}
		if len(characters) != 2 {
			st := ""
			for _, v := range *activeChars {
				st += v.MapleCharacterName + "\n"
			}
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Their character was not found. Available characters:",
					Flags:   discordgo.MessageFlagsEphemeral,
					Files:   []*discordgo.File{{Name: "message.txt", Reader: strings.NewReader(st)}},
				},
			})
			return
		}

		// There is only 2 characters
		// query score
		characterData := map[string]map[string]int{} // character -> date -> score
		characterData[yourChar] = map[string]int{}
		characterData[theirChar] = map[string]int{}
		chartData := data.ChartMakeMultiplePoints{
			Labels:    []string{},
			DataPlots: []data.DataPlot{},
		}
		for currentChar, v := range characters {
			sql := `SELECT character_culvert_scores.culvert_date, character_culvert_scores.score FROM characters INNER JOIN character_culvert_scores ON character_culvert_scores.character_id = characters.id WHERE characters.id = $1 ORDER BY character_culvert_scores.culvert_date DESC LIMIT 8`
			// Concat here is not an sql injection because I trust discord sanitizing the `weeks` variable
			stmt, err := db.DB.Prepare(sql)
			if err != nil {
				log.Println("Failed 1st prepare at culvert command", err)
				return
			}
			defer stmt.Close()
			rows, err := stmt.Query(v.id)
			if err != nil {
				log.Println("Query at culvert command", err)
				return
			}
			defer rows.Close()

			for rows.Next() {
				pt := data.ChartMakerPoints{}
				rows.Scan(&pt.Label, &pt.Score)
				if currentChar == yourChar {
					chartData.Labels = append(chartData.Labels, pt.Label[5:10])
				}
				characterData[currentChar][pt.Label[5:10]] = pt.Score
			}

			if currentChar == yourChar && len(chartData.Labels) == 0 {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "No data on your character...",
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				})
				return
			}
		}
		slices.Reverse(chartData.Labels)

		for charNameLower, characterDataScores := range characterData {
			dataPlot := data.DataPlot{
				CharacterName: characters[charNameLower].name,
				Scores:        []int{},
			}
			for _, label := range chartData.Labels {
				if _, ok := characterDataScores[label]; ok {
					dataPlot.Scores = append(dataPlot.Scores, characterDataScores[label])
				} else {
					dataPlot.Scores = append(dataPlot.Scores, 0)
				}
			}
			chartData.DataPlots = append(chartData.DataPlots, dataPlot)
		}

		yourWin = characterData[yourChar][chartData.Labels[len(chartData.Labels)-1]] >= characterData[theirChar][chartData.Labels[len(chartData.Labels)-1]]

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

		r, err := http.Post("http://"+os.Getenv(data.EnvVarChartMakerHost)+"/chartmaker-multiple", "application/json", bytes.NewBuffer(jsonData))
		if err != nil || r.StatusCode != http.StatusOK {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Looks like my `chartmaker` component is broken... ",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
		} else {
			data, f := apihelpers.GenerateDiscordCulvertDuelOutput(r.Body, yourWin, characters[yourChar].name,
				characters[theirChar].name, nil)
			if f != nil {
				defer f.Close()
			}
			if yourWin && characterData[theirChar][chartData.Labels[len(chartData.Labels)-1]] == 0 {
				data.Embeds[0].Thumbnail.URL = "https://raw.githubusercontent.com/SLAzurin/7tv-to-gif-stuff/refs/heads/master/ness-sandbag-cropped.png"
			}
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: data,
			})
		}
	}
}
