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
	for i := 0; i < weeks; i++ {
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
		}
		newData := m[v.MapleCharacterName]
		newData.Scores = append(newData.Scores, struct {
			Label string `json:"label"`
			Score int    `json:"score"`
		}{
			Label: v.CulvertDate[:10],
			Score: int(v.Score),
		})
		m[v.MapleCharacterName] = newData
	}

	for name, data := range m {
		stats, err := cmdhelpers.GetCharacterStatistics(db, name, data.Scores[len(data.Scores)-1].Label, data.Scores)
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
