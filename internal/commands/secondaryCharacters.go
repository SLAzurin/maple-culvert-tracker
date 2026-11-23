package commands

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/lib/pq"
	"github.com/slazurin/maple-culvert-tracker/.gen/mapleculverttrackerdb/public/model"
	apihelpers "github.com/slazurin/maple-culvert-tracker/internal/api/helpers"
	"github.com/slazurin/maple-culvert-tracker/internal/apiredis"
	"github.com/slazurin/maple-culvert-tracker/internal/commands/helpers"
	"github.com/slazurin/maple-culvert-tracker/internal/db"
)

func editSecondaryCharacter(add bool) func(*discordgo.Session, *discordgo.InteractionCreate) {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		}); err != nil {
			log.Println("editSecondaryCharacter: failed to acknowledge interaction:", err)
			return
		}

		var parent, secondary string
		var skipNameCheck bool
		for _, option := range i.ApplicationCommandData().Options {
			switch option.Name {
			case "character-name":
				parent = strings.TrimSpace(option.StringValue())
			case "secondary-character-name":
				secondary = strings.TrimSpace(option.StringValue())
			case "skip-name-check":
				skipNameCheck = option.BoolValue()
			}
		}
		var message string
		var err error
		switch {
		case secondary == "" || (add && parent == ""):
			message = "Enter a name for each required character."
		case add && skipNameCheck:
			message, err = addSecondaryMapping(db.DB, parent, secondary)
		case add:
			region := apiredis.OPTIONAL_CONF_MAPLE_REGION.GetWithDefault(apiredis.RedisDB, "na")
			message, err = addRankedSecondaryMapping(db.DB, parent, secondary, region)
		default:
			message, err = removeSecondaryMapping(db.DB, secondary)
		}
		if err != nil {
			log.Println("editSecondaryCharacter:", err)
			message = "Could not update secondary characters. Please try again later."
		}
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content:         &message,
			AllowedMentions: &discordgo.MessageAllowedMentions{},
		})
	}
}

func addRankedSecondaryMapping(database *sql.DB, parentName, secondaryName, region string) (string, error) {
	character, err := apihelpers.FetchCharacterData(strings.ToLower(secondaryName), region)
	if err != nil {
		log.Println("addRankedSecondaryMapping: rankings lookup failed:", err)
		return fmt.Sprintf("Could not verify %q in the official MapleStory rankings. Check the name and try again.", secondaryName), nil
	}
	if !strings.EqualFold(character.CharacterName, secondaryName) {
		return fmt.Sprintf("No exact rankings match for %q. MapleStory returned %q.", secondaryName, character.CharacterName), nil
	}
	return addSecondaryMapping(database, parentName, character.CharacterName)
}

func addSecondaryMapping(database *sql.DB, parentName, secondaryName string) (string, error) {
	tx, err := database.Begin()
	if err != nil {
		return "", err
	}
	defer tx.Rollback()
	// Match and insert against one roster. Serialize secondary writers so two
	// concurrent commands cannot add OCR-equivalent names to different parents.
	if _, err := tx.Exec(`LOCK TABLE characters IN SHARE MODE; LOCK TABLE secondary_characters IN SHARE ROW EXCLUSIVE MODE`); err != nil {
		return "", err
	}

	rows, err := tx.Query(`
		SELECT id, maple_character_name, true, lower(maple_character_name) = lower($1), lower(maple_character_name) = lower($2)
		FROM characters
		UNION ALL
		SELECT character_id, maple_character_name, false, false, lower(maple_character_name) = lower($2)
		FROM secondary_characters
		ORDER BY 3 DESC, 1, 2`, parentName, secondaryName)
	if err != nil {
		return "", err
	}
	defer rows.Close()
	var parentID int64
	var parent string
	var conflict string
	var existing bool
	secondaryOCRName := helpers.NormalizeGPQName(secondaryName)
	for rows.Next() {
		var id int64
		var name string
		var primary, isParent, sameName bool
		if err := rows.Scan(&id, &name, &primary, &isParent, &sameName); err != nil {
			return "", err
		}
		if isParent {
			if parent != "" {
				return fmt.Sprintf("More than one tracked character matches %q.", parentName), nil
			}
			parent, parentID = name, id
		}
		if sameName {
			switch {
			case primary:
				conflict = fmt.Sprintf("%q is already tracked as a main character.", name)
			case id != parentID:
				conflict = fmt.Sprintf("%q is already linked to another tracked character.", name)
			default:
				existing = true
			}
		} else if id != parentID && helpers.NormalizeGPQName(name) == secondaryOCRName && conflict == "" {
			conflict = fmt.Sprintf("OCR cannot distinguish %q from %q, which belongs to another tracked character.", secondaryName, name)
		}
	}
	if err := rows.Err(); err != nil {
		return "", err
	}
	if parent == "" {
		return fmt.Sprintf("Tracked character %q was not found.", parentName), nil
	}
	if conflict != "" {
		return conflict, nil
	}
	if existing {
		return fmt.Sprintf("%q is already linked to %q.", secondaryName, parent), nil
	}
	if _, err := tx.Exec(`INSERT INTO secondary_characters (character_id, maple_character_name) VALUES ($1, $2)`, parentID, secondaryName); err != nil {
		return "", err
	}
	if err := tx.Commit(); err != nil {
		return "", err
	}
	return fmt.Sprintf("Linked %q to %q.", secondaryName, parent), nil
}

func removeSecondaryMapping(database *sql.DB, secondaryName string) (string, error) {
	var parent, secondary string
	err := database.QueryRow(`
		DELETE FROM secondary_characters AS s USING characters AS c
		WHERE s.character_id = c.id AND lower(s.maple_character_name) = lower($1)
		RETURNING c.maple_character_name, s.maple_character_name`, secondaryName).Scan(&parent, &secondary)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Sprintf("Secondary character %q was not found.", secondaryName), nil
	}
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Removed %q from %q.", secondary, parent), nil
}

// loadActiveSecondaries uses the same IDs and names as the caller's active roster.
func loadActiveSecondaries(database *sql.DB, active []model.Characters) (map[string][]string, error) {
	secondaries := make(map[string][]string)
	if len(active) == 0 {
		return secondaries, nil
	}
	ids := make([]int64, 0, len(active))
	parents := make(map[int64]string, len(active))
	for _, character := range active {
		ids = append(ids, character.ID)
		parents[character.ID] = character.MapleCharacterName
	}
	rows, err := database.Query(`
		SELECT character_id, maple_character_name FROM secondary_characters
		WHERE character_id = ANY($1) ORDER BY maple_character_name COLLATE "C"`, pq.Array(ids))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id int64
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			return nil, err
		}
		parent := parents[id]
		secondaries[parent] = append(secondaries[parent], name)
	}
	return secondaries, rows.Err()
}

func listSecondaryCharacters(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	}); err != nil {
		log.Println("listSecondaryCharacters: failed to acknowledge interaction:", err)
		return
	}
	var out []byte
	characters, err := helpers.GetActiveCharacters(apiredis.RedisDB, db.DB)
	if err == nil {
		var secondaries map[string][]string
		secondaries, err = loadActiveSecondaries(db.DB, *characters)
		if err == nil {
			out, err = json.MarshalIndent(secondaries, "", "  ")
		}
	}
	message := "Secondary characters for the active roster."
	response := &discordgo.WebhookEdit{Content: &message}
	if err != nil {
		log.Println("listSecondaryCharacters:", err)
		message = "Could not load secondary characters. Please try again later."
	} else {
		response.Files = []*discordgo.File{{Name: "secondary_characters.json", ContentType: "application/json", Reader: bytes.NewReader(out)}}
	}
	s.InteractionResponseEdit(i.Interaction, response)
}
