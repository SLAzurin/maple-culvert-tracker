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
		Name:        "culvert-duel",
		Description: "Flex your culvert score against your Guildmate",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Required:    true,
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "your-character",
				Description: "Your character's name",
			},
			{
				Required:    true,
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "their-character",
				Description: "Their character's name",
			},
		},
	},
	{
		Name:        "culvert-duel-anyone",
		Description: "MODS ONLY: Select 2 guild members to duel their culvert scores",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Required:    true,
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "your-character",
				Description: "Your character's name",
			},
			{
				Required:    true,
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "their-character",
				Description: "Their character's name",
			},
		},
	},
	{
		Name:        "sandbaggers",
		Description: "Shows players with most sandbagged runs over the past 12 weeks",
	},
	{
		Name:        "export-csv",
		Description: "Export weeks of data to csv format (Compatible with all spreadsheet software)",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Required:    false,
				Type:        discordgo.ApplicationCommandOptionInteger,
				MinValue:    &culvertMinWeeks,
				MaxValue:    52,
				Name:        "weeks",
				Description: "Number of weeks to display in the graph",
			},
			{
				Required:    false,
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "date",
				Description: "Date in YYYY-MM-DD format to check historical data",
			},
		},
	},
}
