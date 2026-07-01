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
		scores []helpers.ScoreEntry
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

	// Merge preserving order: first attachment to last, top-to-bottom rows.
	// Duplicate names keep their first-seen position; later scores overwrite.
	merged := []helpers.ScoreEntry{}
	mergedPos := map[string]int{}
	for idx, r := range results {
		if r.err != nil {
			log.Println("parseImages: failed to process image", imageURLs[idx], r.err)
			editContent("Failed to process one of the images. Please ensure they are valid `small` style GPQ score images.")
			return
		}
		for _, e := range r.scores {
			if pos, ok := mergedPos[e.Name]; ok {
				merged[pos].Score = e.Score
			} else {
				mergedPos[e.Name] = len(merged)
				merged = append(merged, e)
			}
		}
	}

	// Fail when any parsed name is not an active character.
	unmatched := []string{}
	for _, e := range merged {
		if !activeSet[e.Name] {
			unmatched = append(unmatched, e.Name)
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

	out, err := marshalOrderedScores(merged)
	if err != nil {
		log.Println("parseImages: failed to marshal result:", err)
		editContent("Internal error building the JSON result.")
		return
	}

	// Non-fatal validation: scores should be in descending order. If not, warn
	// but still attach the output so the user can inspect/correct it.
	msg := "Parsed scores from " + strconv.Itoa(len(imageURLs)) + " image(s)."
	if idx := firstDescendingViolation(merged); idx >= 0 {
		msg += "\n:warning: Scores are not in descending order (`" +
			merged[idx-1].Name + "`: " + strconv.Itoa(merged[idx-1].Score) + " -> `" +
			merged[idx].Name + "`: " + strconv.Itoa(merged[idx].Score) +
			"). The output may be incorrect; please verify."
	}
	*content = msg
	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: content,
		Files: []*discordgo.File{{
			Name:        "gpq_scores.json",
			ContentType: "application/json",
			Reader:      strings.NewReader(string(out)),
		}},
	})
}

// firstDescendingViolation returns the index of the first entry whose score is
// greater than the previous entry's score (i.e. breaks descending order), or -1
// if the entries are in non-increasing order.
func firstDescendingViolation(entries []helpers.ScoreEntry) int {
	for i := 1; i < len(entries); i++ {
		if entries[i].Score > entries[i-1].Score {
			return i
		}
	}
	return -1
}

// marshalOrderedScores emits a JSON object of name -> score with 4-space
// indentation, preserving the given entry order (encoding/json sorts map keys,
// so a map cannot be used here).
func marshalOrderedScores(entries []helpers.ScoreEntry) ([]byte, error) {
	if len(entries) == 0 {
		return []byte("{}"), nil
	}
	var b strings.Builder
	b.WriteString("{\n")
	for idx, e := range entries {
		key, err := json.Marshal(e.Name)
		if err != nil {
			return nil, err
		}
		b.WriteString("    ")
		b.Write(key)
		b.WriteString(": ")
		b.WriteString(strconv.Itoa(e.Score))
		if idx < len(entries)-1 {
			b.WriteByte(',')
		}
		b.WriteByte('\n')
	}
	b.WriteByte('}')
	return []byte(b.String()), nil
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
