package commands

//lint:file-ignore ST1001 Dot imports by jet

import (
	"log"
	"slices"
	"strings"

	"github.com/bwmarrin/discordgo"
	. "github.com/go-jet/jet/v2/postgres"
	"github.com/jedib0t/go-pretty/v6/table"
	. "github.com/slazurin/maple-culvert-tracker/.gen/mapleculverttrackerdb/public/table"
	"github.com/slazurin/maple-culvert-tracker/internal/apiredis"
	"github.com/slazurin/maple-culvert-tracker/internal/commands/helpers"
	cmdhelpers "github.com/slazurin/maple-culvert-tracker/internal/commands/helpers"
	"github.com/slazurin/maple-culvert-tracker/internal/data"
	"github.com/slazurin/maple-culvert-tracker/internal/db"
)

func personalBests(s *discordgo.Session, i *discordgo.InteractionCreate) {
	chars, err := helpers.GetActiveCharacters(apiredis.RedisDB, db.DB)
	if err != nil {
		log.Println("personalBests: get active chars failed", err)
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "personal-bests command failed, could not get active characters",
			},
		})
		return
	}

	if len(*chars) == 0 {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "No active characters found.",
			},
		})
		return
	}

	charIDsInExpr := []Expression{}
	for _, v := range *chars {
		charIDsInExpr = append(charIDsInExpr, Int64(v.ID))
	}

	stmt := SELECT(
		Characters.MapleCharacterName.AS("maple_character_name"),
		MAX(CharacterCulvertScores.Score).AS("personal_best"),
	).FROM(
		CharacterCulvertScores.INNER_JOIN(Characters, Characters.ID.EQ(CharacterCulvertScores.CharacterID)),
	).WHERE(
		CharacterCulvertScores.CharacterID.IN(charIDsInExpr...).AND(
			CharacterCulvertScores.CulvertDate.GT_EQ(DateT(data.Date2mPatch)),
		),
	).GROUP_BY(Characters.MapleCharacterName)

	dest := []struct {
		MapleCharacterName string
		PersonalBest       int32
		pos                int
	}{}

	err = stmt.Query(db.DB, &dest)
	if err != nil {
		log.Println("personalBests: select stmt failed", err)
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "personal-bests command failed getting data from db. See server logs.",
			},
		})
		return
	}

	slices.SortStableFunc(dest, func(a, b struct {
		MapleCharacterName string
		PersonalBest       int32
		pos                int
	}) int {
		if a.PersonalBest != b.PersonalBest {
			return int(b.PersonalBest) - int(a.PersonalBest)
		}
		return strings.Compare(a.MapleCharacterName, b.MapleCharacterName)
	})

	for idx := range dest {
		dest[idx].pos = idx + 1
	}

	columnCount := 1
	if len(dest) > 65 {
		columnCount = 2
	}
	if len(dest) > 130 {
		columnCount = 3
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Personal bests for all active characters",
			Files: []*discordgo.File{{Name: "message.txt", Reader: strings.NewReader(cmdhelpers.FormatNthColumnList(columnCount, dest, table.Row{"Pos", "Character", "Personal Best"}, func(row struct {
				MapleCharacterName string
				PersonalBest       int32
				pos                int
			}, idx int) table.Row {
				return table.Row{row.pos, row.MapleCharacterName, row.PersonalBest}
			}))}},
		},
	})
}
