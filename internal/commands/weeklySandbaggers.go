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
	"github.com/slazurin/maple-culvert-tracker/internal/commands/helpers"
	"github.com/slazurin/maple-culvert-tracker/internal/data"
	"github.com/slazurin/maple-culvert-tracker/internal/db"
)

type statsSecondary = struct {
	Name                 string
	Score                int
	RawStats             *data.CharacterStatistics
	DiffPbPercentage     int
	DiffMedianPercentage int
}

func weeklySandbaggers(s *discordgo.Session, i *discordgo.InteractionCreate) {
	var err error

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})

	options := i.ApplicationCommandData().Options

	stmt := SELECT(MAX(CharacterCulvertScores.CulvertDate).AS("max")).FROM(CharacterCulvertScores)
	rawDateOut := struct {
		Max string
	}{}
	err = stmt.Query(db.DB, &rawDateOut)
	if err != nil {
		content := "Fatal internal error, check server logs"
		log.Println("Failed getting latest date", err)
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		return
	}

	rawDate := rawDateOut.Max[:10]

	date, _ := time.Parse("2006-01-02", rawDate)

	threshold := float64(7) / float64(10) // inverse 30%
	weeks := 12

	for _, v := range options {
		if v.Name == "date" {
			d, err := time.Parse("2006-01-02", v.StringValue()) // YYYY-MM-DD
			if err != nil {
				content := "Invalid date"
				s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
					Content: &content,
				})
				return
			}
			rawDate = d.Format("2006-01-02")
			if d.Weekday() != helpers.GetCulvertResetDay(d) {
				content := "Date " + rawDate + " is not culvert reset day..."
				s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
					Content: &content,
				})
				return
			}
			date = d
		}
		if v.Name == "pb-diff-threshold" {
			percentage := v.IntValue()
			threshold = float64(percentage) / float64(100)
		}
		if v.Name == "weeks" {
			weeks = int(v.IntValue())
		}
	}

	characters := []string{}

	stmt = SELECT(Characters.MapleCharacterName).
		FROM(Characters.INNER_JOIN(CharacterCulvertScores, CharacterCulvertScores.CharacterID.EQ(Characters.ID).AND(CharacterCulvertScores.CulvertDate.EQ(DateT(date))))).
		ORDER_BY(Characters.MapleCharacterName)

	err = stmt.Query(db.DB, &characters)
	if err != nil {
		content := "Failed to fetch all characters"
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		return
	}

	zeroScoreCharacters := []string{}
	bigboyCharacterStatsSecondary := []statsSecondary{}

	// Build all Secondary stats
	for _, v := range characters {
		stmt = SELECT(CharacterCulvertScores.CulvertDate.AS("culvert_date"), CharacterCulvertScores.Score.AS("score")).FROM(Characters.INNER_JOIN(CharacterCulvertScores, CharacterCulvertScores.CharacterID.EQ(Characters.ID))).WHERE(Characters.MapleCharacterName.EQ(String(v))).LIMIT(int64(weeks)).ORDER_BY(CharacterCulvertScores.CulvertDate.DESC())

		scoresRawDb := []struct {
			Score       int
			CulvertDate string
		}{}
		err = stmt.Query(db.DB, &scoresRawDb)
		if err != nil {
			content := "Fatal error, failed to fetch scores for " + v
			s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Content: &content,
			})
			return
		}

		if scoresRawDb[0].Score <= 0 {
			zeroScoreCharacters = append(zeroScoreCharacters, v)
			continue
		}

		chartData := []data.ChartMakerPoints{}
		for _, v := range scoresRawDb {
			d := data.ChartMakerPoints{
				Label:   v.CulvertDate[:10],
				RawDate: v.CulvertDate[:10],
				Score:   v.Score,
			}
			chartData = append(chartData, d)
		}
		slices.Reverse(chartData)

		charaStats, err := helpers.GetCharacterStatistics(db.DB, v, rawDate, chartData)
		if err != nil {
			content := "Fatal error, failed to fetch statistics for " + v
			s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Content: &content,
			})
			return
		}

		latestWeekScore := chartData[len(chartData)-1].Score

		diffPbRatio := float64(latestWeekScore) / float64(charaStats.PersonalBest)
		if diffPbRatio <= threshold {
			secondaryStats := statsSecondary{
				Name:                 v,
				RawStats:             charaStats,
				Score:                latestWeekScore,
				DiffPbPercentage:     diffPercentage(latestWeekScore, charaStats.PersonalBest),
				DiffMedianPercentage: diffPercentage(latestWeekScore, charaStats.Median),
			}
			bigboyCharacterStatsSecondary = append(bigboyCharacterStatsSecondary, secondaryStats)
		}

	}

	slices.SortStableFunc(bigboyCharacterStatsSecondary, func(a statsSecondary, b statsSecondary) int {
		return a.DiffPbPercentage - b.DiffPbPercentage
	})

	columnCount := 1
	if len(bigboyCharacterStatsSecondary) > 65 {
		columnCount = 2
	}

	if len(bigboyCharacterStatsSecondary) > 130 {
		columnCount = 3
	}
	detailsTable := helpers.FormatNthColumnList(columnCount, bigboyCharacterStatsSecondary, table.Row{"", "Score", "Personal Best", "% of", "Median", "% of"}, func(data statsSecondary, idx int) table.Row {
		diffpb := strconv.Itoa(data.DiffPbPercentage) + "%"
		diffMd := strconv.Itoa(data.DiffMedianPercentage) + "%"
		return table.Row{data.Name, data.Score, data.RawStats.PersonalBest, diffpb, data.RawStats.Median, diffMd}
	})

	detailsZeroScoreCharas := strings.Join(zeroScoreCharacters, ",")

	content := "Here are the weekly sandbaggers details for " + rawDate + " over " + strconv.Itoa(weeks) + " weeks"
	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &content,
		Files: []*discordgo.File{
			{
				Name:   "sandbaggers.txt",
				Reader: strings.NewReader(detailsTable),
			},
			{
				Name:   "zero-score-characters.txt",
				Reader: strings.NewReader(detailsZeroScoreCharas),
			},
		},
	})
}

func diffPercentage(v int, against int) int {
	if against == 0 {
		return 0
	}
	ratio := float64(v) / float64(against)
	ratio *= 100
	return int(math.Round(ratio))
}
