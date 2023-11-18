package utils

import (
	"flag"
	"os"
)

type Config struct {
	ServerAddress string
	BaseURL       string
}

func NewConfig() *Config {
	result := new(Config)
	mainPtr := flag.String("a", "localhost:8080", "Base host adress")
	resultPtr := flag.String("b", "http://localhost:8080", "Result host adress")
	flag.Parse()
	var ok bool
	result.ServerAddress, ok = os.LookupEnv("SERVER_ADDRESS")
	if !ok {
		result.ServerAddress = *mainPtr
	}
	result.BaseURL, ok = os.LookupEnv("BASE_URL")
	if !ok {
		result.BaseURL = *resultPtr
	}
	return result
}

func DefaultConfig() *Config {
	result := new(Config)
	result.ServerAddress = "http://localhost:8080"
	result.BaseURL = "localhost:8080"
}
