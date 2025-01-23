package commands

//lint:file-ignore ST1001 Dot imports by jet
import (
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
	. "github.com/go-jet/jet/v2/postgres"
	. "github.com/slazurin/maple-culvert-tracker/.gen/mapleculverttrackerdb/public/table"
	cmdhelpers "github.com/slazurin/maple-culvert-tracker/internal/commands/helpers"
	"github.com/slazurin/maple-culvert-tracker/internal/db"
)

func culvertMegaDetails(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options

	weeks := int64(8)
	date := ""

	for _, v := range options {
		if v.Name == "date" {
			date = v.StringValue()
		}
		if v.Name == "weeks" {
			weeks = v.IntValue()
		}
	}

	if date == "" {
		date = cmdhelpers.GetCulvertResetDate(time.Now()).Format("2006-01-02")
	}

	d, err := time.Parse("2006-01-02", date) // YYYY-MM-DD
	if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Invalid date format, should be YYYY-MM-DD",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	// get all dates necessary
	dates := []time.Time{}
	for i := 0; i < int(weeks); i++ {
		dates = append(dates, cmdhelpers.GetCulvertResetDate(d.AddDate(0, 0, -i*7)))
	}

	allDates := []Expression{}
	for _, v := range dates {
		allDates = append(allDates, DateT(v))
	}

	// get all rows for past x week from weeks value
	stmt := SELECT(Characters.MapleCharacterName.AS("maple_character_name"), CharacterCulvertScores.Score.AS("score"), CharacterCulvertScores.CulvertDate.AS("culvert_date")).FROM(CharacterCulvertScores.INNER_JOIN(Characters, Characters.ID.EQ(CharacterCulvertScores.CharacterID))).WHERE(CharacterCulvertScores.CulvertDate.IN(allDates...).AND(Characters.DiscordUserID.NOT_EQ(String("1")))).ORDER_BY(CharacterCulvertScores.CulvertDate.DESC(), CharacterCulvertScores.Score.DESC())

	dest := []struct {
		CulvertDate        time.Time
		Score              int32
		MapleCharacterName string
	}{}
	err = stmt.Query(db.DB, &dest)
	if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Failed to find all characters' dataset!",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	// log dest
	for _, v := range dest {
		log.Println(v)
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Command under construction, completed!",
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}
