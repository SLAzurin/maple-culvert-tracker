package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/golang-jwt/jwt/v5"
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
		_, err := time.Parse("2006-01-02", date) // YYYY-MM-DD
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
	for rows.Next() {
		count++
		var c string
		var i int64
		rows.Scan(&i, &c)
		choices += c + " "
		characters[strings.ToLower(c)] = struct {
			name string
			id   int64
		}{name: c, id: i}
		lastSeenCharID = i
		lastSeenCharName = c
	}
	rows.Close()
	stmt.Close()

	choicesMsg := "Available characters:" + choices
	if len(choicesMsg) > 2000 {
		choicesMsg = choicesMsg[:1999] + "…"
	}

	if _, ok := characters[charName]; count == 0 || (count > 1 && charName == "") || (!ok && charName != "") {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: choicesMsg,
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
		content := lastSeenCharName
		if date != "" {
			content += " on " + date
		}
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
}
