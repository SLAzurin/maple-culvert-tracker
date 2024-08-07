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

func GetCharacterStatistics(db *sql.DB, characterName string, date string, chartData []data.ChartMakerPoints) (*data.CharacterStatistics, error) {
	r := data.CharacterStatistics{}
	var dateRaw time.Time
	var err error
	dateRaw, err = time.Parse("2006-01-02", date)
	if err != nil {
		dateRaw = time.Now()
	}

	for dateRaw.Weekday() != time.Sunday {
		dateRaw = dateRaw.Add(time.Hour * -24)
	}
	stmt := SELECT(MAX(CharacterCulvertScores.Score).AS("personal_best")).FROM(
		CharacterCulvertScores.INNER_JOIN(Characters, Characters.ID.EQ(CharacterCulvertScores.CharacterID)),
	).WHERE(LOWER(String(characterName)).EQ(LOWER(Characters.MapleCharacterName)))
	pb := struct {
		PersonalBest int64 `sql:"personal_best"`
	}{}

	err = stmt.Query(db, &pb)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	stmt = SELECT(COUNT(CharacterCulvertScores.Score).AS("placement")).FROM(CharacterCulvertScores.INNER_JOIN(Characters, Characters.ID.EQ(CharacterCulvertScores.CharacterID))).WHERE(CharacterCulvertScores.Score.GT_EQ(Int32(int32(chartData[len(chartData)-1].Score))).AND(CharacterCulvertScores.CulvertDate.EQ(DateT(dateRaw))))
	p := struct {
		Placement int32 `sql:"placement"`
	}{}
	err = stmt.Query(db, &p)
	if err != nil {
		log.Println(err)
		return nil, err
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

	for _, v := range chartData {
		if int64(v.Score) < int64(float64(lastKnownGoodScore)*.7) {
			validCount -= 1
		}
		if int64(v.Score) > int64(lastKnownGoodScore) {
			lastKnownGoodScore = int64(v.Score)
		}
		avg += int64(v.Score)
	}
	avg /= int64(len(chartData))

	r.Average = int(avg)
	r.ParticipationCountLabel = strconv.Itoa(validCount) + "/" + strconv.Itoa(len(chartData))
	r.ParticipationPercentRatio = int(math.Round(float64(validCount) / float64(len(chartData)) * 100))
	r.PersonalBest = int(pb.PersonalBest)
	r.GuildTopPlacement = int(p.Placement)

	return &r, nil
}
