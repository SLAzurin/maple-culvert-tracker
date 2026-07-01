package commands

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/slazurin/maple-culvert-tracker/internal/commands/helpers"
)

func submitScoresFromAttachment(s *discordgo.Session, i *discordgo.InteractionCreate) {
	var err error

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})

	options := i.ApplicationCommandData().Options

	culvertDate := helpers.GetCulvertResetDate(time.Now())
	culvertDateStr := culvertDate.Format("2006-01-02")
	overwriteExisting := false
	channelID := ""
	messageID := ""
	attachmentMap := map[string]int{}

	content := new(string)

	for _, v := range options {
		if v.Name == "channel-id" {
			channelID = parseChannelID(v.StringValue())
		}
		if v.Name == "message-id" {
			messageID = strings.TrimSpace(v.StringValue())
		}
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
		if v.Name == "overwrite-existing" {
			overwriteExisting = v.BoolValue()
		}
	}

	if channelID == "" {
		*content = "No channel ID provided."
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: content})
		return
	}
	if messageID == "" {
		*content = "No message ID provided."
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: content})
		return
	}

	msg, err := s.ChannelMessage(channelID, messageID)
	if err != nil {
		log.Println("submitScoresFromAttachment: failed to fetch message", messageID, "in channel", channelID, err)
		*content = "Failed to fetch message `" + messageID + "`. Make sure the message and channel IDs are correct."
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: content})
		return
	}

	// find the first .txt or .json attachment on the message
	var attachment *discordgo.MessageAttachment
	for _, a := range msg.Attachments {
		if strings.HasSuffix(a.Filename, ".txt") || strings.HasSuffix(a.Filename, ".json") {
			attachment = a
			break
		}
	}
	if attachment == nil {
		*content = "No .txt or .json attachment found on the provided message."
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: content})
		return
	}

	if attachment.Size > 2048*1024 {
		*content = "Attachment size exceeds 2MB limit! Please upload a smaller file."
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: content})
		return
	}

	resp, err := http.Get(attachment.URL)
	if err != nil || resp.StatusCode != http.StatusOK {
		*content = "Failed to download the attachment! Please try again."
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: content})
		return
	}

	bodyContent, err := io.ReadAll(resp.Body)
	if err != nil {
		*content = "Error reading attachment content! Please try again."
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: content})
		return
	}
	defer resp.Body.Close()

	err = json.Unmarshal(bodyContent, &attachmentMap)
	if err != nil {
		*content = "Failed to parse attachment content! Please ensure it's valid JSON format of { \"character-name\": 123, \"character-name-2\": 456 }."
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: content})
		return
	}

	// done processing attachment

	finalizeSubmitScores(s, i, content, attachmentMap, culvertDate, culvertDateStr, overwriteExisting)
}
