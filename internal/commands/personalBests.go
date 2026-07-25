package commands

import (
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/slazurin/maple-culvert-tracker/internal/apiredis"
	"github.com/slazurin/maple-culvert-tracker/internal/commands/helpers"
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

	charIDs := make([]int64, 0, len(*chars))
	for _, v := range *chars {
		charIDs = append(charIDs, v.ID)
	}

	scores, err := helpers.FetchCharacterScoresSince(db.DB, charIDs, helpers.PersonalBestRankWindowStart())
	if err != nil {
		log.Println("personalBests: fetch scores failed", err)
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "personal-bests command failed getting data from db. See server logs.",
			},
		})
		return
	}

	dest := helpers.ComputePersonalBestRankMetrics(scores, helpers.PersonalBestRankWeeks)
	if len(dest) == 0 {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "No personal best scores found.",
			},
		})
		return
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Personal bests for all active characters",
			Files: []*discordgo.File{{Name: "message.txt", Reader: strings.NewReader(helpers.FormatNthColumnList(1, dest, table.Row{"Pos", "Character", "Personal Best"}, func(row helpers.PersonalBestRankMetric, idx int) table.Row {
				return table.Row{
					helpers.FormatPersonalBestPos(row.Pos, row.DeltaPos, row.HasPrevRank, row.Streak),
					row.MapleCharacterName,
					row.PersonalBest,
				}
			}))}},
		},
	})
}
