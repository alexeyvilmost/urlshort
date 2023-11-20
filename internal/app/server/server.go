package server

import (
	"log"
	"net/http"

	"github.com/alexeyvilmost/urlshort.git/internal/app/config"
	"github.com/alexeyvilmost/urlshort.git/internal/app/handlers"
	"github.com/go-chi/chi"
)

func StartServer() error {
	config := config.NewConfig()
	handlers := handlers.NewHandlers(config)

	r := chi.NewRouter()
	r.Post("/", handlers.Shortener)
	r.Post("/api/shorten", handlers.ShortenerJSON)
	r.Get("/{short_url}", handlers.Expander)
	log.Println(config.ServerAddress)
	err := http.ListenAndServe(config.ServerAddress, r)
	return err
}
