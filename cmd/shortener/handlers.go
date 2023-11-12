package main

import (
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func shortener(res http.ResponseWriter, req *http.Request) {
	full_url, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(res, "Не удалось распарсить запрос", http.StatusBadRequest)
		return
	}
	short := generateShortKey()
	storage[short] = string(full_url)
	res.WriteHeader(http.StatusCreated)
	io.WriteString(res, *resultHost+"/"+short)
}

func expander(res http.ResponseWriter, req *http.Request) {
	full_url, ok := storage[chi.URLParam(req, "short_url")]
	if !ok {
		http.Error(res, "Такой ссылки нет", http.StatusBadRequest)
		return
	}
	res.Header().Set("Location", full_url)
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
