package commands

//lint:file-ignore ST1001 Dot imports by jet

import (
	"database/sql"
	"encoding/json/v2"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	. "github.com/go-jet/jet/v2/postgres"
	. "github.com/slazurin/maple-culvert-tracker/.gen/mapleculverttrackerdb/public/table"
	apihelpers "github.com/slazurin/maple-culvert-tracker/internal/api/helpers"
	"github.com/slazurin/maple-culvert-tracker/internal/apiredis"
	"github.com/slazurin/maple-culvert-tracker/internal/commands/helpers"
	"github.com/slazurin/maple-culvert-tracker/internal/data"
	"github.com/slazurin/maple-culvert-tracker/internal/db"
)

func submitScores(s *discordgo.Session, i *discordgo.InteractionCreate) {
	var err error

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})

	options := i.ApplicationCommandData().Options

	culvertDate := helpers.GetCulvertResetDate(time.Now())
	culvertDateStr := culvertDate.Format("2006-01-02")
	overwriteExisting := false
	attachmentMap := map[string]int{}

	content := new(string)

	for _, v := range options {
		if v.Name == "date" {
			culvertDateStr = strings.Trim(v.StringValue(), " ")
			culvertDate, err = time.Parse("2006-01-02", culvertDateStr)
			if err != nil {
				*content = "Invalid date format provided! Please use YYYY-MM-DD."
				s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
					Content: content,
				})
				return
			}

			if culvertDate.Weekday() != helpers.GetCulvertResetDay(time.Now()) {
				*content = "The provided date is not a Wednesday! Culvert resets occur on Wednesdays."
				s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
					Content: content,
				})
				return
			}
		}
		if v.Name == "scores-attachment" {
			attachmentID := v.Value.(string)
			attachment, ok := i.ApplicationCommandData().Resolved.Attachments[attachmentID]
			if !ok {
				*content = "Failed to get attachment details! Please try again."
				s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
					Content: content,
				})
				return
			}
			// log.Printf("Received attachment: %s (URL: %s, Size: %d bytes)\n", attachment.Filename, attachment.URL, attachment.Size)
			if attachment.Size > 2048*1024 {
				*content = "Attachment size exceeds 2MB limit! Please upload a smaller file."
				s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
					Content: content,
				})
				return
			}

			if !strings.HasSuffix(attachment.Filename, ".txt") && !strings.HasSuffix(attachment.Filename, ".json") {
				*content = "Invalid attachment format! Please upload a .txt or .json file."
				s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
					Content: content,
				})
				return
			}

			resp, err := http.Get(attachment.URL)
			if err != nil || resp.StatusCode != http.StatusOK {
				*content = "Failed to download your attachment! Please try again."
				s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
					Content: content,
				})
				return
			}

			bodyContent, err := io.ReadAll(resp.Body)
			if err != nil {
				*content = "Error reading attachment content! Please try again."
				s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
					Content: content,
				})
				return
			}

			defer resp.Body.Close()

			err = json.Unmarshal(bodyContent, &attachmentMap)
			if err != nil {
				*content = "Failed to parse attachment content! Please ensure it's valid JSON format of { \"character-name\": 123, \"character-name-2\": 456 }."
				s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
					Content: content,
				})
				return
			}

			// done processing attachment
		}
		if v.Name == "overwrite-existing" {
			overwriteExisting = v.BoolValue()
		}
	}

	// done parsing options

	// query all scores for culvert date to see if any exist
	// get all active tracked characters
	characters, err := helpers.GetActiveCharacters(apiredis.RedisDB, db.DB)
	if err != nil {
		log.Println("submitScores: Error querying active characters:", err)
		*content = "Internal server error while querying active characters. Please try again later."
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: content,
		})
		return
	}

	characterIDs := []Expression{}
	for _, v := range *characters {
		characterIDs = append(characterIDs, Int64(v.ID))
	}

	stmt := SELECT(Characters.ID.AS("id"), Characters.MapleCharacterName.AS("maple_character_name"), CharacterCulvertScores.Score.AS("score")).FROM(Characters.LEFT_JOIN(CharacterCulvertScores, Characters.ID.EQ(CharacterCulvertScores.CharacterID).AND(CharacterCulvertScores.CulvertDate.EQ(DateT(culvertDate))))).WHERE(Characters.ID.IN(characterIDs...))

	trackedCharacterScores := []struct {
		ID                 int
		MapleCharacterName string
		Score              sql.NullInt64
	}{}

	err = stmt.Query(db.DB, &trackedCharacterScores)
	if err != nil {
		log.Println("submitScores: Error querying tracked character scores:", err)
		*content = "Internal server error while querying existing scores. Please try again later."
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: content,
		})
		return
	}

	// validate overwriteExisting
	// check if there are any scores

	newMapIsNew := data.POSTCulvertBody{
		IsNew: true,
		Week:  culvertDateStr,
		Payload: []struct {
			CharacterID int64 `json:"character_id"`
			Score       int   `json:"score"`
		}{},
	}
	newMapIsNotNew := data.POSTCulvertBody{
		IsNew: false,
		Week:  culvertDateStr,
		Payload: []struct {
			CharacterID int64 `json:"character_id"`
			Score       int   `json:"score"`
		}{},
	}
	for _, v := range trackedCharacterScores {
		if _, ok := attachmentMap[v.MapleCharacterName]; ok {
			// break if overwriteExisting is not allowed and score exists
			if v.Score.Valid && attachmentMap[v.MapleCharacterName] > 0 && !overwriteExisting {
				*content = "Existing scores found, Set the `overwrite-existing` option to `True` to overwrite them. No changes were made for " + culvertDateStr + "."
				s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
					Content: content,
				})
				return
			}

			// character found in attachment map, score is nil = isNew: true, add to newMapIsNew
			if !v.Score.Valid { // score is null, must insert
				newMapIsNew.Payload = append(newMapIsNew.Payload, struct {
					CharacterID int64 `json:"character_id"`
					Score       int   `json:"score"`
				}{
					CharacterID: int64(v.ID),
					Score:       attachmentMap[v.MapleCharacterName],
				})
			} else {
				if v.Score.Valid && attachmentMap[v.MapleCharacterName] != int(v.Score.Int64) { // score exists, must update if different
					newMapIsNotNew.Payload = append(newMapIsNotNew.Payload, struct {
						CharacterID int64 `json:"character_id"`
						Score       int   `json:"score"`
					}{
						CharacterID: int64(v.ID),
						Score:       attachmentMap[v.MapleCharacterName],
					})
				}
			}

			// done processing this character, delete from attachmentMap
			delete(attachmentMap, v.MapleCharacterName)
			// the remaining entries in attachmentMap are untracked characters, which are missing in database, we need to send s.InteractionResponseEdit and return early later
		} else {
			// character exists in database, character not in attachment, meaning score this week must edit
			if !v.Score.Valid {
				// score is null, must insert
				newMapIsNew.Payload = append(newMapIsNew.Payload, struct {
					CharacterID int64 `json:"character_id"`
					Score       int   `json:"score"`
				}{
					CharacterID: int64(v.ID),
					Score:       0,
				})
			}
			if v.Score.Valid && v.Score.Int64 > 0 {
				// score exists, and not 0, must update to 0
				newMapIsNotNew.Payload = append(newMapIsNotNew.Payload, struct {
					CharacterID int64 `json:"character_id"`
					Score       int   `json:"score"`
				}{
					CharacterID: int64(v.ID),
					Score:       0,
				})
			}
		}
	}
	// done processing all tracked characters

	// check if there are untracked characters in attachmentMap
	if len(attachmentMap) > 0 {
		untrackedNames := []string{}
		for k := range attachmentMap {
			untrackedNames = append(untrackedNames, k)
		}
		untrackedNamesJSON, _ := json.Marshal(untrackedNames)
		*content = "Failed to submit scores. Correct their name or track these new characters before submitting their scores."
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: content,
			Files: []*discordgo.File{{
				Name:        "names.json",
				ContentType: "application/json",
				Reader:      strings.NewReader(string(untrackedNamesJSON)),
			}},
		})
		return
	}

	// done sorting isNew and isNotNew maps

	// generate api auth token for internal api call
	apiAuth := apihelpers.GenerateAPIAuthToken(i.Member.User.Username, i.Member.User.ID, time.Now().Add(5*time.Minute))
	newMapIsNewSuccess := true
	newMapIsNotNewSuccess := true
	port := os.Getenv("BACKEND_HTTP_PORT")
	if port == "" {
		port = "8080"
	}

	client := &http.Client{}

	if len(newMapIsNew.Payload) > 0 {
		body, _ := json.Marshal(newMapIsNew)
		req, _ := http.NewRequest("POST", "http://localhost:"+port+"/api/maple/characters/culvert", strings.NewReader(string(body)))
		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Authorization", "Bearer "+apiAuth)

		resp, err := client.Do(req)
		if err != nil || resp.StatusCode != http.StatusOK {
			newMapIsNewSuccess = false
			log.Println("submitScores: Error submitting new scores:", resp.StatusCode, err)
		}
	}
	if len(newMapIsNotNew.Payload) > 0 {
		body, _ := json.Marshal(newMapIsNotNew)
		req, _ := http.NewRequest("POST", "http://localhost:"+port+"/api/maple/characters/culvert", strings.NewReader(string(body)))
		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Authorization", "Bearer "+apiAuth)

		resp, err := client.Do(req)
		if err != nil || resp.StatusCode != http.StatusOK {
			newMapIsNotNewSuccess = false
			log.Println("submitScores: Error updating existing scores:", resp.StatusCode, err)
		}
	}

	log.Println("submitScores: len(newMapIsNew.Payload)", len(newMapIsNew.Payload), "len(newMapIsNotNew.Payload)", len(newMapIsNotNew.Payload))
	if newMapIsNewSuccess && newMapIsNotNewSuccess {
		time.Sleep(2 * time.Second) // ensures we do not run in the discord throttling
		*content = "Scores submitted successfully for culvert week of " + culvertDateStr + "."
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: content,
		})
		return
	}

	// not normal past this point
	log.Println("submitScores: One or both of the score submissions failed. submit-new-scores", newMapIsNewSuccess, "update-existing-scores", newMapIsNotNewSuccess)

	*content = "Scores submission failed. See server logs for details."
	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: content,
	})
}
