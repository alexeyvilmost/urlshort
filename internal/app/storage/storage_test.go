package storage

import (
	"os"
	"testing"
)

func TestStorage_Add(t *testing.T) {
	type fields struct {
		container map[string]string
		file      *os.File
	}
	type args struct {
		shortURL string
		fullURL  string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Storage{
				container: tt.fields.container,
				file:      tt.fields.file,
			}
			if err := s.Add(tt.args.shortURL, tt.args.fullURL); (err != nil) != tt.wantErr {
				t.Errorf("Storage.Add() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
