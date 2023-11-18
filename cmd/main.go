package main

import (
	"log"
	"net/http"

	"github.com/alexeyvilmost/urlshort.git/cmd/internal/handlers"
	"github.com/alexeyvilmost/urlshort.git/cmd/internal/utils"
	"github.com/go-chi/chi"
)

func StartServer() {
	config := utils.NewConfig()
	handlers := handlers.NewHandlers(config)

	r := chi.NewRouter()
	r.Post("/", handlers.Shortener)
	r.Post("/api/shorten", handlers.ShortenerJSON)
	r.Get("/{short_url}", handlers.Expander)
	log.Fatal(http.ListenAndServe(config.ServerAddress, r))
}

func main() {
	go StartServer()
}
