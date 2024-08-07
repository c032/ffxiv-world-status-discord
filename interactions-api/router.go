package interactionsapi

import (
	"net/http"

	chi "github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func createRouter(s *Server) (*chi.Mux, error) {
	r := chi.NewRouter()

	r.Use(middleware.RequestLogger(&middleware.DefaultLogFormatter{
		Logger:  s.Logger,
		NoColor: true,
	}))

	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			if !s.isAuthorizedRequest(req) {
				http.Error(w, "", http.StatusUnauthorized)

				return
			}

			next.ServeHTTP(w, req)
		})
	})

	r.Post("/interactions", s.handleInteractionRequest)

	return r, nil
}
