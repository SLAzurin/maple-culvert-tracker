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
	Prev    int
	Current int
	Oldpos  int
	Newpos  int
}

func SendWeeklyDifferences(s *discordgo.Session, db *sql.DB, rdb *redis.Client, submittedDate time.Time, channelID ...string) {
	log.Println("Sending weekly differences to channel", channelID)
	log.Println("Submitted Date", submittedDate.Format("2006-01-02"))
	submittedDate = cmdhelpers.GetCulvertResetDate(submittedDate)
	lastWeek := cmdhelpers.GetCulvertResetDate(submittedDate.Add(time.Hour * -24 * 7))

	// query all summary of current character scores
	stmt := SELECT(Characters.MapleCharacterName.AS("name"), CharacterCulvertScores.Score.AS("score"), CharacterCulvertScores.CulvertDate.AS("culvert_date")).FROM(CharacterCulvertScores.INNER_JOIN(Characters, Characters.ID.EQ(CharacterCulvertScores.CharacterID))).WHERE(CharacterCulvertScores.CulvertDate.IN(DateT(lastWeek), DateT(submittedDate))).ORDER_BY(CharacterCulvertScores.CulvertDate.DESC(), CharacterCulvertScores.Score.DESC(), Characters.MapleCharacterName.ASC())

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

	differences := map[string]differenceStruct{}

	// compare the two and send the message as a file to the channel
	oldposIndex := 0
	newposIndex := 0
	for _, v := range rawData {
		if _, ok := differences[v.Name]; ok {
			differences[v.Name] = differenceStruct{}
		}
		if v.CulvertDate.Day() == lastWeek.Day() {
			oldposIndex++
			old := differences[v.Name]
			old.Oldpos = oldposIndex
			old.Prev = v.Score
			differences[v.Name] = old
		} else {
			newposIndex++
			old := differences[v.Name]
			old.Newpos = newposIndex
			old.Current = v.Score
			differences[v.Name] = old
		}
	}

	// convert differences map to slice
	differencesSlice := []struct {
		Name    string
		RawData *differenceStruct
	}{}
	for name, v := range differences {
		differencesSlice = append(differencesSlice, struct {
			Name    string
			RawData *differenceStruct
		}{name, &v})
	}

	// Send nice ascii table to channel
	columnCount := 1
	if newposIndex > 65 {
		columnCount = 2
	}

	if newposIndex > 130 {
		columnCount = 3
	}
	rawStr := cmdhelpers.FormatNthColumnList(columnCount, differencesSlice, table.Row{"Character", "Score", "Position"}, func(data struct {
		Name    string
		RawData *differenceStruct
	}) table.Row {
		return table.Row{data.Name, strconv.Itoa(data.RawData.Prev) + " => " + strconv.Itoa(data.RawData.Current), strconv.Itoa(data.RawData.Oldpos) + " => " + strconv.Itoa(data.RawData.Newpos)}
	})

	for _, v := range channelID {
		s.ChannelMessageSendComplex(v, &discordgo.MessageSend{
			Content: "Culvert scores updated! ANY GAINS SINCE LAST WEEK?",
			Files:   []*discordgo.File{{Name: "message.txt", Reader: strings.NewReader(rawStr)}},
		})
	}
}
