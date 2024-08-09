package interactionsapi

import (
	"crypto/ed25519"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/bwmarrin/discordgo"
	logger "github.com/c032/go-logger"
	chi "github.com/go-chi/chi/v5"
)

type Server struct {
	loggerMutex sync.Mutex
	Logger      logger.Logger

	API APIClient

	chiRouter *chi.Mux

	DiscordApplicationID string
	DiscordToken         string
	DiscordPublicKey     ed25519.PublicKey

	SkipDiscordRequestValidation bool

	discordSession            *discordgo.Session
	discordRegisteredCommands []*discordgo.ApplicationCommand
}

func (s *Server) logger() logger.Logger {
	s.loggerMutex.Lock()
	l := s.Logger
	s.loggerMutex.Unlock()

	if l == nil {
		return logger.Discard
	}

	return l
}

func (s *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	s.chiRouter.ServeHTTP(w, req)
}

func (s *Server) Initialize() error {
	log := s.logger()
	log.Print("Initializing server.")

	var err error

	err = s.initializeRouter()
	if err != nil {
		return fmt.Errorf("could not initialize router: %w", err)
	}

	err = s.initializeDiscordSession()
	if err != nil {
		return fmt.Errorf("could not initialize Discord session: %w", err)
	}

	err = s.initializeCommands()
	if err != nil {
		return fmt.Errorf("could not initialize Discord commands: %w", err)
	}

	log.Print("Server initializations finished.")

	return nil
}

func (s *Server) Cleanup() error {
	log := s.logger()
	log.Print("Cleaning up.")

	var err error

	err = s.cleanupCommands()
	if err != nil {
		return fmt.Errorf("could not cleanup Discord commands: %w", err)
	}

	err = s.cleanupDiscordSession()
	if err != nil {
		return fmt.Errorf("could not cleanup Discord session: %w", err)
	}

	log.Print("Server cleanup finished.")

	return nil
}

func (s *Server) initializeRouter() error {
	log := s.logger()
	log.Print("Initializing router.")

	r, err := createRouter(s)
	if err != nil {
		return fmt.Errorf("could not create router: %w", err)
	}

	s.chiRouter = r

	return nil
}

func (s *Server) initializeDiscordSession() error {
	log := s.logger()
	log.Print("Initializing Discord session.")

	ds, err := discordgo.New("Bot " + s.DiscordToken)
	if err != nil {
		return fmt.Errorf("could not initialize Discord session: %w", err)
	}

	s.discordSession = ds

	return nil
}

func (s *Server) cleanupDiscordSession() error {
	log := s.logger()
	log.Print("Closing Discord session.")

	err := s.discordSession.Close()
	if err != nil {
		return fmt.Errorf("could not close Discord session: %w", err)
	}

	log.Print("Discord session closed successfully.")

	s.discordSession = nil

	return nil
}

func (s *Server) initializeCommands() error {
	log := s.logger()
	log.Print("Initializing commands.")

	ds := s.discordSession

	registeredCommands := make([]*discordgo.ApplicationCommand, 0, len(Commands))
	for _, discordApplicationCommand := range Commands {
		if discordApplicationCommand == nil {
			continue
		}

		log.WithFields(logger.Fields{
			"command_name": discordApplicationCommand.Name,
		}).Print("Creating command.")

		const guildID = ""

		registeredCommand, err := ds.ApplicationCommandCreate(s.DiscordApplicationID, guildID, discordApplicationCommand)
		if err != nil {
			log.WithFields(logger.Fields{
				"error":        err.Error(),
				"command_name": discordApplicationCommand.Name,
			}).Errorf("Could not create command: %s", err.Error())

			continue
		}

		log.WithFields(logger.Fields{
			"command_id":   registeredCommand.ID,
			"command_name": registeredCommand.Name,
		}).Print("Command created.")

		registeredCommands = append(registeredCommands, registeredCommand)
	}

	s.discordRegisteredCommands = registeredCommands

	return nil
}

func (s *Server) cleanupCommands() error {
	log := s.logger()
	log.Print("Cleaning up commands.")

	ds := s.discordSession

	for _, cmd := range s.discordRegisteredCommands {
		log.WithFields(logger.Fields{
			"command_id":   cmd.ID,
			"command_name": cmd.Name,
		}).Print("Deleting command.")

		const guildID = ""

		ds.ApplicationCommandDelete(s.DiscordApplicationID, guildID, cmd.ID)
	}

	return nil
}

func (s *Server) isAuthorizedRequest(req *http.Request) bool {
	log := s.logger()

	if s.SkipDiscordRequestValidation {
		log.Print("Skipping request validation.")

		return true
	}

	log.Print("Validating request.")

	return discordgo.VerifyInteraction(req, s.DiscordPublicKey)
}

func (s *Server) respondError(w http.ResponseWriter, errorResponse ErrorResponse) {
	log := s.logger()

	if errorResponse.Status < 400 || errorResponse.Status > 599 {
		const defaultStatus = 500

		log.Errorf("`errorResponse` did not contain a valid status (`errorResponse.Status` is %#v). Using %d instead.", errorResponse.Status, defaultStatus)

		errorResponse.Status = defaultStatus
	}

	if errorResponse.Type == "" {
		const defaultErrorType = "unknown"

		log.Errorf("Error type was not provided. Using %#v instead.", defaultErrorType)

		errorResponse.Type = defaultErrorType
	}

	w.Header().Set("Content-Type", "application/problem+json")

	w.WriteHeader(errorResponse.Status)

	enc := json.NewEncoder(w)
	err := enc.Encode(errorResponse)
	if err != nil {
		log.Errorf("could not encode JSON to response: %s", err)
	}
}

func (s *Server) respondJSON(statusCode int, w http.ResponseWriter, v any) {
	log := s.logger()

	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(statusCode)

	enc := json.NewEncoder(w)
	err := enc.Encode(v)
	if err != nil {
		log.Errorf("could not encode JSON to response: %s", err)
	}
}
