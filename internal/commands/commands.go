package commands

import "github.com/bwmarrin/discordgo"

var Commands = []*discordgo.ApplicationCommand{
	{
		Name:        "ping",
		Description: "Shows user details",
	},
}
