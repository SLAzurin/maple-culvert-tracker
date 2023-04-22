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
	},
	{
		Name:                     "login",
		Description:              "Gives you a temporary login code",
		DefaultMemberPermissions: &dmPermissions,
		DMPermission:             &dmPermission,
	},
}
