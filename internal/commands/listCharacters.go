package commands

//lint:file-ignore ST1001 Dot imports by jet

import (
	"encoding/json"
	"errors"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
	. "github.com/go-jet/jet/v2/postgres"
	"github.com/go-jet/jet/v2/qrm"
	. "github.com/slazurin/maple-culvert-tracker/.gen/mapleculverttrackerdb/public/table"
	"github.com/slazurin/maple-culvert-tracker/internal/db"
)

func listCharacters(s *discordgo.Session, i *discordgo.InteractionCreate) {
	characters := []string{}
	stmt := SELECT(Characters.MapleCharacterName).FROM(Characters).WHERE(Characters.DiscordUserID.NOT_EQ(String("1"))).ORDER_BY(Characters.MapleCharacterName.ASC())

	err := stmt.Query(db.DB, &characters)
	if err != nil && !errors.Is(err, qrm.ErrNoRows) {
		log.Println("listCharacters: failed to query characters", err)
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Failed to fetch characters from database!",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	jsonCharacters, _ := json.Marshal(characters)
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Here you go! Copy the contents of the attached file and paste it to the OCR app.",
			Files:   []*discordgo.File{{Name: "characters.json", Reader: strings.NewReader(string(jsonCharacters))}},
		},
	})
}
