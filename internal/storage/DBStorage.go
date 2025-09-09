package storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/scaranin/go-svc-short-url/internal/models"
)

// DBStorageInterface defines the contract for a database-backed storage system.
// It is intended to be implemented by structs that interact with a database,
// but it is currently not fully utilized as the concrete type DBStorage is returned directly.
type DBStorageInterface interface {
	Save(URL *models.URL) (string, error)
	Load(shortURL string) (string, error)
	Ping(ctx context.Context) error
	GetUserURLList(UserID string) ([]models.URLUserList, error)
	Close()
}

// DBStorage provides a PostgreSQL-backed implementation of the models.Storage interface.
// It manages a connection pool to the database for all storage operations.
type DBStorage struct {
	// DSN is the Data Source Name for the PostgreSQL connection.
	DSN string
	// PGXPool is the active connection pool to the database.
	PGXPool *pgxpool.Pool
}

// Save inserts a new URL record into the `MAP_URL` table.
// It includes the user's ID and sets the `is_deleted` flag to false.
// It handles unique constraint violations on `original_url` by returning the
// conflicting short URL and a specific `pgconn.PgError`, allowing the caller
// to manage conflicts (e.g., by returning an HTTP 409 status).
func (dbStore DBStorage) Save(URL *models.URL) (string, error) {
	ctx := context.Background()
	_, err := dbStore.PGXPool.Exec(ctx, "INSERT INTO MAP_URL(correlation_id, short_url, original_url, user_id, is_deleted) VALUES (@P_CORR_ID, @P_SHORT_URL, @P_ORIGINAL_URL, @P_USER_ID, false)",
		pgx.NamedArgs{"P_CORR_ID": URL.CorrelationID, "P_SHORT_URL": URL.ShortURL, "P_ORIGINAL_URL": URL.OriginalURL, "P_USER_ID": URL.UserID},
	)
	if pgErr, ok := err.(*pgconn.PgError); ok {
		if pgErr.Code == pgerrcode.UniqueViolation {
			return URL.ShortURL, pgErr
		}
	}
	return URL.ShortURL, err
}

// Load retrieves the original URL from the database.
// It also checks if the URL has been marked as deleted. If the `is_deleted` flag
// is true, it returns a specific sentinel error `errors.New("ROW_IS_DELETED")`,
// which allows the caller (handler) to return an HTTP 410 Gone status.
func (dbStore DBStorage) Load(shortURL string) (string, error) {
	ctx := context.Background()
	row := dbStore.PGXPool.QueryRow(ctx, "select original_url, is_deleted from MAP_URL WHERE short_url = @P_SHORT_URL",
		pgx.NamedArgs{"P_SHORT_URL": shortURL},
	)

	var originalURL string
	isDeleted := false
	err := row.Scan(&originalURL, &isDeleted)
	if err != nil {
		return originalURL, err
	}
	if isDeleted {
		err = errors.New("ROW_IS_DELETED")
	}
	return originalURL, err
}

// Ping verifies the connection to the database is active.
func (dbStore DBStorage) Ping(ctx context.Context) error {
	return dbStore.PGXPool.Ping(ctx)
}

// GetUserURLList fetches all non-deleted URLs associated with a specific UserID.
// It queries the database and populates a slice of `models.URLUserList`.
func (dbStore DBStorage) GetUserURLList(UserID string) ([]models.URLUserList, error) {
	ctx := context.Background()
	rows, err := dbStore.PGXPool.Query(ctx, "select short_url, original_url from MAP_URL WHERE user_id = @P_USER_ID",
		pgx.NamedArgs{"P_USER_ID": UserID},
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var URLlist []models.URLUserList

	for rows.Next() {
		var URLlistItem models.URLUserList

		err = rows.Scan(&URLlistItem.ShortURL, &URLlistItem.OriginalURL)
		if err != nil {
			return nil, err
		}
		URLlist = append(URLlist, URLlistItem)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return URLlist, err
}

// DeleteBulk performs a "soft delete" on a batch of URLs owned by a specific user.
// It sets the `is_deleted` flag to true for the given short URLs. The entire
// operation is performed within a single database transaction for atomicity: either
// all URLs are marked for deletion, or none are if an error occurs.
func (dbStore DBStorage) DeleteBulk(UserID string, ShortURLs []string) error {
	ctx := context.Background()
	tx, err := dbStore.PGXPool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	_, err = tx.Prepare(ctx, "SetIsDeleted", "UPDATE MAP_URL set is_deleted = true where short_url = $1 and user_id = $2")
	if err != nil {
		return err
	}

	for _, ShortURL := range ShortURLs {
		fmt.Println(ShortURL, UserID)
		_, err := tx.Exec(ctx, "SetIsDeleted", ShortURL, UserID)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

// GetStats retrieves storage statistics from the database, including the total number of users
// and the number of unique short URLs.
//
// Returns:
//   - models.Statistic: a struct containing count Users and URLs
//   - error: an error
func (dbStore DBStorage) GetStats() (models.Statistic, error) {
	ctx := context.Background()
	sql_stmt := `SELECT 
    (SELECT COUNT(user_id) FROM users) AS users_count,
    (SELECT COUNT(distinct short_url) FROM map_url) AS map_url_count`
	row := dbStore.PGXPool.QueryRow(ctx, sql_stmt)

	var stat models.Statistic

	err := row.Scan(&stat.Users, &stat.URLs)
	return stat, err
}

// CreateDBScheme sets up the necessary database schema.
// It creates the `MAP_URL` table and a `UNIQUE INDEX` on `original_url`.
// The method is idempotent, meaning it can be run multiple times without causing
// errors if the schema already exists, as it checks for `DuplicateTable` errors.
func (dbStore DBStorage) CreateDBScheme(ctx context.Context) error {
	_, err := dbStore.PGXPool.Exec(ctx, `CREATE TABLE MAP_URL (
		"correlation_id" TEXT,
        "short_url" TEXT,
		"original_url" TEXT,
		"user_id" TEXT,
		"is_deleted" BOOL
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

// CreateStoreDB is a factory function that initializes and returns a new DBStorage instance.
// It establishes a connection pool, pings the database, and ensures the schema is created.
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

// Close gracefully closes the database connection pool.
func (dbStore *DBStorage) Close() {
	if dbStore.PGXPool != nil {
		dbStore.PGXPool.Close()
	}
}
