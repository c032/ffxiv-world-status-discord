package interactionsapi

import (
	"github.com/bwmarrin/discordgo"
)

const (
	CmdPing       = "ping"
	CmdCharacters = "characters"
)

var Commands = map[string]*discordgo.ApplicationCommand{
	CmdPing: &discordgo.ApplicationCommand{
		Description: "Make the bot respond with a pong.",
	},
	CmdCharacters: &discordgo.ApplicationCommand{
		Description: "Print character creation availability status of all worlds.",
	},
}

func init() {
	for key, cmd := range Commands {
		cmd.Name = key
	}
}
