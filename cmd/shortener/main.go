package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// var mainHost *string
// var resultHost *string

func main() {
	// var ok bool
	// *mainHost, ok = os.LookupEnv("SERVER_ADDRESS")
	// if !ok {
	// 	mainHost = flag.String("a", ":8080", "Base host adress")
	// }
	// *resultHost, ok = os.LookupEnv("BASE_URL")
	// if !ok {
	// 	resultHost = flag.String("b", "http://localhost:8080", "Result host adress")
	// }
	flag.Parse()

	r := chi.NewRouter()

	r.Post("/", shortener)
	r.Get("/{short_url}", expander)

	log.Fatal(http.ListenAndServe("localhost:8080", r))
}
