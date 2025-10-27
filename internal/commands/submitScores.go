package commands

//lint:file-ignore ST1001 Dot imports by jet

import (
	"encoding/json"
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
	"github.com/slazurin/maple-culvert-tracker/internal/commands/helpers"
	"github.com/slazurin/maple-culvert-tracker/internal/data"
	"github.com/slazurin/maple-culvert-tracker/internal/db"
)

func submitScores(s *discordgo.Session, i *discordgo.InteractionCreate) {
	var err error
	options := i.ApplicationCommandData().Options

	isDefaultCulvertWeek := true
	culvertDate := helpers.GetCulvertResetDate(time.Now())
	culvertDateStr := culvertDate.Format("2006-01-02")
	overwriteExisting := false
	attachmentMap := map[string]int{}

	for _, v := range options {
		if v.Name == "culvert-date" {
			isDefaultCulvertWeek = false
			culvertDateStr = strings.Trim(v.StringValue(), " ")
			culvertDate, err = time.Parse(culvertDateStr, "2006-01-02")
			if err != nil {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "Invalid date format provided! Please use YYYY-MM-DD.",
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				})
				return
			}

			if culvertDate.Weekday() != helpers.GetCulvertResetDay(time.Now()) {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "The provided date is not a Wednesday! Culvert resets occur on Wednesdays.",
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				})
				return
			}
		}
		if v.Name == "scores-attachment" {
			attachmentID := v.Value.(string)
			attachment, ok := i.ApplicationCommandData().Resolved.Attachments[attachmentID]
			if !ok {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "Failed to get attachment details! Please try again.",
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				})
				return
			}
			// log.Printf("Received attachment: %s (URL: %s, Size: %d bytes)\n", attachment.Filename, attachment.URL, attachment.Size)
			if attachment.Size > 2048*1024 {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "Attachment size exceeds 2MB limit! Please upload a smaller file.",
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				})
				return
			}

			if !strings.HasSuffix(attachment.Filename, ".txt") && !strings.HasSuffix(attachment.Filename, ".json") {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "Invalid attachment format! Please upload a .txt or .json file.",
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				})
				return
			}

			resp, err := http.Get(attachment.URL)
			if err != nil || resp.StatusCode != http.StatusOK {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "Failed to download your attachment! Please try again.",
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				})
				return
			}

			content, err := io.ReadAll(resp.Body)
			if err != nil {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "Error reading attachment content! Please try again.",
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				})
				return
			}

			defer resp.Body.Close()

			err = json.Unmarshal(content, &attachmentMap)
			if err != nil {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "Failed to parse attachment content! Please ensure it's valid JSON format of { \"character-name\": 123, \"character-name-2\": 456 }.",
						Flags:   discordgo.MessageFlagsEphemeral,
					},
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
	stmt := SELECT(Characters.ID.AS("id"), Characters.MapleCharacterName.AS("maple_character_name"), CharacterCulvertScores.Score.AS("score")).FROM(Characters.LEFT_JOIN(CharacterCulvertScores, Characters.ID.EQ(CharacterCulvertScores.CharacterID))).WHERE(CharacterCulvertScores.CulvertDate.EQ(DateT(culvertDate)).AND(Characters.DiscordUserID.NOT_EQ(String(data.INTERNAL_DISCORD_ID_UNTRACKED))))

	trackedCharacterScores := []struct {
		ID                 int
		MapleCharacterName string
		Score              *int
	}{}

	err = stmt.Query(db.DB, &trackedCharacterScores)
	if err != nil {
		log.Println("submitScores: Error querying tracked character scores:", err)
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Internal server error while querying existing scores. Please try again later.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	log.Println(overwriteExisting, isDefaultCulvertWeek)
	log.Println(attachmentMap)
	log.Println(trackedCharacterScores)

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
			if v.Score != nil && attachmentMap[v.MapleCharacterName] > 0 && !overwriteExisting {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "Existing scores found, Set the 'overwrite-existing' option to `True` to overwrite them. No changes were made.",
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				})
				return
			}

			// character found in attachment map, score is nil = isNew: true, add to newMapIsNew
			if v.Score == nil {
				newMapIsNew.Payload = append(newMapIsNew.Payload, struct {
					CharacterID int64 `json:"character_id"`
					Score       int   `json:"score"`
				}{
					CharacterID: int64(v.ID),
					Score:       attachmentMap[v.MapleCharacterName],
				})
			} else {
				newMapIsNotNew.Payload = append(newMapIsNotNew.Payload, struct {
					CharacterID int64 `json:"character_id"`
					Score       int   `json:"score"`
				}{
					CharacterID: int64(v.ID),
					Score:       attachmentMap[v.MapleCharacterName],
				})
			}

			// done processing this character, delete from attachmentMap
			delete(attachmentMap, v.MapleCharacterName)
			// the remaining entries in attachmentMap are untracked characters, which are missing in database, we need to send s.InteractionRespond and return early later
		} else {
			// character exists in database, character not in attachment, meaning score this week is zero and isNew
			newMapIsNew.Payload = append(newMapIsNew.Payload, struct {
				CharacterID int64 `json:"character_id"`
				Score       int   `json:"score"`
			}{
				CharacterID: int64(v.ID),
				Score:       0,
			})

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
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Failed to submit scores. Correct their name or track these new characters before submitting their scores.",
				Flags:   discordgo.MessageFlagsEphemeral,
				Files: []*discordgo.File{{
					Name:        "names.json",
					ContentType: "application/json",
					Reader:      strings.NewReader(string(untrackedNamesJSON)),
				}},
			},
		})
		return
	}

	// done sorting isNew and isNotNew maps

	// generate api auth token for internal api call
	apiAuth := apihelpers.GenerateAPIAuthToken("discordSubmitScores", time.Now().Add(5*time.Minute))
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

	if newMapIsNewSuccess && newMapIsNotNewSuccess {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Scores submitted successfully for culvert week of " + culvertDateStr + ".",
			},
		})
		return
	}

	// not normal past this point
	log.Println("submitScores: One of the score submissions failed. newMapIsNewSuccess", newMapIsNewSuccess, "newMapIsNotNewSuccess", newMapIsNotNewSuccess)
	log.Println("submitScores: len(newMapIsNew.Payload)", len(newMapIsNew.Payload), "len(newMapIsNotNew.Payload)", len(newMapIsNotNew.Payload))

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Scores submission failed. See server logs for details.",
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}
