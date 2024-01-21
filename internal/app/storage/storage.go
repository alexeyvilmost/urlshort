package storage

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"os"

	"github.com/alexeyvilmost/urlshort.git/internal/app/config"
	"github.com/go-errors/errors"
	pgerrcode "github.com/jackc/pgerrcode"
	_ "github.com/jackc/pgx/v5/stdlib"
	pq "github.com/lib/pq"
	"github.com/rs/zerolog/log"
)

var ErrDuplicateValue = errors.New("Addition attempt failed: key value already exists")
var ErrExistingFullURL = errors.New("Addition attempt failed: full url already exists")

const (
	LocalMode int = iota
	FileMode
	DBMode
)

type Storage struct {
	container map[string]string
	file      *os.File
	DBString  string
	mode      int
}

func NewStorage(config *config.Config) (*Storage, error) {
	var file *os.File
	var err error
	var mode int
	if len(config.DBString) != 0 {
		db, err := sql.Open("pgx", config.DBString)
		if err != nil {
			return &Storage{}, fmt.Errorf("failed to create db from connection string: %w", err)
		}
		defer db.Close()
		_, err = db.Exec("CREATE TABLE ulrs (short TEXT UNIQUE, full TEXT, PRIMARY KEY short);")
		if err != nil {
			return &Storage{}, fmt.Errorf("failed to create table in db: %w", err)
		}
		_, err = db.Exec("CREATE UNIQUE INDEX full_index ON urls (full);")
		if err != nil {
			return &Storage{}, fmt.Errorf("failed to create index in db: %w", err)
		}
		mode = DBMode
	} else if len(config.StorageFile) != 0 {
		file, err = os.Create(config.StorageFile)
		if err != nil {
			return &Storage{}, fmt.Errorf("failed to create file for storage: %w", err)
		}
		mode = FileMode
	} else {
		mode = LocalMode
	}

	result := &Storage{
		container: map[string]string{},
		file:      file,
		DBString:  config.DBString,
		mode:      mode,
	}
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

func (s *Storage) Get(shortURL string) (string, bool) {
	switch s.mode {
	case LocalMode:
		result, ok := s.container[shortURL]
		return result, ok
	case FileMode:
		reader := csv.NewReader(s.file)
		records, err := reader.ReadAll()
		if err != nil {
			log.Error().Err(err).Msg("failed to read file")
			return "", false
		}
		for _, record := range records {
			short := record[0]
			if short == shortURL {
				full := record[1]
				return full, true
			}
		}
		return "", false
	case DBMode:
		db, err := sql.Open("pgx", s.DBString)
		if err != nil {
			log.Error().Err(err).Msg("failed to create db for storage")
			return "", false
		}
		row := db.QueryRow("SELECT full FROM urls WHERE short = $1;", shortURL)
		var desc sql.NullString

		err = row.Scan(&desc)
		if err != nil {
			log.Error().Err(err).Msg("failed to parse db response")
			return "", false
		}
		if desc.Valid {
			return desc.String, true
		}
	}
	log.Error().Msg("unsupported storage mode")
	return "", false
}

func (s *Storage) Add(shortURL, fullURL string) (string, error) {
	_, ok := s.Get(shortURL)
	if ok {
		return "", ErrDuplicateValue
	}
	switch s.mode {
	case LocalMode:
		s.container[shortURL] = fullURL
		return "", nil
	case FileMode:
		writer := csv.NewWriter(s.file)
		defer writer.Flush()
		err := writer.Write([]string{shortURL, fullURL})
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
		_, err = db.Exec("INSERT INTO urls VALUES ($1, $2) ON CONFLICT;", shortURL, fullURL)
		if err != nil {
			pqErr, ok := err.(*pq.Error)
			if ok && pqErr.Code == pgerrcode.UniqueViolation {
				row := db.QueryRow("SELECT full FROM urls WHERE full = $1;", fullURL)
				var desc sql.NullString

				err = row.Scan(&desc)
				if err != nil {
					log.Error().Err(err).Msg("failed to parse db response")
					return "", err
				}
				if desc.Valid {
					return desc.String, ErrExistingFullURL
				}
			}
			log.Error().Err(err).Msg("failed to insert value in db")
			return "", err
		}
		return "", nil
	}
	log.Error().Msg("unsupported storage mode")
	return "", nil
}
