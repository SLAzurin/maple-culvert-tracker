package helpers

//lint:file-ignore ST1001 Dot imports by jet
import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"slices"
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
	characterNamePaddingTotal := 0
	culvertScorePaddingTotal := 0

	for curPos, v := range rawData {
		if submittedDate.Format("2006-01-02") == v.CulvertDate.Format("2006-01-02") {
			characters = append(characters, v.Name)
			charactersModel = append(charactersModel, model.Characters{MapleCharacterName: v.Name, ID: v.ID})
			culvertScoresByCharacterID[v.ID] = struct {
				Name      string
				ThisScore int
				PrevBest  int
			}{Name: v.Name, ThisScore: v.Score, PrevBest: -1}
			if len(v.Name) > characterNamePaddingTotal {
				characterNamePaddingTotal = len(v.Name)
			}
			if len(strconv.Itoa(v.Score)) > culvertScorePaddingTotal {
				culvertScorePaddingTotal = len(strconv.Itoa(v.Score))
			}
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
	hasNewPB := false
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
			hasNewPB = true
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

	if hasNewPB {
		// Fetch congratz message and send if exists
		stmt = SELECT(DiscordPbAnnounecments.MessageID, DiscordPbAnnounecments.Part, DiscordPbAnnounecments.TotalParts).FROM(DiscordPbAnnounecments).WHERE(DiscordPbAnnounecments.Week.EQ(DateT(submittedDate)))
		msgModel := []model.DiscordPbAnnounecments{}
		err = stmt.Query(db, &msgModel)
		if err != nil && err != qrm.ErrNoRows {
			log.Println("DB ERROR SendWeeklyDifferences fetching congratz message", err)
			s.ChannelMessageSendComplex(adminsTextChannel, &discordgo.MessageSend{
				Content: "Failed to post congratz message due to internal error. See server logs.",
			})
			return
		}
		firstEntry := true
		if len(msgModel) > 0 {
			firstEntry = false
		}

		congratzMsgPart := 0
		congratsMsgs := []string{""}
		congratsMsgs[congratzMsgPart] = "Congratulations to the following members for achieving a new personal best this week of " + submittedDate.Format("2006-01-02") + "!\n```"

		// sort characterIDsNewPb by new pb
		slices.SortStableFunc(characterIDsNewPb, func(a, b int) int {
			return culvertScoresByCharacterID[int64(b)].ThisScore - culvertScoresByCharacterID[int64(a)].ThisScore
		})

		for _, v := range characterIDsNewPb {
			scoreData := culvertScoresByCharacterID[int64(v)]

			nextCharaLine := fmt.Sprintf("%-"+strconv.Itoa(characterNamePaddingTotal)+"s: %"+strconv.Itoa(culvertScorePaddingTotal)+"s -> %"+strconv.Itoa(culvertScorePaddingTotal)+"s\n", scoreData.Name, strconv.Itoa(scoreData.PrevBest), strconv.Itoa(scoreData.ThisScore))
			// "`" + scoreData.Name + "`: " + strconv.Itoa(scoreData.PrevBest) + " -> " + strconv.Itoa(scoreData.ThisScore) + "\n"
			if len(congratsMsgs[congratzMsgPart])+len(nextCharaLine)+3 >= 2000 {
				// start new message part
				congratsMsgs[congratzMsgPart] += "```"
				congratzMsgPart++
				congratsMsgs = append(congratsMsgs, "```")
			}
			congratsMsgs[congratzMsgPart] += nextCharaLine
		}
		congratsMsgs[congratzMsgPart] += "```"

		// Check if msgs in msgModel exist
		for _, v := range msgModel {
			m, err := s.ChannelMessage(membersTextChannel, v.MessageID)
			if err != nil || m == nil {
				msgModel[0].TotalParts = int32(-1) // force delete entries from the next if statement block
				firstEntry = false
				break
			}
		}

		if !firstEntry && len(congratsMsgs) != int(msgModel[0].TotalParts) {
			// delete entries and message
			firstEntry = true
			for _, v := range msgModel {
				s.ChannelMessageDelete(membersTextChannel, v.MessageID)
				// if it fails, not a big deal, simply continue normally
			}
			stmt := DiscordPbAnnounecments.DELETE().WHERE(DiscordPbAnnounecments.Week.EQ(DateT(submittedDate)))
			_, err = stmt.Exec(db)
			if err != nil {
				log.Println("DB ERROR SendWeeklyDifferences deleting old congratz messages", err)
				s.ChannelMessageSendComplex(adminsTextChannel, &discordgo.MessageSend{
					Content: "Failed to cleanup internal data for congratz messages. See server logs.",
				})
				return
			}
		}

		if firstEntry {
			// Send message for first entry only
			for i, congratsMsg := range congratsMsgs {
				msg, err := s.ChannelMessageSendComplex(membersTextChannel, &discordgo.MessageSend{
					Content: congratsMsg,
				})
				if err != nil {
					log.Println("Discord ERROR SendWeeklyDifferences sending first congratz message", err)
					return
				}
				// Store message ID
				newMsgModel := model.DiscordPbAnnounecments{
					Week:       submittedDate,
					MessageID:  msg.ID,
					Part:       int32(i),
					TotalParts: int32(len(congratsMsgs)),
				}
				_, err = DiscordPbAnnounecments.INSERT(DiscordPbAnnounecments.Week, DiscordPbAnnounecments.MessageID, DiscordPbAnnounecments.Part, DiscordPbAnnounecments.TotalParts).VALUES(newMsgModel.Week, newMsgModel.MessageID, newMsgModel.Part, newMsgModel.TotalParts).Exec(db)
				if err != nil {
					log.Println("DB ERROR SendWeeklyDifferences inserting first congratz message", err)
					return
				}
			}
		} else {
			// Edit existing message
			for _, msg := range msgModel {
				_, err := s.ChannelMessageEditComplex(&discordgo.MessageEdit{
					Channel: membersTextChannel,
					ID:      msg.MessageID,
					Content: &congratsMsgs[msg.Part],
				})
				if err != nil {
					log.Println("Discord ERROR SendWeeklyDifferences editing congratz message", err)
					return
				}
			}
		}
	}
}
