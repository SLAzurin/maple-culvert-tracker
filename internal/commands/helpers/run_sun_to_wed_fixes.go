package helpers

//lint:file-ignore ST1001 Dot imports by jet
import (
	"database/sql"
	"log"
	"time"

	. "github.com/go-jet/jet/v2/postgres"
	. "github.com/slazurin/maple-culvert-tracker/.gen/mapleculverttrackerdb/public/table"
)

func RunSunToWedFixes(db *sql.DB) {
	// select all dates that are sundays after 2024-08-25
	stmt := SELECT(CharacterCulvertScores.CulvertDate.AS("culvert_date")).FROM(CharacterCulvertScores).WHERE(CharacterCulvertScores.CulvertDate.GT(Date(2024, 8, 25))).GROUP_BY(CharacterCulvertScores.CulvertDate)

	dates := []struct {
		CulvertDate time.Time
	}{}
	err := stmt.Query(db, &dates)
	if err != nil {
		log.Println("Failed to RunSunToWedFixes getting culvert date", err)
		return
	}

	wrongDates := []struct {
		CulvertDate time.Time
	}{}

	for _, v := range dates {
		if v.CulvertDate.Weekday() != GetCulvertResetDay(v.CulvertDate) {
			wrongDates = append(wrongDates, v)
		}
	}

	for _, v := range wrongDates {
		// convert to wednesday
		rawDate := GetCulvertResetDate(v.CulvertDate)
		updateStmt := CharacterCulvertScores.UPDATE(CharacterCulvertScores.CulvertDate).SET(Date(rawDate.Year(), rawDate.Month(), rawDate.Day())).WHERE(CharacterCulvertScores.CulvertDate.EQ(DateT(v.CulvertDate)))
		_, err := updateStmt.Exec(db)
		if err != nil {
			log.Println("Failed to RunSunToWedFixes update date for wrong dates", err)
			return
		}
	}

	log.Println("RunSunToWedFixes done")

}
