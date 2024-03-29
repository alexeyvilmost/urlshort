package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/alexeyvilmost/urlshort.git/internal/app/config"
	"github.com/alexeyvilmost/urlshort.git/internal/app/handlers"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_happypath(t *testing.T) {
	config := &config.Config{
		ServerAddress: "localhost:8080",
		BaseURL:       "http://localhost:8080",
		StorageFile:   "storage.txt",
	}
	h, err := handlers.NewHandlers(config)
	require.NoError(t, err)
	handler := http.HandlerFunc(h.Shortener)

	srv := httptest.NewServer(handler)
	defer srv.Close()

	req := resty.New().R()
	req.Method = http.MethodPost
	req.URL = srv.URL
	req.Body = strings.NewReader("https://some-link.com")
	resp, err := req.Send()
	assert.NoError(t, err, "error making HTTP request")

	assert.Equal(t, http.StatusCreated, resp.StatusCode(), "Код ответа не совпадает с ожидаемым")
	// проверим корректность полученного тела ответа, если мы его ожидаем
	assert.NotEmpty(t, resp.Body, "Тело ответа не совпадает с ожидаемым")

	handler = http.HandlerFunc(h.Expander)
	srv = httptest.NewServer(handler)
	req = resty.New().R()
	req.Method = http.MethodGet
	req.URL = srv.URL + "/" + "2476"
	resp, err = req.Send()
	assert.NoError(t, err, "error making HTTP request")

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode(), "Код ответа не совпадает с ожидаемым")
}
