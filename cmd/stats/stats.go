package main

//lint:file-ignore ST1001 Dot imports by jet

import (
	"log"
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
				CharacterCulvertScores.CulvertDate.DESC(),
			)

		dest := []struct {
			CulvertDate time.Time
			Score       int32
		}{}
		stmt.Query(db.DB, &dest)

		log.Println(dest)
	}

	// sandbaggers for the last 12 weeks

	// sandbaggers this week
	// select maple_character_name from character_culvert_scores inner join characters on characters.id = character_culvert_scores.character_id where culvert_date = '2024-07-21' and score = 0;
}
