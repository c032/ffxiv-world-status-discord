package interactionsapi

import (
	"github.com/bwmarrin/discordgo"
)

const (
	CmdPing   = "ping"
	CmdStatus = "ffxiv-status"
)

var Commands = map[string]*discordgo.ApplicationCommand{
	CmdPing: &discordgo.ApplicationCommand{
		Description: "Make the bot respond with a pong.",
	},
	CmdStatus: &discordgo.ApplicationCommand{
		Description: "Print status of all worlds.",
	},
}

func init() {
	for key, cmd := range Commands {
		cmd.Name = key
	}
}
