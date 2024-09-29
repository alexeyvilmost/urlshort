package storage

import (
	"sync"

	"github.com/go-errors/errors"
)

type LocalStorage struct {
	container map[string]map[string]string
	mut       sync.RWMutex
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
			if result == "-" {
				return "", ErrGone
			}
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
		if result == "-" {
			return "", ErrGone
		}
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
	case ErrGone:
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

func (s *LocalStorage) deleteURL(userID, shortURL string) {
	s.mut.RLock()
	_, ok := s.container[userID]
	if !ok {
		s.container[userID] = map[string]string{}
		return
	}
	_, ok = s.container[userID][shortURL]
	s.mut.RUnlock()
	if ok {
		s.mut.Lock()
		s.container[userID][shortURL] = "-"
		s.mut.Unlock()
	}
}

func (s *LocalStorage) DeleteURLs(userID string, shortURLs []string) {
	for _, url := range shortURLs {
		go s.deleteURL(userID, url)
	}
}
