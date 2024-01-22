package handlers

import (
	"encoding/json"
	"errors"
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

type URLData struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type URLResponse struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

type Handlers struct {
	Storage *storage.Storage
	BaseURL string
}

func NewHandlers(config *config.Config) (*Handlers, error) {
	storage, err := storage.NewStorage(config)
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
	str, err := h.Storage.Add(shortURL, URL)
	for errors.Is(err, storage.ErrDuplicateValue) {
		shortURL = "/" + utils.GenerateShortKey()
		str, err = h.Storage.Add(shortURL, URL)
	}
	if errors.Is(err, storage.ErrExistingFullURL) {
		return h.BaseURL + str, err
	}
	if err != nil {
		return "", fmt.Errorf("failed to add new key-value pair in storage: %w", err)
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
	if errors.Is(err, storage.ErrExistingFullURL) {
		res.Header().Add("Content-Type", "application/json")
		res.WriteHeader(http.StatusConflict)
		result := Result{Result: str}
		json.NewEncoder(res).Encode(result)
		return
	}
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	res.Header().Add("Content-Type", "application/json")
	res.WriteHeader(http.StatusCreated)
	result := Result{Result: str}
	json.NewEncoder(res).Encode(result)
}

func (h Handlers) ShortenBatch(res http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(req.Body)
	var urlDataList []URLData
	var urlResponseList []URLResponse
	err := decoder.Decode(&urlDataList)
	if err != nil {
		log.Info().Err(err).Msg("Не удалось распарсить запрос: ")
		http.Error(res, "Не удалось распарсить запрос", http.StatusBadRequest)
		return
	}
	for _, data := range urlDataList {
		str, err := h.Shorten(data.OriginalURL)
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
		urlResponseList = append(urlResponseList, URLResponse{CorrelationID: data.CorrelationID, ShortURL: str})
	}
	res.Header().Add("Content-Type", "application/json")
	res.WriteHeader(http.StatusCreated)
	json.NewEncoder(res).Encode(urlResponseList)
}

func (h Handlers) Shortener(res http.ResponseWriter, req *http.Request) {
	fullURL, err := io.ReadAll(req.Body)
	if err != nil {
		log.Error().Err(err).Msg("Не удалось распарсить запрос: ")
		http.Error(res, "Не удалось распарсить запрос", http.StatusBadRequest)
		return
	}
	str, err := h.Shorten(string(fullURL))
	if errors.Is(err, storage.ErrExistingFullURL) {
		res.WriteHeader(http.StatusConflict)
		io.WriteString(res, str)
		return
	}
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	res.WriteHeader(http.StatusCreated)
	io.WriteString(res, str)
}

func (h Handlers) Expander(res http.ResponseWriter, req *http.Request) {
	log.Info().Msg(req.URL.Path)
	fullURL, ok, err := h.Storage.Get(req.URL.Path)
	if err != nil {
		http.Error(res, "Внутренняя ошибка: "+err.Error(), http.StatusInternalServerError)
	}
	if !ok {
		http.Error(res, "Такой ссылки нет", http.StatusBadRequest)
		return
	}
	res.Header().Set("Location", fullURL)
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusTemporaryRedirect)
}

func (h Handlers) Ping(res http.ResponseWriter, req *http.Request) {
	ok := h.Storage.CheckDBConn()
	if !ok {
		http.Error(res, "Соединение с БД отсутствует", http.StatusBadRequest)
		return
	}
	res.WriteHeader(http.StatusOK)
}
