package interactionsapi

import (
	"encoding/json"
	"net/http"

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

func (s *Server) handleCommandCharacters(data discordgo.ApplicationCommandInteractionData, w http.ResponseWriter) {
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

	var embeds []*discordgo.MessageEmbed

	if len(maintenanceWorlds) > 0 {
		embed, err := Worlds(maintenanceWorlds).Embed("Maintenance", s.DiscordThumbnailURL)
		if err != nil {
			log.Print(err.Error())

			s.respondError(w, ErrorResponse{
				Status: 500,
				Type:   ErrTypeInternalServerError,
			})

			return
		}

		embeds = append(embeds, embed)
	}

	if len(characterCreationUnavailableWorlds) > 0 {
		embed, err := Worlds(characterCreationUnavailableWorlds).Embed("Character creation unavailable", s.DiscordThumbnailURL)
		if err != nil {
			log.Print(err.Error())

			s.respondError(w, ErrorResponse{
				Status: 500,
				Type:   ErrTypeInternalServerError,
			})

			return
		}

		embeds = append(embeds, embed)
	}

	var content string
	if len(embeds) == 0 {
		content = "Everything looks good."
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
	case CmdCharacters:
		s.handleCommandCharacters(data, w)
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
