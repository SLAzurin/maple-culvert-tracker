package helpers

import (
	"log"
	"os"
	"strconv"

	"github.com/bwmarrin/discordgo"
	"github.com/slazurin/maple-culvert-tracker/internal/apiredis"
)

type announcement struct {
	ComplexMessage       *discordgo.MessageSend
	FilePathsAttachments []string
}

func getAnnouncements() []announcement {
	announcements := []announcement{
		{
			// 0
			ComplexMessage: &discordgo.MessageSend{
				Content: `New features have landed!
- I now tell you all about new features when they get implemented!
- You can now use ` + "`/track-character`" + ` to track new characters using a command instead of logging in! Go check it out!

If anything is broken, please report directly to ` + "`.azuri`.\nEnjoy!",
			},
		},
		{
			// 1
			FilePathsAttachments: []string{
				"./server_assets/announcement1.png",
			},
			ComplexMessage: &discordgo.MessageSend{
				Content: `New features have landed!
- The admin panel now saves unsubmitted scores, so that you won't lose your work-in-progress if you accidentally close your tab!
    - _Small fair warning: Admin console changes are very prone to bugs, but I made sure to test as much as I can before I release it._

If anything is broken, please report directly to ` + "`.azuri`!\nEnjoy the new features!",
			},
		},
		{
			// 2 database migrations
			ComplexMessage: &discordgo.MessageSend{
				Content: `Minor internal database changes were applied to prepare for future updates.
If anything is broken, please report directly to ` + "`.azuri`! Thank you",
			},
		},
		{
			// 3 move reminder to reset -10 minutes
			ComplexMessage: &discordgo.MessageSend{
				Content: `Hi everyone! I've moved the daily update scores reminder to 5 minutes before reset.
Guild members can still run last minute culvert at weekly reset -10 minutes so it'll give them enough time if they're running really late.

That's it, Bot out!`,
			},
		},
		{
			// 4 fix default week to most recent week
			ComplexMessage: &discordgo.MessageSend{
				Content: `Hi everyone! I've fixed the admin panel __not__ defaulting to the most recent week!

Bot out!`,
			},
		},
		{
			// 5 new command culvert-mega-chart and culvert-summary
			ComplexMessage: &discordgo.MessageSend{
				Content: `Hi everyone!

I've added a new dumb command: ` + "`/culvert-mega-chart`" + `!
This command will show the past culvert progression for the whole entire guild.

` + "`/culvert-summary`" + ` is also here now!
You can snoop on the whole guild's past and present culvert scores with this command!
You can choose to order the results by character name or score (default is score).

Bot out!`,
			},
		},
		{
			// 6 new diff summary when submitting scores
			ComplexMessage: &discordgo.MessageSend{
				Content: `Hi everyone!

I've changed the notification when scores are submitted!

The notification in the admin discord channel will now show the culvert scores compared to the previous week.
You will immediately notice this change next time you update your scores. You can't miss it!

Bot out!`,
			},
		},
		{
			// 7 sandbag announcement
			ComplexMessage: &discordgo.MessageSend{
				Content: `I did something. Hehe.`,
			},
		},
		{
			// 8 sandbag announcement culvert-duel
			ComplexMessage: &discordgo.MessageSend{
				Content: `I did something else. Hehe.`,
			},
		},
	}
	return announcements
}

func announceNewFeatures(s *discordgo.Session) {
	log.Println("Announcing new features")
	adminChannelID, err := apiredis.CONF_DISCORD_ADMIN_CHANNEL_ID.Get(apiredis.RedisDB)
	if err != nil {
		log.Println("Failed to announce new features, failed to get main channel ID", err)
		return
	}

	announcements := getAnnouncements()
	announcedVersionStr := apiredis.DATA_DISCORD_NEW_FEATURES_ANNOUNCEMENT_VERSION.GetWithDefault(apiredis.RedisDB, strconv.Itoa(len(announcements)-1))

	announcedVersion := 0

	announcedVersion, err = strconv.Atoi(announcedVersionStr)
	if err != nil {
		announcedVersion = len(announcements) - 1
	}

	hadAnnounced := false

	for announcedVersion < len(announcements) {
		log.Println("Announcing version " + strconv.Itoa(announcedVersion))

		msgComplex := announcements[announcedVersion].ComplexMessage
		if len(announcements[announcedVersion].FilePathsAttachments) > 0 {
			msgComplex.Files = []*discordgo.File{}
		}

		for _, v := range announcements[announcedVersion].FilePathsAttachments {
			f, err := os.Open(v)
			if err != nil {
				log.Println("Failed to announce new features, failed to open file", err)
				return
			}
			defer f.Close()
			msgComplex.Files = append(msgComplex.Files, &discordgo.File{Name: v, Reader: f})
		}

		s.ChannelMessageSendComplex(adminChannelID, msgComplex)

		announcedVersion++
		hadAnnounced = true
	}
	if hadAnnounced {
		apiredis.DATA_DISCORD_NEW_FEATURES_ANNOUNCEMENT_VERSION.Set(apiredis.RedisDB, strconv.Itoa(announcedVersion))
	}
}
