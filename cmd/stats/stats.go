package main

//lint:file-ignore ST1001 Dot imports by jet

import (
	"fmt"
	"slices"
	"strconv"
	"time"

	. "github.com/go-jet/jet/v2/postgres"
	. "github.com/slazurin/maple-culvert-tracker/.gen/mapleculverttrackerdb/public/table"

	_ "github.com/joho/godotenv/autoload"
	"github.com/slazurin/maple-culvert-tracker/internal/db"
)

func main() {
	stmt := SELECT(MAX(CharacterCulvertScores.CulvertDate).AS("culvert_date")).FROM(CharacterCulvertScores)
	dest := struct {
		CulvertDate time.Time
	}{}
	stmt.Query(db.DB, &dest)
	sunday := dest.CulvertDate

	last12WeeksCulvertRaw := []time.Time{}
	for i := 0; i < 12; i++ {
		last12WeeksCulvertRaw = append(last12WeeksCulvertRaw, sunday)
		sunday = sunday.Add(time.Hour * -24 * 7)
	}

	stmt = SELECT(CharacterCulvertScores.CharacterID.AS("character_id"), Characters.MapleCharacterName.AS("maple_character_name")).FROM(
		CharacterCulvertScores.INNER_JOIN(Characters, Characters.ID.EQ(CharacterCulvertScores.CharacterID)),
	).WHERE(CharacterCulvertScores.CulvertDate.EQ(DateT(last12WeeksCulvertRaw[0])))

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
		inClauseDates := []Expression{}
		for _, date := range last12WeeksCulvertRaw {
			inClauseDates = append(inClauseDates, DateT(date))
		}

		stmt := SELECT(
			CharacterCulvertScores.CulvertDate.AS("culvert_date"),
			CharacterCulvertScores.Score.AS("score")).
			FROM(CharacterCulvertScores).
			WHERE(
				CharacterCulvertScores.CharacterID.EQ(Int64(v.CharacterID)).AND(CharacterCulvertScores.CulvertDate.IN(inClauseDates...)),
			).
			ORDER_BY(
				CharacterCulvertScores.CulvertDate.ASC(),
			)

		dest := []struct {
			CulvertDate time.Time
			Score       int32
		}{}
		stmt.Query(db.DB, &dest)

		if len(dest) < 1 {
			continue
		}
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
