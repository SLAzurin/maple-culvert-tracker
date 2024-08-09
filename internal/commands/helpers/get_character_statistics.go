package helpers

//lint:file-ignore ST1001 Dot imports by jet
import (
	"database/sql"

	// . "github.com/go-jet/jet/v2/postgres"
	// . "github.com/slazurin/maple-culvert-tracker/.gen/mapleculverttrackerdb/public/table"
	"github.com/slazurin/maple-culvert-tracker/internal/data"
	// "github.com/slazurin/maple-culvert-tracker/.gen/mapleculverttrackerdb/public/model"
)

func GetCharacterStatistics(db *sql.DB, characterName string, weeks int64, date string) (*data.CharacterStatistics, error) {
	r := data.CharacterStatistics{
		PersonalBest:              60000,
		ParticipationCountLabel:   "12/12",
		ParticipationPercentRatio: 100,
		GuildTopPlacement:         1,
	}
	// stmt := SELECT(Characters.AllColumns).FROM(
	// 	Characters,
	// ).WHERE(LOWER(String(discordID)).EQ(LOWER(Characters.DiscordUserID)))
	// char := []model.Characters{}

	// err := stmt.Query(db, &char)
	// if err != nil {
	// 	return nil, err
	// }

	return &r, nil
}
