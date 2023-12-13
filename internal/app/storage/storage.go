package storage

import (
	"os"

	"github.com/go-errors/errors"
	"github.com/rs/zerolog/log"
)

var ErrDuplicateValue = errors.New("Addition attempt failed: key value already exists")

type Storage struct {
	container map[string]string
	file      *os.File
}

func NewStorage(filename string) *Storage {
	file, err := os.Create(filename)
	if err == nil {
		result := &Storage{
			container: map[string]string{},
			file:      file,
		}
		return result
	} else {
		log.Error().Msg(err.Error())
	}
	return &Storage{}
}

func (s *Storage) Add(shortURL, fullURL string) error {
	_, ok := s.container[shortURL]
	if ok {
		return ErrDuplicateValue
	}
	s.container[shortURL] = fullURL
	s.file.WriteString(shortURL + " : " + fullURL)
	return nil
}

func (s *Storage) Get(shortURL string) (string, bool) {
	result, ok := s.container[shortURL]
	return result, ok
}
