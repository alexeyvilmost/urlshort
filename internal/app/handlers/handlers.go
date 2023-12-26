package handlers

import (
	"encoding/json"
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

func NewHandlers(config *config.Config) *Handlers {
	result := &Handlers{
		BaseURL: config.BaseURL,
		Storage: storage.NewStorage(config.StorageFile),
	}
	return result
}

func (h Handlers) Shorten(URL string) string {
	shortURL := "/" + utils.GenerateShortKey()
	err := h.Storage.Add(shortURL, URL)
	for err == storage.ErrDuplicateValue {
		shortURL = "/" + utils.GenerateShortKey()
		err = h.Storage.Add(shortURL, URL)
	}
	if err != nil {
		log.Error().Msg(err.Error())
	}
	return h.BaseURL + shortURL
}

func (h Handlers) ShortenerJSON(res http.ResponseWriter, req *http.Request) {
	reader, err := utils.ReadCompressed(req)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	decoder := json.NewDecoder(reader)
	var url Request
	err = decoder.Decode(&url)
	if err != nil {
		log.Info().Err(err).Msg("Не удалось распарсить запрос: ")
		http.Error(res, "Не удалось распарсить запрос", http.StatusBadRequest)
		return
	}
	res.Header().Add("Content-Type", "application/json")
	res.WriteHeader(http.StatusCreated)
	result := Result{Result: h.Shorten(url.URL)}
	json.NewEncoder(res).Encode(result)
}

func (h Handlers) Shortener(res http.ResponseWriter, req *http.Request) {
	reader, err := utils.ReadCompressed(req)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	fullURL, err := io.ReadAll(reader)
	if err != nil {
		log.Error().Err(err).Msg("Не удалось распарсить запрос: ")
		http.Error(res, "Не удалось распарсить запрос", http.StatusBadRequest)
		return
	}
	res.WriteHeader(http.StatusCreated)
	io.WriteString(res, h.Shorten(string(fullURL)))
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
