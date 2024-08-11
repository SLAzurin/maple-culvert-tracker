package commands

//lint:file-ignore ST1001 Dot imports by jet

import (
	"encoding/json"
	"errors"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	. "github.com/go-jet/jet/v2/postgres"
	"github.com/go-jet/jet/v2/qrm"
	. "github.com/slazurin/maple-culvert-tracker/.gen/mapleculverttrackerdb/public/table"

	"github.com/slazurin/maple-culvert-tracker/internal/api/helpers"
	"github.com/slazurin/maple-culvert-tracker/internal/apiredis"
	cmdhelpers "github.com/slazurin/maple-culvert-tracker/internal/commands/helpers"
	"github.com/slazurin/maple-culvert-tracker/internal/db"
)

func getSandbaggers() *discordgo.InteractionResponse {
	chars, err := cmdhelpers.GetAcviveCharacters(apiredis.RedisDB, db.DB)
	if err != nil {
		return &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Sandbaggers command failed, database error while getting all active characters wtf?",
			},
		}
	}

	allSandbaggedRuns := []struct {
		Name           string
		SandbaggedRuns []struct {
			S int32  `json:"s"`
			D string `json:"d"`
		}
		SandbaggedRunsCount int
		TotalRuns           int
		ParticipationRatio  string
	}{}

	neverRanCulvert := ""

	for _, v := range *chars {

		t := SELECT(
			CharacterCulvertScores.CulvertDate,
			CharacterCulvertScores.Score).
			FROM(CharacterCulvertScores).
			WHERE(
				CharacterCulvertScores.CharacterID.EQ(Int64(v.ID)),
			).AsTable("t")
		tCulvertDate := CharacterCulvertScores.CulvertDate.From(t)
		tScore := CharacterCulvertScores.Score.From(t)
		// t is all character's scores

		cd := SELECT(
			CharacterCulvertScores.CulvertDate,
		).FROM(
			CharacterCulvertScores,
		).GROUP_BY(
			CharacterCulvertScores.CulvertDate,
		).ORDER_BY(
			CharacterCulvertScores.CulvertDate.DESC(),
		).LIMIT(12).AsTable("cd")

		cdCulvertDate := CharacterCulvertScores.CulvertDate.From(cd)

		stmt := SELECT(cdCulvertDate.AS("culvert_date"), COALESCE(tScore, Int(0)).AS("score")).FROM(cd.LEFT_JOIN(t, tCulvertDate.EQ(cdCulvertDate))).ORDER_BY(cdCulvertDate.ASC())

		dest := []struct {
			CulvertDate time.Time
			Score       int32
		}{}
		err = stmt.Query(db.DB, &dest)
		if err != nil {
			return &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Sandbaggers command failed, database error while getting a character's recent scores wtf?",
				},
			}
		}

		stmt = SELECT(MIN(CharacterCulvertScores.CulvertDate).AS("start_date")).FROM(CharacterCulvertScores).WHERE(CharacterCulvertScores.CharacterID.EQ(Int64(v.ID))).GROUP_BY(CharacterCulvertScores.CulvertDate).ORDER_BY(CharacterCulvertScores.CulvertDate.ASC())

		var initial struct {
			StartDate time.Time
		}
		err = stmt.Query(db.DB, &initial)
		if err != nil {
			if errors.Is(err, qrm.ErrNoRows) {
				// This guy is beyond sandbag
				neverRanCulvert += v.MapleCharacterName + " ∞/∞ [\"never ever ran culvert\"]\n"
				continue
			}
			return &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Sandbaggers command failed, database error while getting a character's first score date wtf?",
				},
			}
		}

		dest = helpers.FilterSlice(dest, func(v struct {
			CulvertDate time.Time
			Score       int32
		}) bool {
			return v.CulvertDate.After(initial.StartDate)
		})

		sandbaggedRuns := struct {
			Name           string
			SandbaggedRuns []struct {
				S int32  `json:"s"`
				D string `json:"d"`
			}
			SandbaggedRunsCount int
			TotalRuns           int
			ParticipationRatio  string
		}{
			Name: v.MapleCharacterName,
			SandbaggedRuns: []struct {
				S int32  `json:"s"`
				D string `json:"d"`
			}{},
			SandbaggedRunsCount: 0,
			TotalRuns:           len(dest),
			ParticipationRatio:  "",
		}

		lastKnownGoodScore := int64(0)
		for _, v := range dest {
			if v.Score == 0 {
				continue
			}
			lastKnownGoodScore = int64(v.Score)
			break
		}

		// sandbag algo: sandbagged scores are scores that fall below 70% of the lastKnownGoodScore or 10k difference as the threshold
		for _, v := range dest {
			threshold := cmdhelpers.GetSandbagThreshold(lastKnownGoodScore)
			if int64(v.Score) <= threshold {
				sandbaggedRuns.SandbaggedRunsCount += 1
				sandbaggedRuns.SandbaggedRuns = append(sandbaggedRuns.SandbaggedRuns, struct {
					S int32  "json:\"s\""
					D string "json:\"d\""
				}{D: v.CulvertDate.Format("2006-01-02"), S: v.Score})
			}
			if v.Score > int32(lastKnownGoodScore) {
				lastKnownGoodScore = int64(v.Score)
			}
		}

		sandbaggedRuns.ParticipationRatio = strconv.Itoa(sandbaggedRuns.SandbaggedRunsCount) + "/" + strconv.Itoa(sandbaggedRuns.TotalRuns)
		if sandbaggedRuns.SandbaggedRunsCount > 0 {
			allSandbaggedRuns = append(allSandbaggedRuns, sandbaggedRuns)
		}
	}

	slices.SortStableFunc(allSandbaggedRuns, func(a struct {
		Name           string
		SandbaggedRuns []struct {
			S int32  "json:\"s\""
			D string "json:\"d\""
		}
		SandbaggedRunsCount int
		TotalRuns           int
		ParticipationRatio  string
	}, b struct {
		Name           string
		SandbaggedRuns []struct {
			S int32  "json:\"s\""
			D string "json:\"d\""
		}
		SandbaggedRunsCount int
		TotalRuns           int
		ParticipationRatio  string
	}) int {
		return a.SandbaggedRunsCount - b.SandbaggedRunsCount
	})
	slices.Reverse(allSandbaggedRuns)

	s := ""

	for _, v := range allSandbaggedRuns {
		slices.Reverse(v.SandbaggedRuns)
		ds, _ := json.Marshal(v.SandbaggedRuns)
		s += v.Name + " " + v.ParticipationRatio + " " + string(ds) + "\n"
	}
	s = neverRanCulvert + s

	return &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Sandbaggers of the week as follows:",
			Files:   []*discordgo.File{{Name: "message.txt", Reader: strings.NewReader(s)}},
		},
	}
}
