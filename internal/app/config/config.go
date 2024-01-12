package config

import (
	"flag"
	"os"

	"github.com/rs/zerolog"
)

type Config struct {
	ServerAddress string
	BaseURL       string
	StorageFile   string
	LogLevel      zerolog.Level
}

func NewConfig() *Config {
	result := new(Config)
	mainPtr := flag.String("a", "localhost:8080", "Base host adress")
	resultPtr := flag.String("b", "http://localhost:8080", "Result host adress")
	StorageFile := flag.String("f", "storage.txt", "Storage filename")
	LogLevel := flag.String("l", "d", "Log level: 'd' for debug, 'i' for info, 'w' for warn and 'e' for error")
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
	result.StorageFile, ok = os.LookupEnv("FILE_STORAGE_PATH")
	if !ok {
		result.StorageFile = *StorageFile
	}
	switch *LogLevel {
	case "d":
		result.LogLevel = zerolog.DebugLevel
	case "i":
		result.LogLevel = zerolog.InfoLevel
	case "w":
		result.LogLevel = zerolog.WarnLevel
	case "e":
		result.LogLevel = zerolog.ErrorLevel
	}
	return result
}
