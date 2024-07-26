package main

//lint:file-ignore ST1001 Dot imports by jet

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"strconv"
	"time"

	. "github.com/go-jet/jet/v2/postgres"
	. "github.com/slazurin/maple-culvert-tracker/.gen/mapleculverttrackerdb/public/table"

	_ "github.com/joho/godotenv/autoload"
	"github.com/slazurin/maple-culvert-tracker/internal/apiredis"
	"github.com/slazurin/maple-culvert-tracker/internal/data"
	"github.com/slazurin/maple-culvert-tracker/internal/db"
)

func filter[T any](ss []T, test func(T) bool) (ret []T) {
	for _, s := range ss {
		if test(s) {
			ret = append(ret, s)
		}
	}
	return
}

func main() {
	discordIDsFullRaw, _ := apiredis.RedisDB.Get(context.Background(), "discord_members_"+os.Getenv("DISCORD_GUILD_ID")).Result()

	discordIDsFull := []data.WebGuildMember{}
	json.Unmarshal([]byte(discordIDsFullRaw), &discordIDsFull)

	discordIDs := []Expression{}
	for _, v := range discordIDsFull {
		discordIDs = append(discordIDs, String(v.DiscordUserID))
	}

	// stmt := SELECT(MAX(CharacterCulvertScores.CulvertDate).AS("culvert_date")).FROM(CharacterCulvertScores)
	// dest := struct {
	// 	CulvertDate time.Time
	// }{}
	// stmt.Query(db.DB, &dest)
	// sunday := dest.CulvertDate

	// last12WeeksCulvertRaw := []time.Time{}
	// for i := 0; i < 12; i++ {
	// 	last12WeeksCulvertRaw = append(last12WeeksCulvertRaw, sunday)
	// 	sunday = sunday.Add(time.Hour * -24 * 7)
	// }

	stmt := SELECT(Characters.ID.AS("character_id"), Characters.MapleCharacterName.AS("maple_character_name")).FROM(
		Characters,
	).WHERE(Characters.DiscordUserID.IN(discordIDs...))

	chars := []struct {
		CharacterID        int64
		MapleCharacterName string
	}{}

	stmt.Query(db.DB, &chars)

	allSandbaggedRuns := []struct {
		Name                string
		SandbaggedRunsDates []string
		SandbaggedRunsCount int
		TotalRuns           int
		ParticipationRatio  string
	}{}

	for _, v := range chars {
		// inClauseDates := []Expression{}
		// for _, date := range last12WeeksCulvertRaw {
		// 	inClauseDates = append(inClauseDates, DateT(date))
		// }

		// select character_culvert_scores.culvert_date, t.score from character_culvert_scores left join (select culvert_date, score from character_culvert_scores where character_id = 111) as t on t.culvert_date = character_culvert_scores.culvert_date group by character_culvert_scores.culvert_date, t.score order by character_culvert_scores.culvert_date desc limit 12;

		t := SELECT(
			CharacterCulvertScores.CulvertDate,
			CharacterCulvertScores.Score).
			FROM(CharacterCulvertScores).
			WHERE(
				CharacterCulvertScores.CharacterID.EQ(Int64(v.CharacterID)),
			).AsTable("t")
		tCulvertDate := CharacterCulvertScores.CulvertDate.From(t)
		tScore := CharacterCulvertScores.Score.From(t)

		stmt := SELECT(
			CharacterCulvertScores.CulvertDate.AS("culvert_date"),
			COALESCE(tScore, Int(0)).AS("score"),
		).FROM(
			CharacterCulvertScores.LEFT_JOIN(t, tCulvertDate.EQ(CharacterCulvertScores.CulvertDate)),
		).GROUP_BY(
			CharacterCulvertScores.CulvertDate,
			tScore,
		).ORDER_BY(
			CharacterCulvertScores.CulvertDate.ASC(),
		).LIMIT(12)

		dest := []struct {
			CulvertDate time.Time
			Score       int32
		}{}
		stmt.Query(db.DB, &dest)

		stmt = SELECT(MIN(CharacterCulvertScores.CulvertDate).AS("start_date")).FROM(CharacterCulvertScores).WHERE(CharacterCulvertScores.CharacterID.EQ(Int64(v.CharacterID))).GROUP_BY(CharacterCulvertScores.CulvertDate).ORDER_BY(CharacterCulvertScores.CulvertDate.ASC())

		var initial struct {
			StartDate time.Time
		}
		stmt.Query(db.DB, &initial)

		dest = filter(dest, func(v struct {
			CulvertDate time.Time
			Score       int32
		}) bool {
			return v.CulvertDate.After(initial.StartDate)
		})

		sandbaggedRuns := struct {
			Name                string
			SandbaggedRunsDates []string
			SandbaggedRunsCount int
			TotalRuns           int
			ParticipationRatio  string
		}{
			Name:                v.MapleCharacterName,
			SandbaggedRunsDates: []string{},
			SandbaggedRunsCount: 0,
			TotalRuns:           len(dest),
			ParticipationRatio:  "",
		}

		lastKnownGoodScore := 0
		for _, v := range dest {
			if v.Score == 0 {
				continue
			}
			lastKnownGoodScore = int(v.Score)
			break
		}

		// sandbag algo: sandbagged scores are scores that fall below 70% of the lastKnownGoodScore
		for i, v := range dest {
			if i == 0 {
				if v.Score == 0 {
					sandbaggedRuns.SandbaggedRunsCount += 1
					sandbaggedRuns.SandbaggedRunsDates = append(sandbaggedRuns.SandbaggedRunsDates, v.CulvertDate.Format("2006-01-02"))
				}
				continue
			}
			if v.Score < int32(float64(lastKnownGoodScore)*.7) {
				sandbaggedRuns.SandbaggedRunsCount += 1
				sandbaggedRuns.SandbaggedRunsDates = append(sandbaggedRuns.SandbaggedRunsDates, v.CulvertDate.Format("2006-01-02"))
			}
			if v.Score > int32(lastKnownGoodScore) {
				lastKnownGoodScore = int(v.Score)
			}
		}

		sandbaggedRuns.ParticipationRatio = strconv.Itoa(sandbaggedRuns.SandbaggedRunsCount) + "/" + strconv.Itoa(sandbaggedRuns.TotalRuns)
		if sandbaggedRuns.SandbaggedRunsCount > 0 {
			allSandbaggedRuns = append(allSandbaggedRuns, sandbaggedRuns)
		}
	}

	slices.SortStableFunc(allSandbaggedRuns, func(a struct {
		Name                string
		SandbaggedRunsDates []string
		SandbaggedRunsCount int
		TotalRuns           int
		ParticipationRatio  string
	}, b struct {
		Name                string
		SandbaggedRunsDates []string
		SandbaggedRunsCount int
		TotalRuns           int
		ParticipationRatio  string
	}) int {
		return a.SandbaggedRunsCount - b.SandbaggedRunsCount
	})
	slices.Reverse(allSandbaggedRuns)

	for _, v := range allSandbaggedRuns {
		fmt.Println(v.Name, v.ParticipationRatio, v.SandbaggedRunsDates)
	}
}
