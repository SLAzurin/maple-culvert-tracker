package helpers

//lint:file-ignore ST1001 Dot imports by jet
import (
	"database/sql"
	"log"
	"math"
	"strconv"
	"time"

	. "github.com/go-jet/jet/v2/postgres"
	. "github.com/slazurin/maple-culvert-tracker/.gen/mapleculverttrackerdb/public/table"
	"github.com/slazurin/maple-culvert-tracker/internal/data"
)

var patch2mDate, _ = time.Parse("2006-01-02", "2025-10-01")

func GetCharacterStatistics(db *sql.DB, characterName string, date string, chartData []data.ChartMakerPoints) (*data.CharacterStatistics, error) {
	r := data.CharacterStatistics{}
	dateRaw, err := time.Parse("2006-01-02", date)
	if err != nil {
		dateRaw, err = GetLatestResetDate(db)
		if err != nil {
			log.Println(err)
			return nil, err
		}
	}
	stmt := SELECT(MAX(CharacterCulvertScores.Score).AS("personal_best")).FROM(
		CharacterCulvertScores.INNER_JOIN(Characters, Characters.ID.EQ(CharacterCulvertScores.CharacterID)),
	).WHERE(LOWER(String(characterName)).EQ(LOWER(Characters.MapleCharacterName)).AND(CharacterCulvertScores.CulvertDate.LT_EQ(DateT(dateRaw))))
	pb := struct {
		PersonalBest int64 `sql:"personal_best"`
	}{}

	err = stmt.Query(db, &pb)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	p := struct {
		Placement int32 `sql:"placement"`
	}{}
	if chartData[len(chartData)-1].Score != 0 {
		stmt = SELECT(COUNT(CharacterCulvertScores.Score).AS("placement")).FROM(CharacterCulvertScores.INNER_JOIN(Characters, Characters.ID.EQ(CharacterCulvertScores.CharacterID))).WHERE(CharacterCulvertScores.Score.GT_EQ(Int32(int32(chartData[len(chartData)-1].Score))).AND(CharacterCulvertScores.CulvertDate.EQ(DateT(dateRaw))))

		err = stmt.Query(db, &p)
		if err != nil {
			log.Println(err)
			return nil, err
		}
	}

	avg := int64(0)
	lastKnownGoodScore := int64(10)
	validCount := len(chartData)
	for _, v := range chartData {
		if v.Score == 0 {
			continue
		}
		lastKnownGoodScore = int64(v.Score)
		break
	}

	week1IsBefore2mPatch := false
	if culvertDate, _ := time.Parse("2006-01-02", chartData[0].RawDate); culvertDate.Before(patch2mDate) || culvertDate.Equal(patch2mDate) {
		week1IsBefore2mPatch = true
	}
	for _, v := range chartData {
		avg += int64(v.Score)
		if week1IsBefore2mPatch {
			culvertDate, _ := time.Parse("2006-01-02", v.RawDate)
			if culvertDate.After(patch2mDate) || culvertDate.Equal(patch2mDate) {
				week1IsBefore2mPatch = false // This ensures we fallback into the else block for the rest of the chartData, no need to re-parse the culvertDate again
				if v.Score <= 0 {
					validCount -= 1
					lastKnownGoodScore = int64(10)
				} else {
					lastKnownGoodScore = int64(v.Score)
				}
				continue
			}
		}
		threshold := GetSandbagThreshold(lastKnownGoodScore)
		if int64(v.Score) < threshold {
			validCount -= 1
		}
		if int64(v.Score) > lastKnownGoodScore {
			lastKnownGoodScore = int64(v.Score)
		}

	}
	avg /= int64(len(chartData))

	r.Average = int(avg)
	r.ParticipationCountLabel = strconv.Itoa(validCount) + "/" + strconv.Itoa(len(chartData))
	r.ParticipationPercentRatio = int(math.Round(float64(validCount) / float64(len(chartData)) * 100))
	r.PersonalBest = int(pb.PersonalBest)
	r.GuildTopPlacement = int(p.Placement)

	return &r, nil
}
