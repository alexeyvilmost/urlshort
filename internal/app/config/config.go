package config

import (
	"flag"
	"os"
)

type Config struct {
	ServerAddress string
	BaseURL       string
	StorageFile   string
}

func NewConfig() *Config {
	result := new(Config)
	mainPtr := flag.String("a", "localhost:8080", "Base host adress")
	resultPtr := flag.String("b", "http://localhost:8080", "Result host adress")
	StorageFile := flag.String("f", "storage.txt", "Storage filename")
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
	return result
}
