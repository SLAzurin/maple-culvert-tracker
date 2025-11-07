package commands

import (
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/slazurin/maple-culvert-tracker/internal/apiredis"
	"github.com/slazurin/maple-culvert-tracker/internal/commands/helpers"
	"github.com/slazurin/maple-culvert-tracker/internal/data"
	"github.com/slazurin/maple-culvert-tracker/internal/db"
)

// I was overcaffeinated when writing this

func sniffOutRatsScoreIsSandbag(typeshit string, pbScore int64, score int64) bool {
	if typeshit == "zero" {
		return score == int64(0)
	}
	return score < helpers.GetSandbagThreshold(pbScore)
}

func sniffOutRats(s *discordgo.Session, i *discordgo.InteractionCreate) {
	var err error

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})

	options := i.ApplicationCommandData().Options

	weeks := 8
	threshold := float64(3) / float64(10)
	typeshit := "zero" // or threshold

	for _, v := range options {
		if v.Name == "weeks" {
			weeks = int(v.IntValue())
		}
		if v.Name == "weeks-percentage-threshold" {
			percentage := v.IntValue()
			threshold = float64(percentage) / float64(100)
		}
		if v.Name == "value-as-offense" {
			typeshit = v.StringValue()
		}
	}

	date := helpers.GetCulvertResetDate(time.Now()).Format("2006-01-02")

	characters, err := helpers.GetActiveCharacters(apiredis.RedisDB, db.DB)
	if err != nil {
		content := "Internal error: Failed GetActiveCharacters, see server logs"
		log.Println(err)
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		return
	}

	foundYouYoureFucked := []struct {
		SixSeven string
		LWeeks   int
		WWeeks   int
	}{}

	for _, char := range *characters {
		charID := char.ID
		charName := char.MapleCharacterName

		additionalWhere := ""
		if date != "" {
			additionalWhere += " AND character_culvert_scores.culvert_date <= $2"
		}
		// query score
		sql := `SELECT culvert_date, score from (SELECT character_culvert_scores.culvert_date as culvert_date, character_culvert_scores.score as score FROM characters INNER JOIN character_culvert_scores ON character_culvert_scores.character_id = characters.id WHERE characters.id = $1` + additionalWhere + ` ORDER BY character_culvert_scores.culvert_date DESC LIMIT ` + strconv.Itoa(weeks) + ") ORDER BY culvert_date ASC"
		// Concat here is not an sql injection because I trust discord sanitizing the `weeks` variable
		stmt, err := db.DB.Prepare(sql)
		if err != nil {
			log.Println("Failed 1st prepare at culvert command", err)
			content := "Failed 1st prepare sniffOutRats"
			s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Content: &content,
			})
			return
		}
		defer stmt.Close()
		args := []any{charID}
		if date != "" {
			args = append(args, date)
		}
		rows, err := stmt.Query(args...)
		if err != nil {
			log.Println("Query at culvert command", err)
			content := "Failed query character data sniffOutRats"
			s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Content: &content,
			})
			return
		}
		defer rows.Close()

		sandbaggedScores := 0
		rollerCoasterUphillCount := 0
		scoreAtFloor := false
		var week1IsBefore2mPatch *bool
		personalWeeks := 0
		lastKnownGoodScore := int64(-1)

		for rows.Next() {
			personalWeeks++
			// ascending dates
			rawDate := ""
			score := int64(-1)
			rows.Scan(&rawDate, &score)
			rawDate = rawDate[:10]

			if lastKnownGoodScore <= 10 {
				lastKnownGoodScore = score
				if sniffOutRatsScoreIsSandbag(typeshit, lastKnownGoodScore, score) {
					scoreAtFloor = true
				}
			}

			if week1IsBefore2mPatch == nil {
				culvertDate, _ := time.Parse("2006-01-02", rawDate)
				b := culvertDate.Before(data.Date2mPatch) || culvertDate.Equal(data.Date2mPatch)
				week1IsBefore2mPatch = &b
			}

			if *week1IsBefore2mPatch {
				culvertDate, _ := time.Parse("2006-01-02", rawDate)
				if culvertDate.After(data.Date2mPatch) || culvertDate.Equal(data.Date2mPatch) {
					*week1IsBefore2mPatch = false // This ensures we fallback into the else block for the rest of the chartData, no need to re-parse the culvertDate again
					lastKnownGoodScore = int64(score)
				}
			}

			if sniffOutRatsScoreIsSandbag(typeshit, lastKnownGoodScore, score) {
				sandbaggedScores++
				if !scoreAtFloor {
					scoreAtFloor = true
				}
				continue
			}

			// not sandbag here onwards
			if scoreAtFloor {
				rollerCoasterUphillCount++
				scoreAtFloor = false
			}
			if score > lastKnownGoodScore {
				lastKnownGoodScore = score
			}
		}
		if personalWeeks == 0 {
			// next char
			continue
		}

		if float64(rollerCoasterUphillCount)/float64(personalWeeks) >= threshold {
			foundYouYoureFucked = append(foundYouYoureFucked, struct {
				SixSeven string
				LWeeks   int
				WWeeks   int
			}{
				SixSeven: charName,
				LWeeks:   rollerCoasterUphillCount,
				WWeeks:   personalWeeks,
			})
		}
	}

	// Done all chars
	if len(foundYouYoureFucked) > 0 {
		content := "Here is the list of rats"
		contentInner := ""
		for _, v := range foundYouYoureFucked {
			contentInner += v.SixSeven + " has a high amount roller coaster pattern count! " + strconv.Itoa(v.LWeeks) + "/" + strconv.Itoa(v.WWeeks) + "\n"
		}
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &content,
			Files: []*discordgo.File{
				{
					Name:   "message.txt",
					Reader: strings.NewReader(contentInner),
				},
			},
		})
		return
	}
	content := "Squeaky clean! No rats!"
	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &content,
	})
}
