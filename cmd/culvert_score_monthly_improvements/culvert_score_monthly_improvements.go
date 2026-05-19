package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	_ "github.com/joho/godotenv/autoload"
	"github.com/slazurin/maple-culvert-tracker/internal/apiredis"
	"github.com/slazurin/maple-culvert-tracker/internal/commands/helpers"
	"github.com/slazurin/maple-culvert-tracker/internal/data"
	"github.com/slazurin/maple-culvert-tracker/internal/db"
)

type characterImprovement struct {
	CharacterName  string
	OldScore       int32
	NewScore       int32
	ImprovementPct float64
}

var slackerRoasts = []string{
	"As for the rest of you... let's just pretend this month didn't happen. 😴",
	"Everyone else: the culvert is that way ➡️. Try harder next month.",
	"To those not on this list... I'm not mad, just disappointed. 😔",
	"The rest of the guild chose violence against their own scores this month. 🪦",
	"If you don't see your name above, maybe try actually doing culvert next month? Just a thought. 🤔",
	"The remaining members have been entered into the Witness Protection Program for their own safety. 🫣",
	"Everyone else... we'll talk later. In private. 😤",
	"Shoutout to everyone not listed — at least you're consistent! Consistently mediocre. 🫠",
}

var s *discordgo.Session

func formatScore(score int32) string {
	if score >= 1000000 {
		millions := float64(score) / 1000000.0
		if millions == float64(int(millions)) {
			return fmt.Sprintf("%dm", int(millions))
		}
		return fmt.Sprintf("%.1fm", millions)
	}
	if score >= 1000 {
		thousands := float64(score) / 1000.0
		if thousands == float64(int(thousands)) {
			return fmt.Sprintf("%dk", int(thousands))
		}
		return fmt.Sprintf("%.1fk", thousands)
	}
	return strconv.Itoa(int(score))
}

func buildEmbed(improvers []characterImprovement, hasSlackers bool, fluffText string, prevMonthStr string, currentDateStr string, minImprovementPct float64) *discordgo.MessageEmbed {
	const maxDescLen = 3300 // Discord client visually truncates at ~3300 rendered runes (API limit is 4096)
	const maxTitleLen = 256
	const maxFooterLen = 2048

	// Build the suffix (slacker roast) first so we know how much space to reserve
	suffix := ""
	if hasSlackers {
		suffix = "\n───────────────────\n\n" + fluffText + "\n"
	}
	suffixRuneLen := len([]rune(suffix))

	var descBuilder strings.Builder
	runeCount := 0

	if len(improvers) > 0 {
		header := "**🏆 Top Improvers**\n\n"
		descBuilder.WriteString(header)
		runeCount += len([]rune(header))
		for i, imp := range improvers {
			medal := ""
			switch i {
			case 0:
				medal = "🥇"
			case 1:
				medal = "🥈"
			case 2:
				medal = "🥉"
			default:
				medal = "⭐"
			}
			line := fmt.Sprintf(
				"%s **%s** improved by **%.0f%%**! %s → %s\n",
				medal, imp.CharacterName, imp.ImprovementPct, formatScore(imp.OldScore), formatScore(imp.NewScore),
			)
			lineRuneLen := len([]rune(line))

			// Check if adding this line + suffix would exceed the limit
			if runeCount+lineRuneLen+suffixRuneLen > maxDescLen {
				break
			}
			descBuilder.WriteString(line)
			runeCount += lineRuneLen
		}
	} else {
		noOne := "No one improved enough this month... 💀\n"
		descBuilder.WriteString(noOne)
		runeCount += len([]rune(noOne))
	}

	descBuilder.WriteString(suffix)

	desc := descBuilder.String()
	title := "📈 Culvert Improvements Over the Last Month"
	footerText := fmt.Sprintf("Period: %s → %s | Minimum threshold: %.0f%%", prevMonthStr, currentDateStr, minImprovementPct)

	return &discordgo.MessageEmbed{
		Title:       title,
		Description: desc,
		Color:       3066993, // #2ECC71 green
		Footer: &discordgo.MessageEmbedFooter{
			Text: footerText,
		},
	}
}

func main() {
	// Parse CLI flags
	dateOverride := flag.String("date", "", "Override date (YYYY-MM-DD) to run for a specific month. Skips the month-boundary guard.")
	flag.Parse()

	// Read minimum improvement percentage from Redis (editable from frontend), default 10%
	minImprovementPct := 10.0
	thresholdStr := apiredis.OPTIONAL_CONF_MONTHLY_IMPROVEMENT_THRESHOLD.GetWithDefault(apiredis.RedisDB, "10")
	if parsed, err := strconv.ParseFloat(thresholdStr, 64); err == nil {
		minImprovementPct = parsed
	}

	var currentReset time.Time

	if *dateOverride != "" {
		parsed, err := time.Parse("2006-01-02", *dateOverride)
		if err != nil {
			log.Fatalf("Invalid date format %q, expected YYYY-MM-DD: %v", *dateOverride, err)
		}
		// Floor to the first culvert reset of that month
		firstOfMonth := time.Date(parsed.Year(), parsed.Month(), 1, 0, 0, 0, 0, parsed.Location())
		for firstOfMonth.Weekday() != helpers.GetCulvertResetDay(firstOfMonth) {
			firstOfMonth = firstOfMonth.AddDate(0, 0, 1)
		}
		currentReset = firstOfMonth
		log.Printf("Date override: floored to first reset of month %s", currentReset.Format("2006-01-02"))
	} else {
		now := time.Now()
		currentReset = helpers.GetCulvertResetDate(now)
		previousReset := helpers.GetCulvertPreviousDate(currentReset)

		// Month-boundary guard: only run on the first culvert reset of the month
		if currentReset.Month() == previousReset.Month() {
			log.Println("Not the first culvert reset of the month, skipping.")
			return
		}
	}

	// Check if there is at least 1 non-zero score for this month's first reset
	var scoreCount int
	err := db.DB.QueryRow(
		"SELECT COUNT(*) FROM character_culvert_scores WHERE culvert_date = $1 AND score > 0",
		currentReset.Format("2006-01-02"),
	).Scan(&scoreCount)
	if err != nil || scoreCount == 0 {
		log.Println("No non-zero scores found for the first week of this month, skipping.")
		return
	}

	log.Println("Running monthly improvements report.")

	// Calculate the first culvert reset of the previous month
	prevMonth := time.Date(currentReset.Year(), currentReset.Month()-1, 1, 0, 0, 0, 0, currentReset.Location())
	for prevMonth.Weekday() != helpers.GetCulvertResetDay(prevMonth) {
		prevMonth = prevMonth.AddDate(0, 0, 1)
	}
	currentDateStr := currentReset.Format("2006-01-02")
	prevMonthStr := prevMonth.Format("2006-01-02")

	log.Printf("Comparing scores: %s → %s\n", prevMonthStr, currentDateStr)

	// Get all active characters
	activeChars, err := helpers.GetActiveCharacters(apiredis.RedisDB, db.DB)
	if err != nil {
		log.Println("Failed to get active characters:", err)
		return
	}

	if activeChars == nil || len(*activeChars) == 0 {
		log.Println("No active characters found.")
		return
	}

	// Query scores for all active characters
	var improvements []characterImprovement

	for _, char := range *activeChars {
		var oldScore, newScore int32

		// Get old high score (best score before the previous month's first reset)
		row := db.DB.QueryRow(
			"SELECT COALESCE(MAX(score), 0) FROM character_culvert_scores WHERE character_id = $1 AND culvert_date < $2 AND score > 0",
			char.ID, prevMonthStr,
		)
		if err := row.Scan(&oldScore); err != nil || oldScore <= 0 {
			continue // Skip if no previous scores
		}

		// Get new high score (best score from previous month's first reset through current month's first reset)
		row = db.DB.QueryRow(
			"SELECT COALESCE(MAX(score), 0) FROM character_culvert_scores WHERE character_id = $1 AND culvert_date >= $2 AND culvert_date <= $3 AND score > 0",
			char.ID, prevMonthStr, currentDateStr,
		)
		if err := row.Scan(&newScore); err != nil || newScore <= 0 {
			continue // Skip if no scores in the comparison period
		}

		improvementPct := (float64(newScore-oldScore) / float64(oldScore)) * 100.0
		improvements = append(improvements, characterImprovement{
			CharacterName:  char.MapleCharacterName,
			OldScore:       oldScore,
			NewScore:       newScore,
			ImprovementPct: improvementPct,
		})
	}

	// Check if we already sent a message for this month
	var existingMessageID, existingChannelID, existingFluffText string
	existingRow := db.DB.QueryRow(
		"SELECT message_id, channel_id, fluff_text FROM discord_monthly_improvements WHERE month = $1",
		currentDateStr,
	)
	isEdit := false
	if err := existingRow.Scan(&existingMessageID, &existingChannelID, &existingFluffText); err == nil {
		isEdit = true
		log.Printf("Found existing message %s for month %s, will edit.", existingMessageID, currentDateStr)
	} else if err != sql.ErrNoRows {
		log.Println("Failed to query existing monthly improvements message:", err)
		return
	}

	if len(improvements) == 0 && !isEdit {
		log.Println("No characters with valid scores at both dates and no existing message to edit.")
		return
	}

	// Sort all by improvement percentage descending
	sort.Slice(improvements, func(i, j int) bool {
		return improvements[i].ImprovementPct > improvements[j].ImprovementPct
	})

	// Split into improvers and slackers
	var improvers []characterImprovement
	hasSlackers := false

	for _, imp := range improvements {
		if imp.ImprovementPct >= minImprovementPct {
			improvers = append(improvers, imp)
		} else {
			hasSlackers = true
		}
	}

	// Determine fluff text: reuse from DB if editing, otherwise pick random
	fluffText := existingFluffText
	if !isEdit {
		fluffText = slackerRoasts[rand.Intn(len(slackerRoasts))]
	}

	embeddedData := buildEmbed(improvers, hasSlackers, fluffText, prevMonthStr, currentDateStr, minImprovementPct)

	// Send or edit on Discord
	s, err = discordgo.New("Bot " + os.Getenv(data.EnvVarDiscordToken))
	if err != nil {
		log.Printf("Invalid bot parameters: %v", err)
		return
	}
	stop := make(chan struct{}, 1)
	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)

		channelID := apiredis.CONF_DISCORD_MEMBERS_MAIN_CHANNEL_ID.GetWithDefault(apiredis.RedisDB, "")
		if channelID == "" {
			log.Println("CONF_DISCORD_MEMBERS_MAIN_CHANNEL_ID is not set, cannot send message.")
			stop <- struct{}{}
			return
		}

		if isEdit {
			// Edit existing message
			_, err := s.ChannelMessageEditComplex(&discordgo.MessageEdit{
				Channel: existingChannelID,
				ID:      existingMessageID,
				Embeds:  &[]*discordgo.MessageEmbed{embeddedData},
			})
			if err != nil {
				log.Println("Failed to edit monthly improvements message:", err)
			} else {
				log.Println("Monthly improvements message edited successfully!")
			}
		} else {
			// Send new message
			msg, err := s.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
				Embeds: []*discordgo.MessageEmbed{embeddedData},
			})
			if err != nil {
				log.Println("Failed to send monthly improvements message:", err)
				stop <- struct{}{}
				return
			}
			log.Println("Monthly improvements message sent successfully!")

			// Save message ID, channel ID, and fluff text to DB
			_, err = db.DB.Exec(
				"INSERT INTO discord_monthly_improvements (month, message_id, channel_id, fluff_text) VALUES ($1, $2, $3, $4)",
				currentDateStr, msg.ID, channelID, fluffText,
			)
			if err != nil {
				log.Println("Failed to save monthly improvements message to DB:", err)
			}
		}

		stop <- struct{}{}
	})
	err = s.Open()
	if err != nil {
		log.Printf("Cannot open the session: %v", err)
		return
	}
	defer s.Close()
	<-stop
}
