package helpers

//lint:file-ignore ST1001 Dot imports by jet
import (
	"database/sql"
	"encoding/json"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	. "github.com/go-jet/jet/v2/postgres"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/slazurin/maple-culvert-tracker/.gen/mapleculverttrackerdb/public/model"
	. "github.com/slazurin/maple-culvert-tracker/.gen/mapleculverttrackerdb/public/table"
	"github.com/slazurin/maple-culvert-tracker/internal/apiredis"
	"github.com/slazurin/maple-culvert-tracker/internal/commands/helpers"
	cmdhelpers "github.com/slazurin/maple-culvert-tracker/internal/commands/helpers"
	"github.com/slazurin/maple-culvert-tracker/internal/data"
	redis "github.com/valkey-io/valkey-go"
)

type weeklyDiffScores struct {
	Name    string
	Prev    int
	Current int
	Oldpos  int
}

func SendWeeklyDifferences(s *discordgo.Session, db *sql.DB, rdb *redis.Client, submittedDate time.Time, channelID ...string) {
	submittedDate = cmdhelpers.GetCulvertResetDate(submittedDate)
	lastWeek := cmdhelpers.GetCulvertResetDate(submittedDate.Add(time.Hour * -24 * 7))

	// query all summary of current character scores
	stmt := SELECT(Characters.ID.AS("id"), Characters.MapleCharacterName.AS("name"), CharacterCulvertScores.Score.AS("score"), CharacterCulvertScores.CulvertDate.AS("culvert_date")).FROM(CharacterCulvertScores.LEFT_JOIN(Characters, Characters.ID.EQ(CharacterCulvertScores.CharacterID))).WHERE(CharacterCulvertScores.CulvertDate.IN(DateT(lastWeek), DateT(submittedDate))).ORDER_BY(CharacterCulvertScores.CulvertDate.DESC(), CharacterCulvertScores.Score.DESC(), Characters.MapleCharacterName.ASC())

	rawData := []struct {
		ID          int64
		Name        string
		Score       int
		CulvertDate time.Time
	}{}

	err := stmt.Query(db, &rawData)
	if err != nil {
		log.Println("DB ERROR SendWeeklyDifferences", err)
		return
	}

	nameToIdxMap := map[string]int{}
	diffs := []weeklyDiffScores{}
	noLongerExistsFromLastWeek := []string{}
	cutoffPos := -1
	characters := []string{}
	charactersModel := []model.Characters{}

	for curPos, v := range rawData {
		if submittedDate.Format("2006-01-02") == v.CulvertDate.Format("2006-01-02") {
			characters = append(characters, v.Name)
			charactersModel = append(charactersModel, model.Characters{MapleCharacterName: v.Name, ID: v.ID})
		}
		if _, ok := nameToIdxMap[v.Name]; !ok && cutoffPos == -1 {
			nameToIdxMap[v.Name] = curPos
		}
		if v.CulvertDate.Format("2006-01-02") == lastWeek.Format("2006-01-02") && cutoffPos == -1 {
			cutoffPos = curPos
		}
		if cutoffPos != -1 {
			if _, ok := nameToIdxMap[v.Name]; ok {
				diffs[nameToIdxMap[v.Name]].Oldpos = curPos + 1 - cutoffPos
				diffs[nameToIdxMap[v.Name]].Prev = v.Score
			} else {
				noLongerExistsFromLastWeek = append(noLongerExistsFromLastWeek, v.Name)
			}
		} else {
			diffs = append(diffs, weeklyDiffScores{
				Name:    v.Name,
				Prev:    -1,
				Current: v.Score,
				Oldpos:  -1,
			})
		}
	}

	// Send nice ascii table to channel
	columnCount := 1
	if len(diffs) > 65 {
		columnCount = 2
	}

	if len(diffs) > 130 {
		columnCount = 3
	}

	rawStr := cmdhelpers.FormatNthColumnList(columnCount, diffs, table.Row{"Character", "Score", "Position"}, func(data weeklyDiffScores, idx int) table.Row {
		return table.Row{data.Name, strconv.Itoa(data.Prev) + " -> " + strconv.Itoa(data.Current), strconv.Itoa(data.Oldpos) + " -> " + strconv.Itoa(idx+1)}
	})

	// display weekly sandbaggers
	shouldShowWeeklySandbaggers := false
	var sandbaggersTable *string
	var sandbaggersZeroScoreList *string
	const medianWeeks = 12

	// get optional conf from valkey
	showSandbaggersRaw := apiredis.OPTIONAL_CONF_SUBMIT_SCORES_SHOW_SANDBAGGERS.GetWithDefault(apiredis.RedisDB, "false")
	json.Unmarshal([]byte(showSandbaggersRaw), &shouldShowWeeklySandbaggers)
	// if it is not "true", always treat false, so ignore error

	if shouldShowWeeklySandbaggers {
		sandbaggers, err := cmdhelpers.GetWeeklySandbaggers(characters, submittedDate.Format("2006-01-02"), medianWeeks, cmdhelpers.SandbagThreshold)
		if err != nil {
			log.Println("send_weekly_differences:GetWeeklySandbaggers", err)
			return
		}
		detailsTable := helpers.FormatNthColumnList(1, sandbaggers.NewSandbaggers, table.Row{"", "Score", "Personal Best", "% of", "Median", "% of"}, func(data data.WeeklySandbaggersStats, idx int) table.Row {
			diffpb := strconv.Itoa(data.DiffPbPercentage) + "%"
			diffMd := strconv.Itoa(data.DiffMedianPercentage) + "%"
			return table.Row{data.Name, data.Score, data.RawStats.PersonalBest, diffpb, data.RawStats.Median, diffMd}
		})
		sandbaggersTable = &detailsTable

		detailsZeroScoreCharas := strings.Join(sandbaggers.ZeroScoreSandbaggers, ",")
		sandbaggersZeroScoreList = &detailsZeroScoreCharas
	}

	shouldShowWeeklyRats := false
	showRatsRaw := apiredis.OPTIONAL_CONF_SUBMIT_SCORES_SHOW_RATS.GetWithDefault(apiredis.RedisDB, "false")
	json.Unmarshal([]byte(showRatsRaw), &shouldShowWeeklyRats)
	// if it is not "true", always treat false, so ignore error

	rats := []struct {
		SixSeven string
		LWeeks   int
		WWeeks   int
	}{}
	calculateRatWeeks := 12
	if shouldShowWeeklyRats {
		rats, err = cmdhelpers.GetStinkyRats(db, charactersModel, submittedDate.Format("2006-01-02"), calculateRatWeeks, float64(4)/float64(12), "zero")
	}

	for _, v := range channelID {
		s.ChannelMessageSendComplex(v, &discordgo.MessageSend{
			Content: "Culvert scores updated! These are the changes from " + lastWeek.Format("2006-01-02") + " to " + submittedDate.Format("2006-01-02"),
			Files:   []*discordgo.File{{Name: "message.txt", Reader: strings.NewReader(rawStr)}},
		})
		if len(noLongerExistsFromLastWeek) > 0 {
			s.ChannelMessageSendComplex(v, &discordgo.MessageSend{
				Content: "These characters no longer exist in the last week: " + strings.Join(noLongerExistsFromLastWeek, ", "),
			})
		}

		if shouldShowWeeklySandbaggers {
			s.ChannelMessageSendComplex(v, &discordgo.MessageSend{
				Content: "Sandbaggers for " + submittedDate.Format("2006-01-02") + ", median over " + strconv.Itoa(medianWeeks) + " weeks",
				Files:   []*discordgo.File{{Name: "sandbaggers.txt", Reader: strings.NewReader(*sandbaggersTable)}, {Name: "zero-scores.txt", Reader: strings.NewReader(*sandbaggersZeroScoreList)}},
			})
		}

		if shouldShowWeeklyRats {
			if len(rats) > 0 {
				contentInner := ""
				for _, v := range rats {
					contentInner += v.SixSeven + " has a high amount roller coaster pattern count! " + strconv.Itoa(v.LWeeks) + "/" + strconv.Itoa(v.WWeeks) + "\n"
				}
				s.ChannelMessageSendComplex(v, &discordgo.MessageSend{
					Content: "Rats for " + submittedDate.Format("2006-01-02") + ", calculated over " + strconv.Itoa(calculateRatWeeks) + " weeks",
					Files:   []*discordgo.File{{Name: "stinky-rats.txt", Reader: strings.NewReader(contentInner)}},
				})
			} else {
				s.ChannelMessageSendComplex(v, &discordgo.MessageSend{
					Content: "Squeaky clean! No rats this week!",
				})
			}
		}
	}

}
