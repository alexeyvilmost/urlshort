package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Result struct {
	Result string `json:"result"`
}

type Request struct {
	URL string `json:"url"`
}

func shortenerJSON(res http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(req.Body)
	var url Request
	err := decoder.Decode(&url)
	if err != nil {
		panic(err)
	}
	short := generateShortKey()
	storage[short] = string(url.URL)
	res.Header().Add("Content-Type", "application/json")
	res.WriteHeader(http.StatusCreated)
	result := Result{Result: resultHost + "/" + short}
	json.NewEncoder(res).Encode(result)
}

func shortener(res http.ResponseWriter, req *http.Request) {
	fullUrl, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(res, "Не удалось распарсить запрос", http.StatusBadRequest)
		return
	}
	short := generateShortKey()
	storage[short] = string(fullUrl)
	res.WriteHeader(http.StatusCreated)
	io.WriteString(res, resultHost+"/"+short)
}

func expander(res http.ResponseWriter, req *http.Request) {
	fmt.Println(chi.URLParam(req, "short_url"))
	fullUrl, ok := storage[chi.URLParam(req, "short_url")]
	if !ok {
		http.Error(res, "Такой ссылки нет", http.StatusBadRequest)
		return
	}
	res.Header().Set("Location", fullUrl)
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusTemporaryRedirect)
}

func balancer(res http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodPost {
		shortener(res, req)
	} else if req.Method == http.MethodGet {
		expander(res, req)
	} else {
		res.WriteHeader(http.StatusBadRequest)
	}
}
