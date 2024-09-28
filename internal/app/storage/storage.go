package storage

import (
	"github.com/alexeyvilmost/urlshort.git/internal/app/config"
	"github.com/go-errors/errors"
	_ "github.com/jackc/pgx/v5/stdlib"
)

var ErrDuplicateValue = errors.New("addition attempt failed: key value already exists")
var ErrExistingFullURL = errors.New("addition attempt failed: full url already exists")
var ErrNoValue = errors.New("no value presented for this shortUrl")

const (
	LocalMode string = "Local"
	FileMode  string = "File"
	DBMode    string = "DB"
)

type StorageI interface {
	CheckDBConn() bool
	Get(shortURL string) (string, error)
	GetByUser(shortURL, userID string) (string, error)
	GetUserURLs(userID string) ([]UserURLs, error)
	Add(userID, shortURL, fullURL string) (string, error)
}
type UserURLs struct {
	OriginalURL string `json:"original_url"`
	ShortURL    string `json:"short_url"`
}

func NewStorage(config *config.Config) (StorageI, error) {
	if len(config.DBString) != 0 {
		return NewDBStorage(config)
	} else if len(config.StorageFile) != 0 {
		return NewFileStorage(config)
	} else {
		return NewLocalStorage()
	}
}
