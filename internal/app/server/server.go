package server

import (
	"log"
	"net/http"

	"github.com/alexeyvilmost/urlshort.git/internal/app/config"
	"github.com/alexeyvilmost/urlshort.git/internal/app/handlers"
	"github.com/alexeyvilmost/urlshort.git/internal/app/logging"
	"github.com/go-chi/chi"
	"github.com/rs/zerolog"
)

func StartServer() error {
	config := config.NewConfig()
	handlers := handlers.NewHandlers(config)

	r := chi.NewRouter()
	r.Post("/", logging.WithLogging(handlers.Shortener))
	r.Post("/api/shorten", logging.WithLogging(handlers.ShortenerJSON))
	r.Get("/{short_url}", logging.WithLogging(handlers.Expander))
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	log.Println(config.ServerAddress)
	err := http.ListenAndServe(config.ServerAddress, r)
	return err
}
