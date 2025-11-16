package commands

//lint:file-ignore ST1001 Dot imports by jet
import (
	"log"
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

func weeklySandbaggers(s *discordgo.Session, i *discordgo.InteractionCreate) {
	var err error

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})

	options := i.ApplicationCommandData().Options

	stmt := SELECT(MAX(CharacterCulvertScores.CulvertDate).AS("max")).FROM(CharacterCulvertScores)
	rawDateOut := struct {
		Max time.Time
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

	rawDate := rawDateOut.Max.Format("2006-01-02")

	date := rawDateOut.Max

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

	// Build sandbaggers stats
	sandbaggers, err := helpers.GetWeeklySandbaggers(characters, rawDate, weeks, threshold)
	if err != nil {
		log.Println("weeklySandbaggers.go:GetWeeklySandbaggers:", err)
		content := "Failed to get weekly sandbaggers data, see server logs..."
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		return
	}

	detailsTable := helpers.FormatNthColumnList(1, sandbaggers.NewSandbaggers, table.Row{"", "Score", "Personal Best", "% of", "Median", "% of"}, func(data data.WeeklySandbaggersStats, idx int) table.Row {
		diffpb := strconv.Itoa(data.DiffPbPercentage) + "%"
		diffMd := strconv.Itoa(data.DiffMedianPercentage) + "%"
		return table.Row{data.Name, data.Score, data.RawStats.PersonalBest, diffpb, data.RawStats.Median, diffMd}
	})

	detailsZeroScoreCharas := strings.Join(sandbaggers.ZeroScoreSandbaggers, ",")

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
