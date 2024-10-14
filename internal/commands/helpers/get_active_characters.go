package helpers

//lint:file-ignore ST1001 Dot imports by jet
import (
	"database/sql"
	"encoding/json"

	. "github.com/go-jet/jet/v2/postgres"
	. "github.com/slazurin/maple-culvert-tracker/.gen/mapleculverttrackerdb/public/table"

	"github.com/redis/go-redis/v9"
	"github.com/slazurin/maple-culvert-tracker/.gen/mapleculverttrackerdb/public/model"
	"github.com/slazurin/maple-culvert-tracker/internal/apiredis"
	"github.com/slazurin/maple-culvert-tracker/internal/data"
)

func GetAcviveCharacters(r *redis.Client, db *sql.DB) (*[]model.Characters, error) {
	discordIDsFullRaw, err := apiredis.DATA_DISCORD_MEMBERS.Get(r)
	if err != nil {
		return nil, err
	}
	discordIDsFull := []data.WebGuildMember{}
	err = json.Unmarshal([]byte(discordIDsFullRaw), &discordIDsFull)
	if err != nil {
		return nil, err
	}

	discordIDs := []Expression{}
	for _, v := range discordIDsFull {
		discordIDs = append(discordIDs, String(v.DiscordUserID))
	}
	stmt := SELECT(Characters.AllColumns).FROM(
		Characters,
	).WHERE(Characters.DiscordUserID.IN(discordIDs...)).ORDER_BY(Characters.MapleCharacterName)
	chars := []model.Characters{}

	err = stmt.Query(db, &chars)
	if err != nil {
		return nil, err
	}

	return &chars, nil
}
