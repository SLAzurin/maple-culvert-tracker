package helpers

//lint:file-ignore ST1001 Dot imports by jet
import (
	"database/sql"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	. "github.com/go-jet/jet/v2/postgres"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/redis/go-redis/v9"
	. "github.com/slazurin/maple-culvert-tracker/.gen/mapleculverttrackerdb/public/table"
	cmdhelpers "github.com/slazurin/maple-culvert-tracker/internal/commands/helpers"
)

type differenceStruct struct {
	Name    string
	Prev    int
	Current int
	Oldpos  int
}

func SendWeeklyDifferences(s *discordgo.Session, db *sql.DB, rdb *redis.Client, submittedDate time.Time, channelID ...string) {
	submittedDate = cmdhelpers.GetCulvertResetDate(submittedDate)
	lastWeek := cmdhelpers.GetCulvertResetDate(submittedDate.Add(time.Hour * -24 * 7))

	// query all summary of current character scores
	stmt := SELECT(Characters.MapleCharacterName.AS("name"), CharacterCulvertScores.Score.AS("score"), CharacterCulvertScores.CulvertDate.AS("culvert_date")).FROM(CharacterCulvertScores.LEFT_JOIN(Characters, Characters.ID.EQ(CharacterCulvertScores.CharacterID))).WHERE(CharacterCulvertScores.CulvertDate.IN(DateT(lastWeek), DateT(submittedDate))).ORDER_BY(CharacterCulvertScores.CulvertDate.DESC(), CharacterCulvertScores.Score.DESC(), Characters.MapleCharacterName.ASC())

	rawData := []struct {
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
	diffs := []differenceStruct{}
	noLongerExistsFromLastWeek := []string{}
	cutoffPos := -1
	for curPos, v := range rawData {
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
			diffs = append(diffs, differenceStruct{
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

	rawStr := cmdhelpers.FormatNthColumnList(columnCount, diffs, table.Row{"Character", "Score", "Position"}, func(data differenceStruct, idx int) table.Row {
		return table.Row{data.Name, strconv.Itoa(data.Prev) + " -> " + strconv.Itoa(data.Current), strconv.Itoa(data.Oldpos) + " -> " + strconv.Itoa(idx+1)}
	})

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
	}

}
