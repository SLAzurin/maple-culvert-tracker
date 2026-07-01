package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/slazurin/maple-culvert-tracker/internal/apiredis"
	"github.com/slazurin/maple-culvert-tracker/internal/commands/helpers"
	"github.com/slazurin/maple-culvert-tracker/internal/db"
)

func parseImages(s *discordgo.Session, i *discordgo.InteractionCreate) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})

	content := new(string)
	editContent := func(msg string) {
		*content = msg
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: content})
	}

	channelID := ""
	messageIDs := []string{}
	for _, v := range i.ApplicationCommandData().Options {
		if v.Name == "channel-id" {
			channelID = parseChannelID(v.StringValue())
		}
		if strings.HasPrefix(v.Name, "message-id") {
			id := strings.TrimSpace(v.StringValue())
			if id != "" {
				messageIDs = append(messageIDs, id)
			}
		}
	}
	if channelID == "" {
		editContent("No channel ID provided.")
		return
	}
	if len(messageIDs) == 0 {
		editContent("No message IDs provided.")
		return
	}

	// Collect image attachments from the referenced messages (kept in memory only).
	imageURLs := []string{}
	for _, msgID := range messageIDs {
		msg, err := s.ChannelMessage(channelID, msgID)
		if err != nil {
			log.Println("parseImages: failed to fetch message", msgID, "in channel", channelID, err)
			editContent("Failed to fetch message `" + msgID + "`. Make sure the message and channel IDs are correct.")
			return
		}
		if len(msg.Attachments) > 0 {
			log.Printf("parseImages: message %s has %d attachment(s)", msgID, len(msg.Attachments))
			for _, a := range msg.Attachments {
				log.Printf("  - %s (ContentType=%q, Size=%d bytes)", a.Filename, a.ContentType, a.Size)
			}
		}
		for _, a := range msg.Attachments {
			if isImageAttachment(a) {
				imageURLs = append(imageURLs, a.URL)
			}
		}
	}
	if len(imageURLs) == 0 {
		editContent("No image attachments found on the provided message(s).")
		return
	}

	font, err := helpers.LoadGPQFont()
	if err != nil {
		log.Println("parseImages: failed to load font templates:", err)
		editContent("Internal error loading font templates. Please try again later.")
		return
	}

	characters, err := helpers.GetActiveCharacters(apiredis.RedisDB, db.DB)
	if err != nil {
		log.Println("parseImages: failed to query active characters:", err)
		editContent("Internal error querying active characters. Please try again later.")
		return
	}
	memberNames := make([]string, 0, len(*characters))
	activeSet := make(map[string]bool, len(*characters))
	for _, c := range *characters {
		memberNames = append(memberNames, c.MapleCharacterName)
		activeSet[c.MapleCharacterName] = true
	}

	// Process each image in parallel: download into memory then parse.
	type imgResult struct {
		scores map[string]int
		err    error
	}
	results := make([]imgResult, len(imageURLs))
	var wg sync.WaitGroup
	for idx, url := range imageURLs {
		wg.Add(1)
		go func(idx int, url string) {
			defer wg.Done()
			data, err := downloadBytes(url)
			if err != nil {
				results[idx] = imgResult{err: err}
				return
			}
			scores, err := helpers.ParseSmallImage(data, memberNames, font)
			results[idx] = imgResult{scores: scores, err: err}
		}(idx, url)
	}
	wg.Wait()

	merged := map[string]int{}
	for idx, r := range results {
		if r.err != nil {
			log.Println("parseImages: failed to process image", imageURLs[idx], r.err)
			editContent("Failed to process one of the images. Please ensure they are valid `small` style GPQ score images.")
			return
		}
		for name, score := range r.scores {
			merged[name] = score
		}
	}

	// Fail when any parsed name is not an active character.
	unmatched := []string{}
	for name := range merged {
		if !activeSet[name] {
			unmatched = append(unmatched, name)
		}
	}
	if len(unmatched) > 0 {
		sort.Strings(unmatched)
		unmatchedJSON, _ := json.MarshalIndent(unmatched, "", "    ")
		*content = "Some parsed character names did not match any active character. Fix the images or track these characters first."
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: content,
			Files: []*discordgo.File{{
				Name:        "unmatched.json",
				ContentType: "application/json",
				Reader:      strings.NewReader(string(unmatchedJSON)),
			}},
		})
		return
	}

	out, err := json.MarshalIndent(merged, "", "    ")
	if err != nil {
		log.Println("parseImages: failed to marshal result:", err)
		editContent("Internal error building the JSON result.")
		return
	}

	*content = "Parsed scores from " + strconv.Itoa(len(imageURLs)) + " image(s)."
	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: content,
		Files: []*discordgo.File{{
			Name:        "gpq_scores.json",
			ContentType: "application/json",
			Reader:      strings.NewReader(string(out)),
		}},
	})
}

// parseChannelID accepts either a raw ID ("123") or a channel mention
// ("<#123>") and returns the bare ID.
func parseChannelID(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "<#")
	s = strings.TrimSuffix(s, ">")
	return strings.TrimSpace(s)
}

func isImageAttachment(a *discordgo.MessageAttachment) bool {
	if strings.HasPrefix(a.ContentType, "image/") {
		return true
	}
	fn := strings.ToLower(a.Filename)
	// Accept common image extensions, even if ContentType is empty
	return strings.HasSuffix(fn, ".png") ||
		strings.HasSuffix(fn, ".jpg") ||
		strings.HasSuffix(fn, ".jpeg") ||
		strings.HasSuffix(fn, ".gif") ||
		strings.HasSuffix(fn, ".webp")
}

func downloadBytes(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download %s: status %d", url, resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}
