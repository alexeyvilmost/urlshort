package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/rs/zerolog/log"

	"github.com/alexeyvilmost/urlshort.git/internal/app/config"
	"github.com/alexeyvilmost/urlshort.git/internal/app/storage"
	"github.com/alexeyvilmost/urlshort.git/internal/app/utils"
)

type Result struct {
	Result string `json:"result"`
}

type Request struct {
	URL string `json:"url"`
}

type Handlers struct {
	BaseURL string
	Storage *storage.Storage
}

func NewHandlers(config *config.Config) (*Handlers, error) {
	storage, err := storage.NewStorage(config.StorageFile)
	if err != nil {
		return &Handlers{}, fmt.Errorf("failed to create storage: %w", err)
	}
	result := &Handlers{
		BaseURL: config.BaseURL,
		Storage: storage,
	}
	return result, nil
}

func (h Handlers) Shorten(URL string) (string, error) {
	shortURL := "/" + utils.GenerateShortKey()
	err := h.Storage.Add(shortURL, URL)
	for err == storage.ErrDuplicateValue {
		shortURL = "/" + utils.GenerateShortKey()
		err = h.Storage.Add(shortURL, URL)
	}
	if err != nil {
		return "", fmt.Errorf("Failed to add new key-value pair in storage: %w", err)
	}
	return h.BaseURL + shortURL, nil
}

func (h Handlers) ShortenerJSON(res http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(req.Body)
	var url Request
	err := decoder.Decode(&url)
	if err != nil {
		log.Info().Err(err).Msg("Не удалось распарсить запрос: ")
		http.Error(res, "Не удалось распарсить запрос", http.StatusBadRequest)
		return
	}
	str, err := h.Shorten(url.URL)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	res.Header().Add("Content-Type", "application/json")
	res.WriteHeader(http.StatusCreated)
	result := Result{Result: str}
	json.NewEncoder(res).Encode(result)
}

func (h Handlers) Shortener(res http.ResponseWriter, req *http.Request) {
	fullURL, err := io.ReadAll(req.Body)
	if err != nil {
		log.Error().Err(err).Msg("Не удалось распарсить запрос: ")
		http.Error(res, "Не удалось распарсить запрос", http.StatusBadRequest)
		return
	}
	str, err := h.Shorten(string(fullURL))
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	res.WriteHeader(http.StatusCreated)
	io.WriteString(res, str)
}

func (h Handlers) Expander(res http.ResponseWriter, req *http.Request) {
	fullURL, ok := h.Storage.Get(req.URL.Path)
	if !ok {
		http.Error(res, "Такой ссылки нет", http.StatusBadRequest)
		return
	}
	res.Header().Set("Location", fullURL)
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusTemporaryRedirect)
}
