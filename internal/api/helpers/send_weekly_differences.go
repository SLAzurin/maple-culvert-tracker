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
	"github.com/go-jet/jet/v2/qrm"
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

func SendWeeklyDifferences(s *discordgo.Session, db *sql.DB, rdb *redis.Client, submittedDate time.Time, adminsTextChannel string, membersTextChannel string) {
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
	culvertScoresByCharacterID := map[int64]struct {
		Name      string
		ThisScore int
		PrevBest  int
	}{}

	for curPos, v := range rawData {
		if submittedDate.Format("2006-01-02") == v.CulvertDate.Format("2006-01-02") {
			characters = append(characters, v.Name)
			charactersModel = append(charactersModel, model.Characters{MapleCharacterName: v.Name, ID: v.ID})
			culvertScoresByCharacterID[v.ID] = struct {
				Name      string
				ThisScore int
				PrevBest  int
			}{Name: v.Name, ThisScore: v.Score, PrevBest: -1}
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

	// Get previous best scores
	characterIDsNewPb := []int{}
	characterIds := []Expression{}
	for k := range culvertScoresByCharacterID {
		characterIds = append(characterIds, Int(int64(k)))
	}
	whereClause := CharacterCulvertScores.CharacterID.IN(characterIds...).AND(CharacterCulvertScores.CulvertDate.LT(DateT(submittedDate)))
	if submittedDate.After(data.Date2mPatch) {
		whereClause = whereClause.AND(CharacterCulvertScores.CulvertDate.GT_EQ(DateT(data.Date2mPatch)))
	}
	stmt = SELECT(CharacterCulvertScores.CharacterID.AS("character_id"), MAX(CharacterCulvertScores.Score).AS("max_score")).FROM(CharacterCulvertScores).WHERE(whereClause).GROUP_BY(CharacterCulvertScores.CharacterID)
	prevBests := []struct {
		CharacterID int64
		MaxScore    int
	}{}
	err = stmt.Query(db, &prevBests)
	if err != nil {
		log.Println("DB ERROR SendWeeklyDifferences fetching previous bests", err)
		return
	}
	for _, v := range prevBests {
		if v.MaxScore < culvertScoresByCharacterID[v.CharacterID].ThisScore {
			// This is a new personal best
			culvertScoresByCharacterID[v.CharacterID] = struct {
				Name      string
				ThisScore int
				PrevBest  int
			}{
				Name:      culvertScoresByCharacterID[v.CharacterID].Name,
				ThisScore: culvertScoresByCharacterID[v.CharacterID].ThisScore,
				PrevBest:  v.MaxScore,
			}
			characterIDsNewPb = append(characterIDsNewPb, int(v.CharacterID))
		}
	}
	// done getting previous bests

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

	s.ChannelMessageSendComplex(adminsTextChannel, &discordgo.MessageSend{
		Content: "Culvert scores updated! These are the changes from " + lastWeek.Format("2006-01-02") + " to " + submittedDate.Format("2006-01-02"),
		Files:   []*discordgo.File{{Name: "message.txt", Reader: strings.NewReader(rawStr)}},
	})
	if len(noLongerExistsFromLastWeek) > 0 {
		s.ChannelMessageSendComplex(adminsTextChannel, &discordgo.MessageSend{
			Content: "These characters no longer exist in the last week: " + strings.Join(noLongerExistsFromLastWeek, ", "),
		})
	}

	if shouldShowWeeklySandbaggers {
		s.ChannelMessageSendComplex(adminsTextChannel, &discordgo.MessageSend{
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
			s.ChannelMessageSendComplex(adminsTextChannel, &discordgo.MessageSend{
				Content: "Rats for " + submittedDate.Format("2006-01-02") + ", calculated over " + strconv.Itoa(calculateRatWeeks) + " weeks",
				Files:   []*discordgo.File{{Name: "stinky-rats.txt", Reader: strings.NewReader(contentInner)}},
			})
		} else {
			s.ChannelMessageSendComplex(adminsTextChannel, &discordgo.MessageSend{
				Content: "Squeaky clean! No rats this week!",
			})
		}
	}

	// Fetch congratz message and send if exists
	stmt = SELECT(DiscordPbAnnounecments.MessageID).FROM(DiscordPbAnnounecments).WHERE(DiscordPbAnnounecments.Week.EQ(DateT(submittedDate))).LIMIT(1)
	msgModel := model.DiscordPbAnnounecments{}
	err = stmt.Query(db, &msgModel)
	firstEntry := false
	if err == qrm.ErrNoRows {
		firstEntry = true
	} else if err != nil {
		log.Println("DB ERROR SendWeeklyDifferences fetching congratz message", err)
		s.ChannelMessageSendComplex(adminsTextChannel, &discordgo.MessageSend{
			Content: "Failed to post congratz message due to internal error. See server logs.",
		})
		return
	}

	if msgModel.MessageID == "" && !firstEntry {
		// Failed to find message
		log.Println("DB ERROR SendWeeklyDifferences fetching congratz message, no message id found, but not first entry")
		s.ChannelMessageSendComplex(adminsTextChannel, &discordgo.MessageSend{
			Content: "Failed to post congratz message due to GIGA TERRIBLE internal error. See server logs.",
		})
		return
	}

	congratsMsg := "Congratulations to the following characters for achieving a new personal best this week of " + submittedDate.Format("2006-01-02") + "!"

	attachmentMsg := "New Personal Bests:\n"
	for _, v := range characterIDsNewPb {
		scoreData := culvertScoresByCharacterID[int64(v)]
		attachmentMsg += scoreData.Name + ": " + strconv.Itoa(scoreData.PrevBest) + " -> " + strconv.Itoa(scoreData.ThisScore) + "\n"
	}

	if firstEntry {
		// Send message for first entry only
		msg, err := s.ChannelMessageSendComplex(membersTextChannel, &discordgo.MessageSend{
			Content: congratsMsg,
			Files: []*discordgo.File{
				{
					Name:   "new-pbs.txt",
					Reader: strings.NewReader(attachmentMsg),
				},
			},
		})
		if err != nil {
			log.Println("Discord ERROR SendWeeklyDifferences sending first congratz message", err)
			return
		}
		// Store message ID
		newMsgModel := model.DiscordPbAnnounecments{
			Week:      submittedDate,
			MessageID: msg.ID,
		}
		_, err = DiscordPbAnnounecments.INSERT(DiscordPbAnnounecments.Week, DiscordPbAnnounecments.MessageID).VALUES(newMsgModel.Week, newMsgModel.MessageID).Exec(db)
		if err != nil {
			log.Println("DB ERROR SendWeeklyDifferences inserting first congratz message", err)
			return
		}
	} else {
		// Edit existing message
		_, err := s.ChannelMessageEditComplex(&discordgo.MessageEdit{
			Channel: membersTextChannel,
			ID:      msgModel.MessageID,
			Content: &congratsMsg,
			Files: []*discordgo.File{
				{
					Name:   "new-pbs.txt",
					Reader: strings.NewReader(attachmentMsg),
				},
			},
		})
		if err != nil {
			log.Println("Discord ERROR SendWeeklyDifferences editing congratz message", err)
			return
		}
	}

}
