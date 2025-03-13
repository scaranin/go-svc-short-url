package storage

import (
	"context"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/scaranin/go-svc-short-url/internal/models"
)

type DBStorage struct {
	DSN     string
	PGXPool *pgxpool.Pool
}

func (dbStore DBStorage) Save(URL *models.URL) (string, error) {
	ctx := context.Background()
	_, err := dbStore.PGXPool.Exec(ctx, "INSERT INTO MAP_URL(correlation_id, short_url, original_url) VALUES (@P_CORR_ID, @P_SHORT_URL, @P_ORIGINAL_URL)",
		pgx.NamedArgs{"@P_CORR_ID": URL.CorrelationID, "P_SHORT_URL": URL.ShortURL, "P_ORIGINAL_URL": URL.OriginalURL},
	)
	if pgErr, ok := err.(*pgconn.PgError); ok {
		if pgErr.Code == pgerrcode.UniqueViolation {
			shortUrl, _ := dbStore.Load(URL.ShortURL)
			return shortUrl, pgErr
		}
	}
	return URL.ShortURL, err
}

func (dbStore DBStorage) Load(shortURL string) (string, error) {
	ctx := context.Background()
	row := dbStore.PGXPool.QueryRow(ctx, "select original_url from MAP_URL WHERE short_url = @P_SHORT_URL",
		pgx.NamedArgs{"P_SHORT_URL": shortURL},
	)

	var originalURL string
	err := row.Scan(&originalURL)
	if err != nil {
		return originalURL, err
	}
	return originalURL, err
}

func (dbStore DBStorage) Ping(ctx context.Context) error {
	return dbStore.PGXPool.Ping(ctx)
}

func (dbStore DBStorage) CreateDBScheme(ctx context.Context) error {
	_, err := dbStore.PGXPool.Exec(ctx, `CREATE TABLE MAP_URL (
		"correlation_id" TEXT,
        "short_url" TEXT,
		"original_url" TEXT
      )`)
	if pgErr, ok := err.(*pgconn.PgError); ok {
		if pgErr.Code == pgerrcode.DuplicateTable {
			err = nil
		}
	}
	return err

}

func CreateStoreDB(DSN string) (DBStorage, error) {
	var dbStore DBStorage
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, DSN)
	if err != nil {
		return dbStore, err
	}
	dbStore.PGXPool = pool

	err = dbStore.Ping(ctx)
	if err != nil {
		return dbStore, err
	}

	err = dbStore.CreateDBScheme(ctx)
	if err != nil {
		return dbStore, err
	}

	return dbStore, err
}

func (dbStore *DBStorage) Close() {
	dbStore.PGXPool.Close()
}
