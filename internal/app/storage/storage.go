package storage

import "github.com/go-errors/errors"

var ErrDuplicateValue = errors.New("Addition attempt failed: key value already exists")

type Storage struct {
	container map[string]string
}

func NewStorage() *Storage {
	result := &Storage{
		container: map[string]string{},
	}
	return result
}

func (s *Storage) Add(shortURL, fullURL string) error {
	_, ok := s.container[shortURL]
	if ok {
		return ErrDuplicateValue
	}
	s.container[shortURL] = fullURL
	return nil
}

func (s *Storage) Get(shortURL string) string {
	return s.container[shortURL]
}
