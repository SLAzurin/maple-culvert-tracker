package commands

//lint:file-ignore ST1001 Dot imports by jet

import (
	"log"
	"math"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	. "github.com/go-jet/jet/v2/postgres"
	"github.com/jedib0t/go-pretty/v6/table"
	. "github.com/slazurin/maple-culvert-tracker/.gen/mapleculverttrackerdb/public/table"
	"github.com/slazurin/maple-culvert-tracker/internal/apiredis"
	"github.com/slazurin/maple-culvert-tracker/internal/commands/helpers"
	cmdhelpers "github.com/slazurin/maple-culvert-tracker/internal/commands/helpers"
	"github.com/slazurin/maple-culvert-tracker/internal/db"
)

func participation(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Parse discord param character-name
	date := ""
	weeks := int64(4)
	options := i.ApplicationCommandData().Options
	for _, v := range options {
		if v.Name == "date" {
			date = v.StringValue()
		}
		if v.Name == "weeks" {
			weeks = v.IntValue()
		}
	}
	rawDate, err := helpers.GetLatestResetDate(db.DB)
	if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Failed to retrieve latest culvert reset date. See server logs.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}
	// Validate date format
	if date != "" {
		newDate, err := time.Parse(time.DateOnly, date) // YYYY-MM-DD
		if err != nil || newDate.After(rawDate) {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Invalid date format, should be YYYY-MM-DD or past culvert date",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
			return
		}
		rawDate = helpers.GetCulvertResetDate(newDate)
	}

	allDates := []Expression{}
	allDatesRaw := []time.Time{} // newer to older
	for i := 0; i < int(weeks); i++ {
		allDates = append(allDates, DateT(rawDate))
		allDatesRaw = append(allDatesRaw, rawDate)
		rawDate = helpers.GetCulvertPreviousDate(rawDate)
	}

	chars, err := helpers.GetActiveCharacters(apiredis.RedisDB, db.DB)
	if err != nil {
		log.Println("participation cmd get active chars failed", err)
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "participation cmd command failed, could not get active characters",
			},
		})
		return
	}

	charIDsInExpr := []Expression{}
	for _, v := range *chars {
		charIDsInExpr = append(charIDsInExpr, Int64(v.ID))
	}

	/*
		SELECT * data for x weeks and x date
		for each dataset, create map[string]map[string]int representing map[date]map[character_name]score
		create new csv from the map above
	*/

	stmt := SELECT(Characters.MapleCharacterName.AS("name"), CharacterCulvertScores.CulvertDate.AS("culvert_date"), CharacterCulvertScores.Score.AS("score")).FROM(CharacterCulvertScores.INNER_JOIN(Characters, Characters.ID.EQ(CharacterCulvertScores.CharacterID))).WHERE(CharacterCulvertScores.CulvertDate.IN(allDates...).AND(CharacterCulvertScores.CharacterID.IN(charIDsInExpr...))).ORDER_BY(Characters.MapleCharacterName.ASC())

	dest := []struct {
		CulvertDate time.Time
		Score       int64
		Name        string
	}{}
	err = stmt.Query(db.DB, &dest)
	if err != nil {
		log.Println("participation cmd select stmt failed", err)
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "participation cmd failed getting culvert data from db. See server logs.",
			},
		})
		return
	}

	m := map[string]map[string]int64{} // m -> char -> date -> score

	for _, v := range dest {
		d := v.CulvertDate.Format(time.DateOnly)
		if _, ok := m[v.Name]; !ok {
			m[v.Name] = map[string]int64{}
		}
		m[v.Name][d] = v.Score
	}

	slices.Reverse(allDatesRaw)
	participationList := []struct {
		Name       string
		TotalWeeks int
		WeeksCount int
		Percentage string
	}{}
	for charName, scores := range m {
		oldScore := int64(0)
		charData := struct {
			Name       string
			TotalWeeks int
			WeeksCount int
			Percentage string
		}{
			TotalWeeks: len(allDatesRaw),
			WeeksCount: 0,
			Percentage: "",
			Name:       charName,
		}
		for _, v := range allDatesRaw {
			d := v.Format(time.DateOnly)
			score := int64(-1)
			if s, ok := scores[d]; ok {
				score = s
			}
			if score == -1 {
				charData.TotalWeeks--
			}
			if score > helpers.GetSandbagThresholdScore(apiredis.RedisDB, oldScore) {
				charData.WeeksCount++
			}
			if score > oldScore {
				oldScore = score
			}
		}
		charData.Percentage = strconv.Itoa(int(math.Round((float64(charData.WeeksCount)/float64(charData.TotalWeeks))*100))) + "%"
		participationList = append(participationList, charData)
	}

	slices.SortStableFunc(participationList, func(a struct {
		Name       string
		TotalWeeks int
		WeeksCount int
		Percentage string
	}, b struct {
		Name       string
		TotalWeeks int
		WeeksCount int
		Percentage string
	}) int {
		return strings.Compare(a.Name, b.Name)
	})

	columnCount := 1
	if len(participationList) > 65 {
		columnCount = 2
	}

	if len(participationList) > 130 {
		columnCount = 3
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Participation for " + strconv.Itoa(int(weeks)) + " weeks on " + rawDate.Format(time.DateOnly),
			Files: []*discordgo.File{{Name: "message.txt", Reader: strings.NewReader(cmdhelpers.FormatNthColumnList(columnCount, participationList, table.Row{"Name", "Weeks", "%"}, func(data struct {
				Name       string
				TotalWeeks int
				WeeksCount int
				Percentage string
			}, idx int) table.Row {
				return table.Row{data.Name, strconv.Itoa(data.WeeksCount) + "/" + strconv.Itoa(data.TotalWeeks), data.Percentage}
			}))}},
		},
	})
}
