package storage

import (
	"github.com/go-errors/errors"
)

type LocalStorage struct {
	container map[string]map[string]string
}

func NewLocalStorage() (*LocalStorage, error) {
	result := &LocalStorage{
		container: map[string]map[string]string{},
	}
	return result, nil
}

func (s *LocalStorage) CheckDBConn() bool {
	return false
}

func (s *LocalStorage) Get(shortURL string) (string, error) {
	for _, user := range s.container {
		result, ok := user[shortURL]
		if ok {
			return result, nil
		}
	}
	return "", ErrNoValue
}

func (s *LocalStorage) GetByUser(shortURL, userID string) (string, error) {
	_, ok := s.container[userID]
	if !ok {
		s.container[userID] = map[string]string{}
	}
	result, ok := s.container[userID][shortURL]
	if ok {
		return result, nil
	}
	return "", ErrNoValue
}

func (s *LocalStorage) GetUserURLs(userID string) ([]UserURLs, error) {
	return nil, errors.New("user urls only supported in db storage mode")
}

func (s *LocalStorage) Add(userID, shortURL, fullURL string) (string, error) {
	_, err := s.GetByUser(shortURL, userID)
	switch err {
	case nil:
		return "", ErrDuplicateValue
	case ErrNoValue:
		// pass
	default:
		return "", err
	}

	_, ok := s.container[userID]
	if !ok {
		s.container[userID] = map[string]string{}
	}
	s.container[userID][shortURL] = fullURL
	return "", nil
}
