package storage

import (
	"fmt"
	"os"

	"github.com/go-errors/errors"
)

var ErrDuplicateValue = errors.New("Addition attempt failed: key value already exists")

type Storage struct {
	container map[string]string
	file      *os.File
}

func NewStorage(filename string) (*Storage, error) {
	var file *os.File
	var err error
	if len(filename) != 0 {
		file, err = os.Create(filename)
	}
	if err != nil {
		return &Storage{}, fmt.Errorf("failed to create file for storage: %w", err)
	}
	result := &Storage{
		container: map[string]string{},
		file:      file,
	}
	return result, nil
}

func (s *Storage) Add(shortURL, fullURL string) error {
	_, ok := s.container[shortURL]
	if ok {
		return ErrDuplicateValue
	}
	s.container[shortURL] = fullURL
	if s.file != nil {
		_, err := s.file.WriteString(shortURL + " : " + fullURL)
		return err
	}
	return nil
}

func (s *Storage) Get(shortURL string) (string, bool) {
	result, ok := s.container[shortURL]
	return result, ok
}
