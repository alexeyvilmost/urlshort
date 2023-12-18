package handlers

import (
	"net/http"
	"testing"

	"github.com/alexeyvilmost/urlshort.git/internal/app/storage"
)

func TestHandlers_Shortener(t *testing.T) {
	type fields struct {
		BaseURL string
		Storage *storage.Storage
	}
	type args struct {
		res http.ResponseWriter
		req *http.Request
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := Handlers{
				BaseURL: tt.fields.BaseURL,
				Storage: tt.fields.Storage,
			}
			h.Shortener(tt.args.res, tt.args.req)
		})
	}
}
