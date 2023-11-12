package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
)

var mainPtr *string
var mainHost string
var resultPtr *string
var resultHost string

func main() {
	mainPtr = flag.String("a", "localhost:8080", "Base host adress")
	resultPtr = flag.String("b", "http://localhost:8080", "Result host adress")
	flag.Parse()

	var ok bool
	mainHost, ok = os.LookupEnv("SERVER_ADDRESS")
	if !ok {
		mainHost = *mainPtr
	}
	resultHost, ok = os.LookupEnv("BASE_URL")
	if !ok {
		resultHost = *resultPtr
	}

	r := chi.NewRouter()

	r.Post("/", shortener)
	r.Post("/api/shorten", shortenerJSON)

	// if mainHost == resultHost {
	r.Get("/{short_url}", expander)
	// } else {
	// 	r1 := chi.NewRouter()
	// 	r1.Get("/{short_url}", expander)
	// 	go func() {
	// 		log.Fatal(http.ListenAndServe(resultHost, r1))
	// 	}()
	// }
	log.Fatal(http.ListenAndServe(mainHost, r))

}
