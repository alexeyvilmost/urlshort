package storage

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"os"

	"github.com/alexeyvilmost/urlshort.git/internal/app/config"
	"github.com/go-errors/errors"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/rs/zerolog/log"
)

var ErrDuplicateValue = errors.New("addition attempt failed: key value already exists")
var ErrExistingFullURL = errors.New("addition attempt failed: full url already exists")
var ErrNoValue = errors.New("no value presented for this shortUrl")

const (
	LocalMode string = "Local"
	FileMode  string = "File"
	DBMode    string = "DB"
)

type Storage struct {
	container map[string]string
	filename  string
	DBString  string
	mode      string
}

func NewStorage(config *config.Config) (*Storage, error) {
	var file *os.File
	var err error
	var mode string
	log.Debug().Msg("FileString: " + config.StorageFile)
	if len(config.DBString) != 0 {
		db, err := sql.Open("pgx", config.DBString)
		if err != nil {
			return &Storage{}, fmt.Errorf("failed to create db from connection string: %w", err)
		}
		defer db.Close()
		_, err = db.Exec("CREATE TABLE IF NOT EXISTS urls (short_url TEXT UNIQUE, full_url TEXT, PRIMARY KEY (short_url));")
		if err != nil {
			return &Storage{}, fmt.Errorf("failed to create table in db: %w", err)
		}
		_, err = db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS full_index ON urls (full_url);")
		if err != nil {
			return &Storage{}, fmt.Errorf("failed to create index in db: %w", err)
		}
		mode = DBMode
	} else if len(config.StorageFile) != 0 {
		file, err = os.Create(config.StorageFile)
		if err != nil {
			return &Storage{}, fmt.Errorf("failed to create file for storage: %w", err)
		}
		defer file.Close()
		mode = FileMode
	} else {
		mode = LocalMode
	}

	result := &Storage{
		container: map[string]string{},
		filename:  config.StorageFile,
		DBString:  config.DBString,
		mode:      mode,
	}
	log.Debug().Msg("Choosen mode: " + mode)
	return result, nil
}

func (s *Storage) CheckDBConn() bool {
	if s.mode != DBMode {
		log.Error().Msg("no connection string presented")
	}
	db, err := sql.Open("pgx", s.DBString)
	if err != nil {
		log.Error().Err(err).Msg("failed to create db for storage")
		return false
	}
	defer db.Close()
	return true
}

func (s *Storage) Get(shortURL string) (string, error) {
	switch s.mode {
	case LocalMode:
		result, ok := s.container[shortURL]
		if !ok {
			return "", ErrNoValue
		}
		return result, nil
	case FileMode:
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
	case DBMode:
		db, err := sql.Open("pgx", s.DBString)
		if err != nil {
			log.Error().Err(err).Msg("failed to create db for storage")
			return "", ErrNoValue
		}
		row := db.QueryRow("SELECT full_url FROM urls WHERE short_url = $1;", shortURL)
		var result string

		err = row.Scan(&result)
		if err != nil {
			if err == sql.ErrNoRows {
				return "", ErrNoValue
			}
			log.Error().Err(err).Msg("get: failed to parse db response")
			return "", err
		}
		return result, nil
	}
	log.Error().Msg("unsupported storage mode")
	return "", errors.New("unsupported storage mode")
}

func (s *Storage) Add(shortURL, fullURL string) (string, error) {
	_, err := s.Get(shortURL)
	switch err {
	case nil:
		return "", ErrDuplicateValue
	case ErrNoValue:
		// pass
	default:
		return "", err
	}
	switch s.mode {
	case LocalMode:
		s.container[shortURL] = fullURL
		return "", nil
	case FileMode:
		file, err := os.OpenFile(s.filename, os.O_WRONLY, 0666)
		if err != nil {
			log.Error().Err(err).Msg("failed to open file")
			return "", err
		}
		defer file.Close()
		writer := csv.NewWriter(file)
		defer writer.Flush()
		err = writer.Write([]string{shortURL, fullURL})
		log.Info().Msg("data written: " + shortURL + ", " + fullURL)
		if err != nil {
			log.Error().Err(err).Msg("failed to write in file")
			return "", err
		}
		return "", nil
	case DBMode:
		db, err := sql.Open("pgx", s.DBString)
		if err != nil {
			log.Error().Err(err).Msg("failed to create db for storage")
			return "", err
		}
		row := db.QueryRow("INSERT INTO urls VALUES ($1, $2) ON CONFLICT (full_url) DO NOTHING RETURNING short_url;", shortURL, fullURL)
		var str string
		err = row.Scan(&str)
		if err != nil {
			// if nothing return, value already presented
			if err == sql.ErrNoRows {
				log.Info().Msg("searching for full_url: " + fullURL)
				row := db.QueryRow("SELECT short_url FROM urls WHERE full_url = $1;", fullURL)
				var result string

				err = row.Scan(&result)
				if err != nil {
					log.Error().Err(err).Msg("add: failed to parse db response")
					return "", err
				}
				return result, ErrExistingFullURL
			}
			log.Error().Err(err).Msg("failed to insert value in db")
			return "", err
		}
		return "", nil
	}
	log.Error().Msg("unsupported storage mode")
	return "", nil
}
