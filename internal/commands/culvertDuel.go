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
	"github.com/slazurin/maple-culvert-tracker/internal/data"
	"github.com/slazurin/maple-culvert-tracker/internal/db"
)

func culvertDuel(s *discordgo.Session, i *discordgo.InteractionCreate) {
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

	sql := `SELECT id, maple_character_name FROM characters WHERE characters.discord_user_id = $1 AND LOWER(characters.maple_character_name) = $2 ORDER BY id`

	// Count # of chars
	stmt, err := db.DB.Prepare(sql)
	if err != nil {
		log.Println("Failed prepare find characters", err)
		return
	}
	defer stmt.Close()
	rows, err := stmt.Query(i.Member.User.ID, yourChar)
	if err != nil {
		log.Println("Query at find characters", err)
		return
	}
	defer rows.Close()
	characters := map[string]struct {
		name string
		id   int64
	}{}
	if rows.Next() {
		var c string
		var i int64
		rows.Scan(&i, &c)
		characters[yourChar] = struct {
			name string
			id   int64
		}{name: c, id: i}
	} else {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Your character not found. Make sure the accents are correct if there are any.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	// query theirs
	//This query can possibly find unlinked characters anyway.
	// I'm pretty sure the frontend filters out the characters linked to discord users who have left the server
	// So in theory, even if I exclude the manually unlinked, there can be orphan character rows
	sql = `SELECT id, maple_character_name FROM characters WHERE LOWER(characters.maple_character_name) = $1 AND id != '0' ORDER BY id`

	// Count # of chars
	stmt2, err := db.DB.Prepare(sql)
	if err != nil {
		log.Println("Failed prepare find characters theirs", err)
		return
	}
	defer stmt2.Close()
	rows2, err := stmt2.Query(theirChar)
	if err != nil {
		log.Println("Query at find characters theirs", err)
		return
	}
	defer rows2.Close()
	if rows2.Next() {
		var c string
		var i int64
		rows2.Scan(&i, &c)
		characters[theirChar] = struct {
			name string
			id   int64
		}{name: c, id: i}
	} else {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Their character not found. Make sure the accents are correct if there are any.",
				Flags:   discordgo.MessageFlagsEphemeral,
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
		sql = `SELECT character_culvert_scores.culvert_date, character_culvert_scores.score FROM characters INNER JOIN character_culvert_scores ON character_culvert_scores.character_id = characters.id WHERE characters.id = $1 ORDER BY character_culvert_scores.culvert_date DESC LIMIT 8`
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

	r, err := http.Post("http://"+os.Getenv("CHARTMAKER_HOST")+"/chartmaker-multiple", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Looks like my `chartmaker` component is broken... ",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	} else {
		content := getRandomFluffDuelText(yourWin, characters[yourChar].name, characters[theirChar].name)
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: content,
				Files:   []*discordgo.File{{Name: i.ID + ".png", Reader: r.Body}},
				// Flags: discordgo.MessageFlagsEphemeral,
			},
		})
	}
}

func getRandomFluffDuelText(yourWin bool, yourChar string, theirChar string) string {
	return "getRandomFluffDuelText UNDER CONSTRUCTION"
}
