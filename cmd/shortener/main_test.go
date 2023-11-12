package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

func Test_happypath(t *testing.T) {
	// testCases := []struct {
	// 	method       string
	// 	expectedCode int
	// 	requestBody  io.Reader
	// 	expectedBody string
	// }{
	// 	{method: http.MethodGet, expectedCode: http.StatusBadRequest, requestBody: nil, expectedBody: ""},
	// 	{method: http.MethodPost, expectedCode: http.StatusCreated, requestBody: strings.NewReader("https://some-link.com"), expectedBody: ""},
	// }

	handler := http.HandlerFunc(balancer)

	srv := httptest.NewServer(handler)
	defer srv.Close()
	// for _, tc := range testCases {

	req := resty.New().R()
	req.Method = http.MethodPost
	req.URL = srv.URL
	req.Body = strings.NewReader("https://some-link.com")
	resp, err := req.Send()
	assert.NoError(t, err, "error making HTTP request")

	assert.Equal(t, http.StatusCreated, resp.StatusCode(), "Код ответа не совпадает с ожидаемым")
	// проверим корректность полученного тела ответа, если мы его ожидаем
	assert.NotEmpty(t, resp.Body, "Тело ответа не совпадает с ожидаемым")

	req = resty.New().R()
	req.Method = http.MethodGet
	req.URL = srv.URL + "/" + "2476"
	resp, err = req.Send()
	assert.NoError(t, err, "error making HTTP request")

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode(), "Код ответа не совпадает с ожидаемым")
}
