package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/jedib0t/go-pretty/v6/table"
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
	secondaries, err := loadActiveSecondaries(db.DB, *characters)
	if err != nil {
		log.Println("parseImages: failed to query secondary characters:", err)
		editContent("Internal error querying secondary characters. Please try again later.")
		return
	}
	activeNames := make([]string, 0, len(*characters))
	for _, c := range *characters {
		activeNames = append(activeNames, c.MapleCharacterName)
	}
	roster, err := newImageRoster(activeNames, secondaries)
	if err != nil {
		editContent(err.Error())
		return
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
			scores, err := helpers.ParseSmallImageRows(data, font)
			results[idx] = imgResult{scores: scores, err: err}
		}(idx, url)
	}
	wg.Wait()

	var rows []helpers.ScoreEntry
	for idx, r := range results {
		if r.err != nil {
			log.Println("parseImages: failed to process image", imageURLs[idx], r.err)
			editContent("Failed to process one of the images. Please ensure they are valid `small` style GPQ score images.")
			return
		}
		rows = append(rows, r.scores...)
	}
	message, files := imageScoreResponse(roster, rows, len(imageURLs))
	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content:         &message,
		Files:           files,
		AllowedMentions: &discordgo.MessageAllowedMentions{},
	})
}

func imageScoreResponse(roster imageRoster, rows []helpers.ScoreEntry, imageCount int) (string, []*discordgo.File) {
	result := roster.mergeScores(rows)
	var messages []string
	var files []*discordgo.File
	if len(result.unmatched) > 0 {
		messages = append(messages, "Some parsed character names did not match any active character. Fix the images or track these characters first.")
		files = append(files, &discordgo.File{Name: "unmatched.txt", ContentType: "text/plain", Reader: strings.NewReader(formatUnmatchedScoresTable(result.unmatched))})
	}
	if len(result.ambiguous) > 0 {
		messages = append(messages, "Some parsed names match more than one active character. Review the attached rows.")
		files = append(files, &discordgo.File{Name: "ambiguous_scores.txt", ContentType: "text/plain", Reader: strings.NewReader(formatAmbiguousImageScores(result.ambiguous))})
	}
	if len(result.duplicates) > 0 {
		messages = append(messages, "Multiple character names mapped to the same active character. See the attached names and scores.")
		files = append(files, &discordgo.File{Name: "secondary_character_scores.txt", ContentType: "text/plain", Reader: strings.NewReader(formatDuplicateImageScores(result.duplicates))})
	}
	if result.conflict {
		messages = append(messages, "Conflicting scores need review. No score JSON was generated.")
	}
	if result.conflict || len(result.unmatched) > 0 || len(result.ambiguous) > 0 {
		return strings.Join(messages, "\n"), files
	}

	out, err := marshalOrderedScores(result.scores)
	if err != nil {
		log.Println("parseImages: failed to marshal result:", err)
		return "Internal error building the JSON result.", files
	}

	// Non-fatal validation: scores should be in descending order. If not, warn
	// but still attach the output so the user can inspect/correct it.
	msg := "Parsed scores from " + strconv.Itoa(imageCount) + " image(s)."
	scores := result.scores
	if idx := firstDescendingViolation(scores); idx >= 0 {
		msg += fmt.Sprintf("\n:warning: Scores are not in descending order (`%s`: %d, then `%s`: %d). The output may be incorrect; please verify.",
			scores[idx-1].Name, scores[idx-1].Score, scores[idx].Name, scores[idx].Score)
	}
	return strings.Join(append([]string{msg}, messages...), "\n"), append(files, &discordgo.File{
		Name: "gpq_scores.json", ContentType: "application/json", Reader: strings.NewReader(string(out)),
	})
}

func collectUnmatchedScores(entries []helpers.ScoreEntry, activeSet map[string]bool) []helpers.ScoreEntry {
	unmatched := make([]helpers.ScoreEntry, 0)
	for _, entry := range entries {
		if !activeSet[entry.Name] {
			unmatched = append(unmatched, entry)
		}
	}
	return unmatched
}

func formatUnmatchedScoresTable(entries []helpers.ScoreEntry) string {
	t := table.NewWriter()
	t.AppendHeader(table.Row{"Missing Active Character", "Score"})
	for _, entry := range entries {
		t.AppendRow(table.Row{entry.Name, entry.Score})
	}
	return t.Render()
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
