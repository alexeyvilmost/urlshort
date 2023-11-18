package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

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
	Storage map[string]string
}

func NewHandlers(config *utils.Config) *Handlers {
	result := new(Handlers)
	result.BaseURL = config.BaseURL
	result.Storage = map[string]string{}
	return result
}

func (h Handlers) Shorten(URL string) string {
	short := utils.GenerateShortKey(&h.Storage)
	h.Storage["/"+short] = string(URL)
	return h.BaseURL + "/" + short
}

func (h Handlers) ShortenerJSON(res http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(req.Body)
	var url Request
	err := decoder.Decode(&url)
	if err != nil {
		log.Println("Не удалось распарсить запрос")
		http.Error(res, "Не удалось распарсить запрос", http.StatusBadRequest)
		return
	}
	res.Header().Add("Content-Type", "application/json")
	res.WriteHeader(http.StatusCreated)
	result := Result{Result: h.Shorten(url.URL)}
	json.NewEncoder(res).Encode(result)
}

func (h Handlers) Shortener(res http.ResponseWriter, req *http.Request) {
	fullURL, err := io.ReadAll(req.Body)
	if err != nil {
		log.Println("Не удалось распарсить запрос")
		http.Error(res, "Не удалось распарсить запрос", http.StatusBadRequest)
		return
	}
	res.WriteHeader(http.StatusCreated)
	io.WriteString(res, h.Shorten(string(fullURL)))
}

func (h Handlers) Expander(res http.ResponseWriter, req *http.Request) {
	log.Println(len(h.Storage))
	for k, v := range h.Storage {
		log.Println(k, ": ", v)
	}
	fullURL, ok := h.Storage[req.URL.Path]
	if !ok {
		http.Error(res, "Такой ссылки нет", http.StatusBadRequest)
		return
	}
	res.Header().Set("Location", fullURL)
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusTemporaryRedirect)
}
