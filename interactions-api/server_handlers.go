package interactionsapi

import (
	"encoding/json"
	"net/http"

	"github.com/bwmarrin/discordgo"
	logger "github.com/c032/go-logger"
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
	resp := &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Temporarily unavailable, this feature is currently not implemented.",
		},
	}

	s.respondJSON(200, w, resp)
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
