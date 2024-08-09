package interactionsapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/bwmarrin/discordgo"
	logger "github.com/c032/go-logger"

	"github.com/c032/ffxiv-world-status-discord/ffxivapi"
)

func (s *Server) handleInteractionPing(interaction *discordgo.Interaction, w http.ResponseWriter, req *http.Request) {
	resp := &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponsePong,
	}

	s.respondJSON(200, w, resp)
}

func (s *Server) handleCommandPing(data discordgo.ApplicationCommandInteractionData, w http.ResponseWriter) {
	resp := &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Pong.",
		},
	}

	s.respondJSON(200, w, resp)
}

func (s *Server) handleCommandStatus(data discordgo.ApplicationCommandInteractionData, w http.ResponseWriter) {
	log := s.logger()

	var (
		maintenanceWorlds                  []ffxivapi.World
		characterCreationUnavailableWorlds []ffxivapi.World

		err error
		wr  *ffxivapi.WorldsResponse
	)

	wr, err = s.API.Worlds()
	if err != nil {
		log.Error(err.Error())

		s.respondJSON(200, w, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Could not check availability.",
			},
		})

		return
	}

	for _, world := range wr.Worlds {
		if world.IsMaintenance {
			maintenanceWorlds = append(maintenanceWorlds, world)
		}
		if !world.CanCreateNewCharacters {
			characterCreationUnavailableWorlds = append(characterCreationUnavailableWorlds, world)
		}
	}

	var fields []*discordgo.MessageEmbedField

	if len(maintenanceWorlds) > 0 {
		var names []string

		for _, w := range maintenanceWorlds {
			names = append(names, fmt.Sprintf("%s (%s)", w.Name, w.Group))
		}

		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   "Maintenance",
			Value:  strings.Join(names, "\n"),
			Inline: true,
		})
	}

	if len(characterCreationUnavailableWorlds) > 0 {
		var names []string

		for _, w := range characterCreationUnavailableWorlds {
			names = append(names, fmt.Sprintf("%s (%s)", w.Name, w.Group))
		}

		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   "Character creation unavailable",
			Value:  strings.Join(names, "\n"),
			Inline: true,
		})
	}

	var content string
	var embeds []*discordgo.MessageEmbed

	if len(fields) == 0 {
		content = "Everything looks good."
	} else {
		embed := &discordgo.MessageEmbed{
			Title:  "FFXIV World Status",
			Fields: fields,
		}

		if s.DiscordThumbnailURL != "" {
			embed.Thumbnail = &discordgo.MessageEmbedThumbnail{
				URL: s.DiscordThumbnailURL,
			}
		}

		embeds = append(embeds, embed)
	}

	interactionResponse := &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds:  embeds,
			Content: content,
		},
	}

	s.respondJSON(200, w, interactionResponse)
}

func (s *Server) handleInteractionApplicationCommand(interaction *discordgo.Interaction, w http.ResponseWriter, req *http.Request) {
	log := s.logger()

	// This might panic.
	data := interaction.ApplicationCommandData()

	log.WithFields(logger.Fields{
		"data": data,
	}).Print("Processing command.")

	switch data.Name {
	case CmdPing:
		s.handleCommandPing(data, w)
	case CmdStatus:
		s.handleCommandStatus(data, w)
	default:
		log.Print("Command not recognized: %s", data.Name)

		s.respondError(w, ErrorResponse{
			Status: 400,
			Type:   ErrTypeUnknownCommand,
		})

		return
	}
}

func (s *Server) handleInteractionRequest(w http.ResponseWriter, req *http.Request) {
	log := s.logger()

	var interaction *discordgo.Interaction

	dec := json.NewDecoder(req.Body)
	err := dec.Decode(&interaction)
	if err != nil {
		log.Errorf("could not request decode body as JSON: %s", err.Error())

		http.Error(w, "", http.StatusInternalServerError)

		return
	}

	switch interaction.Type {
	case discordgo.InteractionPing:
		s.handleInteractionPing(interaction, w, req)
	case discordgo.InteractionApplicationCommand:
		s.handleInteractionApplicationCommand(interaction, w, req)
	default:
		http.Error(w, "", http.StatusBadRequest)
	}
}
