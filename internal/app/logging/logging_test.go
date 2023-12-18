package logging

import (
	"net/http"
	"reflect"
	"testing"
)

func TestWithLogging(t *testing.T) {
	type args struct {
		h http.HandlerFunc
	}
	tests := []struct {
		name string
		args args
		want http.HandlerFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := WithLogging(tt.args.h); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WithLogging() = %v, want %v", got, tt.want)
			}
		})
	}
}
