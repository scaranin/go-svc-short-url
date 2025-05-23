package storage

import (
	"context"
	"log"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/scaranin/go-svc-short-url/internal/models"
)

type DBStorageInterface interface {
	Save(URL *models.URL) (string, error)
	Load(shortURL string) (string, error)
	Ping(ctx context.Context) error
	GetUserURLList(UserID string) ([]models.URLUserList, error)
	Close()
}

type DBStorage struct {
	DSN     string
	PGXPool *pgxpool.Pool
}

func (dbStore DBStorage) Save(URL *models.URL) (string, error) {
	ctx := context.Background()
	_, err := dbStore.PGXPool.Exec(ctx, "INSERT INTO MAP_URL(correlation_id, short_url, original_url, user_id) VALUES (@P_CORR_ID, @P_SHORT_URL, @P_ORIGINAL_URL, @P_USER_ID)",
		pgx.NamedArgs{"@P_CORR_ID": URL.CorrelationID, "P_SHORT_URL": URL.ShortURL, "P_ORIGINAL_URL": URL.OriginalURL, "P_USER_ID": URL.UserID},
	)
	if pgErr, ok := err.(*pgconn.PgError); ok {
		if pgErr.Code == pgerrcode.UniqueViolation {
			return URL.ShortURL, pgErr
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

func (dbStore DBStorage) GetUserURLList(UserID string) ([]models.URLUserList, error) {
	ctx := context.Background()
	rows, err := dbStore.PGXPool.Query(ctx, "select short_url, original_url from MAP_URL WHERE user_id = @P_USER_ID",
		pgx.NamedArgs{"P_USER_ID": UserID},
	)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var URLlist []models.URLUserList

	for rows.Next() {
		var URLlistItem models.URLUserList

		err = rows.Scan(&URLlistItem.ShortURL, &URLlistItem.OriginalURL)
		if err != nil {
			log.Fatal(err)
		}
		URLlist = append(URLlist, URLlistItem)
	}

	if err = rows.Err(); err != nil {
		log.Fatal(err)
	}

	if err != nil {
		return URLlist, err
	}

	return URLlist, err
}

func (dbStore DBStorage) CreateDBScheme(ctx context.Context) error {
	_, err := dbStore.PGXPool.Exec(ctx, `CREATE TABLE MAP_URL (
		"correlation_id" TEXT,
        "short_url" TEXT,
		"original_url" TEXT,
		"user_id" TEXT
      )`)
	if pgErr, ok := err.(*pgconn.PgError); ok {
		if pgErr.Code == pgerrcode.DuplicateTable {
			err = nil
		}
	}
	if err == nil {
		_, err = dbStore.PGXPool.Exec(ctx, `CREATE UNIQUE INDEX idx_original_url ON MAP_URL(original_url)`)
		if pgErr, ok := err.(*pgconn.PgError); ok {
			if pgErr.Code == pgerrcode.DuplicateTable {
				err = nil
			}
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
