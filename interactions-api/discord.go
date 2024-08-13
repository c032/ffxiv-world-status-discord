package interactionsapi

import (
	"slices"
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/c032/ffxiv-world-status-discord/ffxivapi"
)

type Worlds []ffxivapi.World

func (worlds Worlds) Embed(title string, thumbnailURL string) (*discordgo.MessageEmbed, error) {
	groups := map[string][]string{}

	for _, w := range worlds {
		groups[w.Group] = append(groups[w.Group], w.Name)
	}

	var groupNames []string
	for groupName, _ := range groups {
		groupNames = append(groupNames, groupName)
	}
	slices.Sort(groupNames)

	var fields []*discordgo.MessageEmbedField
	for _, groupName := range groupNames {
		worldNames := groups[groupName]

		slices.Sort(worldNames)

		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   groupName,
			Value:  strings.Join(worldNames, "\n"),
			Inline: true,
		})
	}

	embed := &discordgo.MessageEmbed{
		Title:  title,
		Fields: fields,
	}

	if thumbnailURL != "" {
		embed.Thumbnail = &discordgo.MessageEmbedThumbnail{
			URL: thumbnailURL,
		}
	}

	return embed, nil
}
