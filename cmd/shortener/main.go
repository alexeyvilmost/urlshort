package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func main() {
	r := chi.NewRouter()

	r.Route("/", func(r chi.Router) {
		r.Post("", shortener)
		r.Get("{short_url}", expander)
	})

	log.Fatal(http.ListenAndServe(":8080", r))
}
