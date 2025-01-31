package commands

//lint:file-ignore ST1001 Dot imports by jet
import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"slices"
	"time"

	"github.com/bwmarrin/discordgo"
	. "github.com/go-jet/jet/v2/postgres"
	. "github.com/slazurin/maple-culvert-tracker/.gen/mapleculverttrackerdb/public/table"
	cmdhelpers "github.com/slazurin/maple-culvert-tracker/internal/commands/helpers"
	"github.com/slazurin/maple-culvert-tracker/internal/data"
	"github.com/slazurin/maple-culvert-tracker/internal/db"
)

func culvertMegaDetails(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options

	weeks := int64(8)
	date := ""

	for _, v := range options {
		if v.Name == "date" {
			date = v.StringValue()
		}
		if v.Name == "weeks" {
			weeks = v.IntValue()
		}
	}

	if date == "" {
		date = cmdhelpers.GetCulvertResetDate(time.Now()).Format("2006-01-02")
	}

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

	// get all dates necessary
	dates := []time.Time{}
	for i := 0; i < int(weeks); i++ {
		dates = append(dates, cmdhelpers.GetCulvertResetDate(d.AddDate(0, 0, -i*7)))
	}

	allDates := []Expression{}
	dateLabels := []string{}
	for _, v := range dates {
		allDates = append(allDates, DateT(v))
		dateLabels = append(dateLabels, v.Format("2006-01-02"))
	}
	slices.Reverse(dateLabels)

	// get all rows for past x week from weeks value
	stmt := SELECT(Characters.MapleCharacterName.AS("maple_character_name"), CharacterCulvertScores.Score.AS("score"), CharacterCulvertScores.CulvertDate.AS("culvert_date")).FROM(CharacterCulvertScores.INNER_JOIN(Characters, Characters.ID.EQ(CharacterCulvertScores.CharacterID))).WHERE(CharacterCulvertScores.CulvertDate.IN(allDates...).AND(Characters.DiscordUserID.NOT_EQ(String("1")))).ORDER_BY(Characters.MapleCharacterName.ASC(), CharacterCulvertScores.CulvertDate.DESC())

	dest := []struct {
		CulvertDate        time.Time
		Score              int32
		MapleCharacterName string
	}{}
	err = stmt.Query(db.DB, &dest)
	if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Failed to find all characters' dataset!",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	// log dest
	for _, v := range dest {
		log.Println(v)
	}

	if len(dest) < 1 {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "There is no data!",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	chartData := data.ChartMakeMultiplePoints{
		Labels:    []string{},
		DataPlots: []data.DataPlot{},
	}

	chartData.Labels = dateLabels
	currentChar := ""
	rawRowData := []*struct {
		CulvertDate        time.Time
		Score              int32
		MapleCharacterName string
	}{}
	for i, v := range dest {
		if currentChar == "" {
			currentChar = v.MapleCharacterName
		}
		if v.MapleCharacterName == currentChar && i < len(dest)-1 {
			rawRowData = append(rawRowData, &v)
			continue
		}
		// add data to chart plot data
		// convert row data to map of date => score
		scores := map[string]int{}
		for _, v := range dateLabels {
			scores[v] = 0
		}
		for _, v := range rawRowData {
			scores[v.CulvertDate.Format("2006-01-02")] = int(v.Score)
		}
		currentScores := []int{}
		for _, v := range dateLabels {
			currentScores = append(currentScores, scores[v])
		}

		// actually append
		chartData.DataPlots = append(chartData.DataPlots, data.DataPlot{
			CharacterName: currentChar,
			Scores:        currentScores,
		})

		// reset data to next character's data
		currentChar = v.MapleCharacterName
		rawRowData = []*struct {
			CulvertDate        time.Time
			Score              int32
			MapleCharacterName string
		}{
			&v,
		}
	}

	// json format chartData
	jsonData, err := json.Marshal(chartData)
	log.Println(string(jsonData), err)

	r, err := http.Post("http://"+os.Getenv(data.EnvVarChartMakerHost)+"/chartmaker-multiple", "application/json", bytes.NewBuffer(jsonData))
	if err != nil || r.StatusCode != http.StatusOK {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Looks like my `chartmaker` component is broken... ",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	} else {
		defer r.Body.Close()
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags: discordgo.MessageFlagsEphemeral,
				Files: []*discordgo.File{{Name: "image.png", Reader: r.Body}},
			},
		})
		return
	}
}
