package storage

import (
	"database/sql"
	"fmt"

	"github.com/alexeyvilmost/urlshort.git/internal/app/config"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/rs/zerolog/log"
)

type DBStorage struct {
	DBString string
}

func NewDBStorage(config *config.Config) (*DBStorage, error) {
	db, err := sql.Open("pgx", config.DBString)
	if err != nil {
		return &DBStorage{}, fmt.Errorf("failed to create db from connection string: %w", err)
	}
	defer db.Close()
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS urls (short_url TEXT UNIQUE, full_url TEXT, user_id UUID, PRIMARY KEY (short_url), UNIQUE (full_url, user_id));")
	if err != nil {
		return &DBStorage{}, fmt.Errorf("failed to create table in db: %w", err)
	}
	_, err = db.Exec("CREATE INDEX IF NOT EXISTS full_index ON urls (full_url);")
	if err != nil {
		return &DBStorage{}, fmt.Errorf("failed to create index in db: %w", err)
	}
	_, err = db.Exec("CREATE INDEX IF NOT EXISTS user_index ON urls (user_id);")
	if err != nil {
		return &DBStorage{}, fmt.Errorf("failed to create index in db: %w", err)
	}
	result := &DBStorage{
		DBString: config.DBString,
	}
	return result, nil
}

func (s *DBStorage) CheckDBConn() bool {
	db, err := sql.Open("pgx", s.DBString)
	if err != nil {
		log.Error().Err(err).Msg("failed to create db for storage")
		return false
	}
	defer db.Close()
	return true
}

func (s *DBStorage) Get(shortURL string) (string, error) {
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

func (s *DBStorage) GetByUser(shortURL, userID string) (string, error) {
	db, err := sql.Open("pgx", s.DBString)
	if err != nil {
		log.Error().Err(err).Msg("failed to create db for storage")
		return "", ErrNoValue
	}
	row := db.QueryRow("SELECT full_url FROM urls WHERE short_url = $1 AND user_id = $2;", shortURL, userID)
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

func (s *DBStorage) GetUserURLs(userID string) ([]UserURLs, error) {
	db, err := sql.Open("pgx", s.DBString)
	if err != nil {
		log.Error().Err(err).Msg("failed to open db for storage")
		return nil, ErrNoValue
	}
	rows, err := db.Query("SELECT short_url, full_url FROM urls WHERE user_id = $1;", userID)
	if err != nil {
		return nil, err
	} else if rows.Err() != nil {
		return nil, rows.Err()
	}
	var result []UserURLs

	for rows.Next() {
		var url UserURLs
		if err := rows.Scan(&url.ShortURL, &url.OriginalURL); err != nil {
			return result, err
		}
		result = append(result, url)
	}
	return result, nil
}

func (s *DBStorage) Add(userID, shortURL, fullURL string) (string, error) {
	_, err := s.GetByUser(shortURL, userID)
	switch err {
	case nil:
		return "", ErrDuplicateValue
	case ErrNoValue:
		// pass
	default:
		return "", err
	}
	db, err := sql.Open("pgx", s.DBString)
	if err != nil {
		log.Error().Err(err).Msg("failed to create db for storage")
		return "", err
	}
	row := db.QueryRow("INSERT INTO urls VALUES ($1, $2, $3) ON CONFLICT DO NOTHING RETURNING short_url;", shortURL, fullURL, userID)
	var str string
	err = row.Scan(&str)
	if err != nil {
		// if nothing return, value already presented
		if err == sql.ErrNoRows {
			log.Info().Msg("searching for full_url: " + fullURL)
			row := db.QueryRow("SELECT short_url FROM urls WHERE full_url = $1 AND user_id = $2;", fullURL, userID)
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
