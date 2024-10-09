package storage

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/alexeyvilmost/urlshort.git/internal/app/config"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/rs/zerolog/log"
)

type DBStorage struct {
	db *sql.DB
}

func NewDBStorage(config *config.Config) (*DBStorage, error) {
	db, err := sql.Open("pgx", config.DBString)
	if err != nil {
		return &DBStorage{}, fmt.Errorf("failed to create db from connection string: %w", err)
	}
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS urls (short_url TEXT UNIQUE, full_url TEXT, user_id UUID, is_deleted BOOL, PRIMARY KEY (short_url), UNIQUE (full_url, user_id));")
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
		db: db,
	}
	return result, nil
}

func (s *DBStorage) CheckDBConn(ctx context.Context) bool {
	conn, err := s.db.Conn(ctx)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func (s *DBStorage) Get(ctx context.Context, shortURL string) (string, error) {
	row, err := s.db.QueryContext(ctx, "SELECT full_url, is_deleted FROM urls WHERE short_url = $1;", shortURL)
	if err != nil {
		return "", err
	}
	var result string
	var deleted bool

	err = row.Scan(&result, &deleted)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", ErrNoValue
		}
		log.Error().Err(err).Msg("get: failed to parse db response")
		return "", err
	}
	if deleted {
		return "", ErrGone
	}
	return result, nil
}

func (s *DBStorage) GetByUser(ctx context.Context, shortURL, userID string) (string, error) {
	row, err := s.db.QueryContext(ctx, "SELECT full_url, is_deleted FROM urls WHERE short_url = $1 AND user_id = $2;", shortURL, userID)
	if err != nil {
		return "", err
	}
	var result string
	var deleted bool

	err = row.Scan(&result, &deleted)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", ErrNoValue
		}
		log.Error().Err(err).Msg("get: failed to parse db response")
		return "", err
	}
	if deleted {
		return "", ErrGone
	}
	return result, nil
}

func (s *DBStorage) GetUserURLs(ctx context.Context, userID string) ([]UserURLs, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT short_url, full_url FROM urls WHERE user_id = $1 AND is_deleted != TRUE;", userID)
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

func (s *DBStorage) Add(ctx context.Context, userID, shortURL, fullURL string) (string, error) {
	_, err := s.GetByUser(ctx, shortURL, userID)
	switch err {
	case nil:
		return "", ErrDuplicateValue
	case ErrNoValue:
		// pass
	default:
		return "", err
	}
	row, err := s.db.QueryContext(ctx, "INSERT INTO urls VALUES ($1, $2, $3, FALSE) ON CONFLICT DO NOTHING RETURNING short_url;", shortURL, fullURL, userID)
	if err != nil {
		return "", err
	}
	var str string
	err = row.Scan(&str)
	if err != nil {
		// if nothing return, value already presented
		if err == sql.ErrNoRows {
			log.Info().Msg("searching for full_url: " + fullURL)
			row := s.db.QueryRow("SELECT short_url FROM urls WHERE full_url = $1 AND user_id = $2;", fullURL, userID)
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

func (s *DBStorage) DeleteURLs(ctx context.Context, userID string, shortURLs []string) {
	query := "UPDATE urls SET is_deleted = TRUE WHERE user_id ='" + userID + "' AND short_url IN ('" + strings.Join(shortURLs, "','") + "');"
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil || rows.Err() != nil {
		log.Error().Err(err).Err(rows.Err()).Msg("Can't delete URLs")
	}
}
