package main

import (
	"log"

	"github.com/alexeyvilmost/urlshort.git/internal/app/server"
)

func main() {
	err := server.StartServer()
	if err != nil {
		log.Fatal(err)
	}
}
