package commands

import (
	"github.com/bwmarrin/discordgo"
)

var culvertMinWeeks = float64(8)

var Commands = []*discordgo.ApplicationCommand{
	{
		Name:        "ping",
		Description: "Shows user details",
	},
	{
		Name:        "culvert",
		Description: "Shows your past culvert progression",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Required:    false,
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "character-name",
				Description: "Your character's name",
			},
			{
				Required:    false,
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "date",
				Description: "Date in YYYY-MM-DD format to check historical data",
			},
			{
				Required:    false,
				Type:        discordgo.ApplicationCommandOptionInteger,
				MinValue:    &culvertMinWeeks,
				MaxValue:    52,
				Name:        "weeks",
				Description: "Number of weeks to display in the graph",
			},
		},
	},
	{
		Name:        "culvert-anyone",
		Description: "Shows the past culvert progression for a given character",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Required:    false,
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "character-name",
				Description: "The character's name",
			},
			{
				Required:    false,
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "date",
				Description: "Date in YYYY-MM-DD format to check historical data",
			},
			{
				Required:    false,
				Type:        discordgo.ApplicationCommandOptionInteger,
				MinValue:    &culvertMinWeeks,
				MaxValue:    52,
				Name:        "weeks",
				Description: "Number of weeks to display in the graph",
			},
		},
	},
	{
		Name:        "login",
		Description: "Gives you a temporary login code for the Admin Console",
	},
	{
		Name:        "sandbaggers",
		Description: "Shows players with most sandbagged runs over the past 12 weeks",
	},
}
