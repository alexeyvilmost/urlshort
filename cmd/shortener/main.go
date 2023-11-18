package main

import (
	"log"
	"net/http"

	"github.com/alexeyvilmost/urlshort.git/internal/app/handlers"
	"github.com/alexeyvilmost/urlshort.git/internal/app/utils"
	"github.com/go-chi/chi"
)

func StartServer() {
	config := utils.NewConfig()
	handlers := handlers.NewHandlers(config)

	r := chi.NewRouter()
	r.Post("/", handlers.Shortener)
	r.Post("/api/shorten", handlers.ShortenerJSON)
	r.Get("/{short_url}", handlers.Expander)
	log.Println(config.ServerAddress)
	err := http.ListenAndServe("localhost:8080", r)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	StartServer()
}
