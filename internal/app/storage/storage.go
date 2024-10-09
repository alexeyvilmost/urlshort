package storage

import (
	"context"

	"github.com/alexeyvilmost/urlshort.git/internal/app/config"
	"github.com/go-errors/errors"
	_ "github.com/jackc/pgx/v5/stdlib"
)

var ErrDuplicateValue = errors.New("addition attempt failed: key value already exists")
var ErrExistingFullURL = errors.New("addition attempt failed: full url already exists")
var ErrNoValue = errors.New("no value presented for this shortUrl")
var ErrGone = errors.New("this url was deleted")

type StorageI interface {
	CheckDBConn(ctx context.Context) bool
	Get(ctx context.Context, shortURL string) (string, error)
	GetByUser(ctx context.Context, shortURL, userID string) (string, error)
	GetUserURLs(ctx context.Context, userID string) ([]UserURLs, error)
	Add(ctx context.Context, userID, shortURL, fullURL string) (string, error)
	DeleteURLs(ctx context.Context, userID string, shortURLs []string)
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
