package commands

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/slazurin/maple-culvert-tracker/internal/api/helpers"
	cmdhelpers "github.com/slazurin/maple-culvert-tracker/internal/commands/helpers"
	"github.com/slazurin/maple-culvert-tracker/internal/data"
	"github.com/slazurin/maple-culvert-tracker/internal/db"
)

func culvertBase(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Parse discord param character-name
	charName := ""
	date := ""
	weeks := int64(8)
	options := i.ApplicationCommandData().Options
	for _, v := range options {
		if v.Name == "character-name" {
			charName = strings.ToLower(v.StringValue())
		}
		if v.Name == "date" {
			date = v.StringValue()
		}
		if v.Name == "weeks" {
			weeks = v.IntValue()
		}
	}

	// Validate date format
	if date != "" {
		d, err := time.Parse("2006-01-02", date) // YYYY-MM-DD
		if err != nil {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Invalid date format, should be YYYY-MM-DD",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
			return
		}
		date = cmdhelpers.GetCulvertResetDate(d).Format("2006-01-02")
	}

	// Command name = culvert
	sql := `SELECT id, maple_character_name FROM characters WHERE characters.discord_user_id = $1 ORDER BY id`
	if i.ApplicationCommandData().Name == "culvert-anyone" {
		sql = `SELECT id, maple_character_name FROM characters WHERE characters.discord_user_id != '1' ORDER BY maple_character_name`
	}

	// Count # of chars
	stmt, err := db.DB.Prepare(sql)
	if err != nil {
		log.Println("Failed prepare find characters", err)
		return
	}
	args := []any{}
	if strings.Contains(sql, "$1") {
		args = append(args, i.Member.User.ID)
	}
	rows, err := stmt.Query(args...)
	if err != nil {
		log.Println("Query at find characters", err)
		return
	}
	count := 0
	characters := map[string]struct {
		name string
		id   int64
	}{}
	choices := ""
	lastSeenCharName := ""
	var lastSeenCharID int64 = 0
	choicesNumOfCharInLine := 0
	for rows.Next() {
		count++
		var c string
		var i int64
		rows.Scan(&i, &c)
		if choicesNumOfCharInLine < 3 {
			choices += c + ","
		} else {
			choices += c + "\n"
		}
		choicesNumOfCharInLine++
		if choicesNumOfCharInLine > 3 {
			choicesNumOfCharInLine = 0
		}
		characters[strings.ToLower(c)] = struct {
			name string
			id   int64
		}{name: c, id: i}
		lastSeenCharID = i
		lastSeenCharName = c
	}
	rows.Close()
	stmt.Close()

	choicesMsg := "Available characters:"

	if _, ok := characters[charName]; count == 0 || (count > 1 && charName == "") || (!ok && charName != "") {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: choicesMsg,
				Files:   []*discordgo.File{{Name: "message.csv", Reader: strings.NewReader(choices)}},
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	} else if ok {
		lastSeenCharID = characters[charName].id
		lastSeenCharName = characters[charName].name
	}
	// There is only 1 character, and at this point charID is correct too.

	additionalWhere := ""
	if date != "" {
		additionalWhere += " AND character_culvert_scores.culvert_date <= $2"
	}
	// query score
	sql = `SELECT character_culvert_scores.culvert_date, character_culvert_scores.score FROM characters INNER JOIN character_culvert_scores ON character_culvert_scores.character_id = characters.id WHERE characters.id = $1` + additionalWhere + ` ORDER BY character_culvert_scores.culvert_date DESC LIMIT ` + strconv.FormatInt(weeks, 10)
	// Concat here is not an sql injection because I trust discord sanitizing the `weeks` variable
	stmt, err = db.DB.Prepare(sql)
	if err != nil {
		log.Println("Failed 1st prepare at culvert command", err)
		return
	}
	defer stmt.Close()
	args = []any{lastSeenCharID}
	if date != "" {
		args = append(args, date)
	}
	rows, err = stmt.Query(args...)
	if err != nil {
		log.Println("Query at culvert command", err)
		return
	}
	defer rows.Close()
	chartData := []data.ChartMakerPoints{}
	for rows.Next() {
		pt := data.ChartMakerPoints{}
		rows.Scan(&pt.Label, &pt.Score)
		pt.Label = pt.Label[5:10]
		chartData = append(chartData, pt)
	}

	if len(chartData) == 0 {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "No data on " + lastSeenCharName + "...",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}
	slices.Reverse(chartData)

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

	statistics, _ := cmdhelpers.GetCharacterStatistics(db.DB, lastSeenCharName, date, chartData)
	// Code below handles statistics nil value
	// Error here does not break execution

	// Sample below
	// jsonData := []byte(`[{"label":"2/26","score":0},{"label":"3/5","score":1233},{"label":"3/12","score":8000},{"label":"3/19","score":8100},{"label":"3/26","score":5600},{"label":"4/2","score":5500},{"label":"4/9","score":25000}]`)
	r, err := http.Post("http://"+os.Getenv(data.EnvVarChartMakerHost)+"/chartmaker", "application/json", bytes.NewBuffer(jsonData))
	if err != nil || r.StatusCode != http.StatusOK {
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
			Data: helpers.GenerateDiscordCulvertOutput(r.Body, lastSeenCharName, date, statistics),
		})
	}
}
