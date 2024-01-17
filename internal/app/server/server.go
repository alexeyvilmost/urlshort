package server

import (
	"fmt"
	"net/http"

	"github.com/alexeyvilmost/urlshort.git/internal/app/compressing"
	"github.com/alexeyvilmost/urlshort.git/internal/app/config"
	"github.com/alexeyvilmost/urlshort.git/internal/app/handlers"
	"github.com/alexeyvilmost/urlshort.git/internal/app/logging"
	"github.com/go-chi/chi"
	"github.com/rs/zerolog"
)

func StartServer() error {
	config := config.NewConfig()
	handlers, err := handlers.NewHandlers(config)
	if err != nil {
		return fmt.Errorf("failed to create handlers: %w", err)
	}

	r := chi.NewRouter()
	r.Use(compressing.WithCompress, logging.WithLogging)
	r.Post("/", handlers.Shortener)
	r.Post("/api/shorten", handlers.ShortenerJSON)
	r.Get("/{short_url}", handlers.Expander)
	r.Get("/ping", handlers.Ping)
	zerolog.SetGlobalLevel(config.LogLevel)
	err = http.ListenAndServe(config.ServerAddress, r)
	return err
}
