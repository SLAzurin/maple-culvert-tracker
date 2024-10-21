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
			FilePathsAttachments: []string{
				"./server_assets/announcement0.png",
			},
			ComplexMessage: &discordgo.MessageSend{
				Content: `New features have landed!
- I now tell you all about new features when they get implemented!
- You can now use ` + "`/track-character`" + ` to track new characters using a command instead of logging in! Go check it out!

If anything is broken, please report directly to ` + "`.azuri`.\nEnjoy!",
			},
		},
	}
	return announcements
}

// - The admin panel now saves unsubmitted scores, so that you won't lose your work-in-progress if you accidentally close your tab!, revert this: bf69ff1321fb60e744ecda98d7253ff2a17e8cfa

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
