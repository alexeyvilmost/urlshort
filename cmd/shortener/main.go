package main

import (
	"io"
	"math/rand"
	"net/http"
	"strings"
)

var shortcuts = map[string]string{}

func generateShortKey() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const keyLength = 6

	shortKey := make([]byte, keyLength)
	for i := range shortKey {
		shortKey[i] = charset[rand.Intn(len(charset))]
	}
	return string(shortKey)
}

func shortener(res http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodPost {
		full_url, err := io.ReadAll(req.Body)
		if err != nil {
			http.Error(res, "Не удалось распарсить запрос", http.StatusBadRequest)
			return
		}
		short := generateShortKey()
		shortcuts[short] = string(full_url)
		res.WriteHeader(http.StatusCreated)
		io.WriteString(res, "http://localhost:8080/"+short)
	} else {
		uri := strings.ReplaceAll(req.RequestURI, "/", "")
		full_url, ok := shortcuts[uri]
		if !ok {
			http.Error(res, "Такой ссылки нет", http.StatusBadRequest)
			return
		}
		res.Header().Set("Location", full_url)
		res.WriteHeader(http.StatusTemporaryRedirect)
	}
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc(`/`, shortener)

	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}
