package storage

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/alexeyvilmost/urlshort.git/internal/app/config"
	"github.com/go-errors/errors"
	"github.com/rs/zerolog/log"
)

var ErrDuplicateValue = errors.New("Addition attempt failed: key value already exists")

type Storage struct {
	container map[string]string
	file      *os.File
	DBString  string
}

func NewStorage(config *config.Config) (*Storage, error) {
	var file *os.File
	var err error
	if len(config.StorageFile) != 0 {
		file, err = os.Create(config.StorageFile)
	}
	if err != nil {
		return &Storage{}, fmt.Errorf("failed to create file for storage: %w", err)
	}

	result := &Storage{
		container: map[string]string{},
		file:      file,
		DBString:  config.DBString,
	}
	return result, nil
}

func (s *Storage) CheckDBConn() bool {
	db, err := sql.Open("pgx", s.DBString)
	if err != nil {
		log.Error().Err(err).Msg("failed to create db for storage")
		return false
	}
	defer db.Close()
	return true
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
