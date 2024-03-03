package commands

import (
	"github.com/bwmarrin/discordgo"
)

var dmPermission = false
var dmPermissions int64 = discordgo.PermissionBanMembers

var Commands = []*discordgo.ApplicationCommand{
	{
		Name:        "ping",
		Description: "Shows user details",
	},
	{
		Name:        "culvert",
		Description: "Shows your past 12 months culvert progression",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Required:    false,
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "character-name",
				Description: "Your character's name",
			},
		},
	},
	{
		Name:        "culvert-anyone",
		Description: "Shows the past 12 months culvert progression for a given character",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Required:    false,
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "character-name",
				Description: "The character's name",
			},
		},
		DefaultMemberPermissions: &dmPermissions,
		DMPermission:             &dmPermission,
	},
	{
		Name:                     "login",
		Description:              "Gives you a temporary login code",
		DefaultMemberPermissions: &dmPermissions,
		DMPermission:             &dmPermission,
	},
}
