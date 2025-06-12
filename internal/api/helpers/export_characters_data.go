package helpers

//lint:file-ignore ST1001 Dot imports by jet
import (
	"database/sql"
	"errors"
	"log"
	"time"

	cmdhelpers "github.com/slazurin/maple-culvert-tracker/internal/commands/helpers"

	. "github.com/go-jet/jet/v2/postgres"
	. "github.com/slazurin/maple-culvert-tracker/.gen/mapleculverttrackerdb/public/table"
)

func ExportCharactersData(db *sql.DB, weeks int, asOf time.Time) ([]struct {
	Name   string
	Scores []struct {
		Date  string
		Score int
	}
	Average      int
	PreviousBest int
}, error) {
	retVal := []struct {
		Name   string
		Scores []struct {
			Date  string
			Score int
		}
		Average      int
		PreviousBest int
	}{}

	if weeks < 1 {
		return retVal, errors.New("weeks must be greater than 0")
	}
	thisWeek := cmdhelpers.GetCulvertResetDate(asOf)
	asOf = thisWeek
	weeksDate := []Expression{}
	weeksDateRaw := []time.Time{}
	for i := 0; i < weeks; i++ {
		weeksDateRaw = append(weeksDateRaw, thisWeek)
		weeksDate = append(weeksDate, DateT(thisWeek))
		thisWeek = cmdhelpers.GetCulvertPreviousDate(thisWeek)
	}

	// Fetch all characters for the latest week

	rows, err := db.Query("SELECT character_id FROM character_culvert_scores WHERE culvert_date = $1", asOf.Format("2006-01-02"))
	if err != nil {
		return retVal, err
	}

	characterIDs := []Expression{}
	for rows.Next() {
		var characterID int
		err = rows.Scan(&characterID)
		if err != nil {
			return retVal, err
		}
		characterIDs = append(characterIDs, Int(int64(characterID)))
	}

	if len(characterIDs) < 1 {
		return retVal, errors.New("no characters found")
	}

	log.Println("Character count", len(characterIDs))

	// Fetch scores for characters for the last n weeks
	stmt := SELECT(CharacterCulvertScores.CulvertDate.AS("culvert_date"), CharacterCulvertScores.Score.AS("score"), Characters.MapleCharacterName.AS("maple_character_name")).FROM(
		CharacterCulvertScores.INNER_JOIN(Characters, Characters.ID.EQ(CharacterCulvertScores.CharacterID))).WHERE(
		CharacterCulvertScores.CulvertDate.IN(
			weeksDate...,
		).AND(CharacterCulvertScores.CharacterID.IN(characterIDs...))).ORDER_BY(Characters.MapleCharacterName.ASC(),
		CharacterCulvertScores.CulvertDate.ASC())

	dest := []struct {
		CulvertDate        string
		Score              int32
		MapleCharacterName string
	}{}

	err = stmt.Query(db, &dest)
	if err != nil {
		return nil, err
	}

	m := map[string]struct {
		Scores []struct {
			Label string `json:"label"`
			Score int    `json:"score"`
		}
		Average      int
		PreviousBest int
	}{}

	mScoresOnly := map[string]map[string]int{}
	for _, v := range dest {
		if _, ok := m[v.MapleCharacterName]; !ok {
			m[v.MapleCharacterName] = struct {
				Scores []struct {
					Label string `json:"label"`
					Score int    `json:"score"`
				}
				Average      int
				PreviousBest int
			}{
				Scores: []struct {
					Label string `json:"label"`
					Score int    `json:"score"`
				}{},
				Average:      0,
				PreviousBest: 0,
			}
			mScoresOnly[v.MapleCharacterName] = map[string]int{}
		}
		newData := mScoresOnly[v.MapleCharacterName]
		newData[v.CulvertDate[:10]] = int(v.Score)
		mScoresOnly[v.MapleCharacterName] = newData
	}

	for name, data := range m {
		for _, date := range weeksDateRaw {
			if _, ok := mScoresOnly[name][date.Format("2006-01-02")]; !ok {
				data.Scores = append(data.Scores, struct {
					Label string `json:"label"`
					Score int    `json:"score"`
				}{
					Label: date.Format("2006-01-02"),
					Score: 0,
				})
			} else {
				data.Scores = append(data.Scores, struct {
					Label string `json:"label"`
					Score int    `json:"score"`
				}{
					Label: date.Format("2006-01-02"),
					Score: mScoresOnly[name][date.Format("2006-01-02")],
				})
			}
		}
		stats, err := cmdhelpers.GetCharacterStatistics(db, name, asOf.Format("2006-01-02"), data.Scores)
		if err != nil {
			panic(err)
		}

		scores := []struct {
			Date  string
			Score int
		}{}
		for _, v := range data.Scores {
			scores = append(scores, struct {
				Date  string
				Score int
			}{
				Date:  v.Label,
				Score: v.Score,
			})
		}
		retVal = append(retVal,
			struct {
				Name   string
				Scores []struct {
					Date  string
					Score int
				}
				Average      int
				PreviousBest int
			}{
				Name:         name,
				Scores:       scores,
				Average:      stats.Average,
				PreviousBest: stats.PersonalBest,
			},
		)
	}

	return retVal, nil
}
