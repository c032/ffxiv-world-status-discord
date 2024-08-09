package main

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	logger "github.com/c032/go-logger"

	ffxivapi "github.com/c032/ffxiv-world-status-discord/ffxivapi"
	iapi "github.com/c032/ffxiv-world-status-discord/interactions-api"
)

var (
	ErrShutdown = errors.New("received signal to shutdown")
)

func mustReadRequiredEnvironmentVariable(key string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		err := fmt.Errorf("environment variable %s is missing or empty", key)

		panic(err)
	}

	return value
}

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}

	return v
}

func actualMain() int {
	log := logger.Default()

	rootCtx := context.Background()
	ctx, cancel := context.WithCancelCause(rootCtx)

	var (
		err error

		ac ffxivapi.Client
	)

	apiBaseURL := mustReadRequiredEnvironmentVariable("FFXIV_API_URL")
	apiToken := mustReadRequiredEnvironmentVariable("FFXIV_API_TOKEN")

	ac, err = ffxivapi.NewClient(ffxivapi.ClientOptions{
		BaseURL: apiBaseURL,
		Token:   apiToken,
	})
	if err != nil {
		panic(err)
	}

	rawDiscordPublicKey := strings.TrimSpace(string(must(ioutil.ReadFile(mustReadRequiredEnvironmentVariable("DISCORD_PUBLIC_KEY_FILE")))))
	discordPublicKeyBytes := must(hex.DecodeString(rawDiscordPublicKey))
	discordPublicKey := ed25519.PublicKey(discordPublicKeyBytes)

	discordToken := strings.TrimSpace(string(must(ioutil.ReadFile(mustReadRequiredEnvironmentVariable("DISCORD_TOKEN_FILE")))))
	discordApplicationID := mustReadRequiredEnvironmentVariable("DISCORD_APPLICATION_ID")
	addr := mustReadRequiredEnvironmentVariable("INTERACTIONS_API_LISTEN_ADDRESS")
	skipDiscordRequestValidation := os.Getenv("SKIP_DISCORD_REQUEST_VALIDATION") == "1"

	s := &iapi.Server{
		Logger: log,
		API:    ac,

		DiscordApplicationID:         discordApplicationID,
		DiscordPublicKey:             discordPublicKey,
		DiscordToken:                 discordToken,
		SkipDiscordRequestValidation: skipDiscordRequestValidation,
	}

	err = s.Initialize()
	if err != nil {
		panic(err)
	}
	defer s.Cleanup()

	chSignals := make(chan os.Signal, 1)
	signal.Notify(chSignals, os.Interrupt, syscall.SIGTERM)
	go func() {
		for s := range chSignals {
			if s == os.Interrupt {
				log.Print("Received SIGINT.")
			} else if s == syscall.SIGTERM {
				log.Print("Received SIGTERM.")
			} else {
				log.Print("Received unexpected signal. Ignoring.")

				continue
			}

			cancel(ErrShutdown)
		}
	}()

	const httpServerTimeout = 60 * time.Second

	hs := &http.Server{
		Addr:    addr,
		Handler: s,

		ReadTimeout:       httpServerTimeout,
		ReadHeaderTimeout: httpServerTimeout,
		WriteTimeout:      httpServerTimeout,
		IdleTimeout:       httpServerTimeout,
	}

	go func(log logger.Logger, hs *http.Server, cancel context.CancelCauseFunc) {
		log.Printf("Listening on %s", addr)

		err = hs.ListenAndServe()
		if err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				// Normal shutdown.
				//
				// No need to call `cancel` because it was already called
				// somewhere else (e.g. because of a SIGINT signal), and this
				// was the reason for the server shutting down.
			} else {
				log.Errorf("%s", err.Error())

				cancel(ErrShutdown)
			}

			return
		}
	}(log, hs, cancel)

	log.Print("Main function is ready. Waiting for interrupts.")
	<-ctx.Done()

	log.Print("Gracefully shutting down HTTP server.")

	err = hs.Shutdown(rootCtx)
	if err != nil {
		log.Errorf("Error during HTTP server shutdown: %s", err)
	}

	err = ctx.Err()
	if err != nil {
		if !errors.Is(err, context.Canceled) {
			log.Error(err)

			return 1
		}

		cause := context.Cause(ctx)
		if cause != nil {
			if cause == err {
				// Canceled without cause.
			} else {
				if errors.Is(cause, ErrShutdown) {
					// Graceful shutdown.
				} else {
					log.Error(err)

					return 1
				}
			}
		}
	}

	return 0
}

func main() {
	exitCode := actualMain()
	if exitCode != 0 {
		os.Exit(exitCode)
	}
}
