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
	Storage storage.StorageI
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

func (h Handlers) Shorten(URL, userID string) (string, error) {
	shortURL := utils.GenerateShortKey()
	str, err := h.Storage.Add(userID, shortURL, URL)
	for errors.Is(err, storage.ErrDuplicateValue) {
		shortURL = utils.GenerateShortKey()
		str, err = h.Storage.Add(userID, shortURL, URL)
	}
	if errors.Is(err, storage.ErrExistingFullURL) {
		return h.BaseURL + "/" + str, err
	}
	if err != nil {
		return "", fmt.Errorf("failed to add new key-value pair in storage: %w", err)
	}
	return h.BaseURL + "/" + shortURL, nil
}

func (h Handlers) ShortenerJSON(res http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(req.Body)
	var url Request
	err := decoder.Decode(&url)
	if err != nil {
		log.Error().Err(err).Msg("Не удалось распарсить запрос: ")
		http.Error(res, "Не удалось распарсить запрос", http.StatusBadRequest)
		return
	}
	userID := req.Header.Get("user-id-auth")
	str, err := h.Shorten(url.URL, userID)
	if errors.Is(err, storage.ErrExistingFullURL) {
		res.Header().Add("Content-Type", "application/json")
		res.WriteHeader(http.StatusConflict)
		result := Result{Result: str}
		json.NewEncoder(res).Encode(result)
		return
	}
	if err != nil {
		log.Error().Err(err).Msg("Не удалось добавить url")
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
		log.Error().Err(err).Msg("Не удалось распарсить запрос: ")
		http.Error(res, "Не удалось распарсить запрос", http.StatusBadRequest)
		return
	}
	userID := req.Header.Get("user-id-auth")
	for _, data := range urlDataList {
		str, err := h.Shorten(data.OriginalURL, userID)
		switch err {
		case nil:
			// pass
		case storage.ErrExistingFullURL:
			// pass
		default:
			log.Error().Err(err).Msg("Не удалось добавить url")
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
	userID := req.Header.Get("user-id-auth")
	str, err := h.Shorten(string(fullURL), userID)
	if errors.Is(err, storage.ErrExistingFullURL) {
		res.WriteHeader(http.StatusConflict)
		io.WriteString(res, str)
		return
	}
	if err != nil {
		log.Error().Err(err).Msg("Не удалось добавить url")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	res.WriteHeader(http.StatusCreated)
	io.WriteString(res, str)
}

func (h Handlers) Expander(res http.ResponseWriter, req *http.Request) {
	req.URL.Path = req.URL.Path[1:]
	fullURL, err := h.Storage.Get(req.URL.Path)
	if errors.Is(err, storage.ErrNoValue) {
		http.Error(res, "Такой ссылки нет", http.StatusBadRequest)
		return
	}
	if errors.Is(err, storage.ErrGone) {
		http.Error(res, "Ссылка была удалена", http.StatusGone)
		return
	}
	if err != nil {
		log.Info().Err(err).Msg("Внутренняя ошибка")
		http.Error(res, "Внутренняя ошибка", http.StatusInternalServerError)
		return
	}
	res.Header().Set("Location", fullURL)
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusTemporaryRedirect)
}

func (h Handlers) UserURLs(res http.ResponseWriter, req *http.Request) {
	if req.Header.Get("is-new-user") == "true" {
		http.Error(res, "Без авторизации", http.StatusUnauthorized)
		return
	}
	userID := req.Header.Get("user-id-auth")
	urls, err := h.Storage.GetUserURLs(userID)
	if err != nil {
		log.Info().Err(err).Msg("Внутренняя ошибка")
		http.Error(res, "Внутренняя ошибка", http.StatusInternalServerError)
		return
	}
	if len(urls) == 0 {
		http.Error(res, "Ссылок нет", http.StatusNoContent)
	}
	for i, v := range urls {
		urls[i].ShortURL = h.BaseURL + "/" + v.ShortURL
	}
	res.Header().Add("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	json.NewEncoder(res).Encode(urls)
}

func (h Handlers) DeteleURLs(res http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(req.Body)
	var shortURLs []string
	err := decoder.Decode(&shortURLs)
	if err != nil {
		log.Error().Err(err).Msg("Не удалось распарсить запрос: ")
		http.Error(res, "Не удалось распарсить запрос", http.StatusBadRequest)
		return
	}
	if req.Header.Get("is-new-user") == "true" {
		http.Error(res, "Без авторизации", http.StatusUnauthorized)
		return
	}
	userID := req.Header.Get("user-id-auth")
	go h.Storage.DeleteURLs(userID, shortURLs)
	res.WriteHeader(http.StatusAccepted)
}

func (h Handlers) Ping(res http.ResponseWriter, req *http.Request) {
	ok := h.Storage.CheckDBConn()
	if !ok {
		http.Error(res, "Соединение с БД отсутствует", http.StatusInternalServerError)
		return
	}
	res.WriteHeader(http.StatusOK)
}
