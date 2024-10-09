package storage

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"

	"github.com/alexeyvilmost/urlshort.git/internal/app/config"
	"github.com/go-errors/errors"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/rs/zerolog/log"
)

type FileStorage struct {
	filename string
}

func NewFileStorage(config *config.Config) (*FileStorage, error) {
	var file *os.File
	var err error
	log.Debug().Msg("FileString: " + config.StorageFile)
	file, err = os.Create(config.StorageFile)
	if err != nil {
		return &FileStorage{}, fmt.Errorf("failed to create file for storage: %w", err)
	}
	defer file.Close()

	result := &FileStorage{
		filename: config.StorageFile,
	}
	return result, nil
}

func (s *FileStorage) CheckDBConn(context.Context) bool {
	return true
}

func (s *FileStorage) Get(_ context.Context, shortURL string) (string, error) {
	file, err := os.Open(s.filename)
	if err != nil {
		log.Error().Err(err).Msg("failed to open file")
		return "", err
	}
	defer file.Close()
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		log.Error().Err(err).Msg("failed to read file")
		return "", err
	}
	for _, record := range records {
		short := record[0]
		log.Info().Msg("Searching for " + shortURL + ", checking " + short)
		if short == shortURL {
			full := record[1]
			return full, nil
		}
	}
	return "", ErrNoValue
}

func (s *FileStorage) GetByUser(_ context.Context, shortURL, userID string) (string, error) {
	file, err := os.Open(s.filename)
	if err != nil {
		log.Error().Err(err).Msg("failed to open file")
		return "", err
	}
	defer file.Close()
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		log.Error().Err(err).Msg("failed to read file")
		return "", err
	}
	for _, record := range records {
		short := record[0]
		log.Info().Msg("Searching for " + shortURL + ", checking " + short)
		if short == shortURL && userID == record[2] {
			full := record[1]
			return full, nil
		}
	}
	return "", ErrNoValue
}

func (s *FileStorage) GetUserURLs(_ context.Context, userID string) ([]UserURLs, error) {
	return nil, errors.New("user urls only supported in db storage mode")
}

func (s *FileStorage) Add(ctx context.Context, userID, shortURL, fullURL string) (string, error) {
	log.Info().Msg("UserID:" + userID)
	_, err := s.GetByUser(ctx, shortURL, userID)
	switch err {
	case nil:
		return "", ErrDuplicateValue
	case ErrNoValue:
		// pass
	default:
		return "", err
	}
	file, err := os.OpenFile(s.filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Error().Err(err).Msg("failed to open file")
		return "", err
	}
	defer file.Close()
	writer := csv.NewWriter(file)
	defer writer.Flush()
	err = writer.Write([]string{shortURL, fullURL, userID})
	if err != nil {
		log.Error().Err(err).Msg("failed to write in file")
		return "", err
	}
	return "", nil
}

func (s *FileStorage) DeleteURLs(_ context.Context, userID string, shortURLs []string) {
	log.Info().Msg("Not supported in file storage")
}
