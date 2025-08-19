package handlers

import (
	"crypto/sha1"
	"encoding/base64"

	"github.com/scaranin/go-svc-short-url/internal/auth"
	"github.com/scaranin/go-svc-short-url/internal/config"
	"github.com/scaranin/go-svc-short-url/internal/models"
)

const (
	// contentTypeTextPlain is a constant for the "text/plain" MIME type.
	contentTypeTextPlain string = "text/plain"
	// contentTypeApJSON is a constant for the "application/json" MIME type.
	contentTypeApJSON string = "application/json"
)

// URLHandler is the primary struct that holds the service's dependencies and configuration.
// It orchestrates operations by interacting with storage and authentication components.
type URLHandler struct {
	// URLMap stores URL mappings in memory.
	// Deprecated: Its usage is generally superseded by the Storage field for persistent logic.
	URLMap map[string]string
	// BaseURL is the prefix for all generated short URLs returned to the client.
	BaseURL string
	// FileProducer is a pointer to a model responsible for writing data to a file.
	FileProducer *models.Producer
	// DSN is the Data Source Name, a connection string for the database.
	DSN string
	// Storage provides an interface for interacting with the persistence layer.
	Storage models.Storage
	// Auth holds authentication-related configuration and state, such as the
	// current user's ID, derived from request context.
	Auth auth.AuthConfig
}

// CreateHandle initializes and returns a new URLHandler instance.
// It is configured with application settings, a storage backend, and authentication configuration.
func CreateHandle(cfg config.ShortenerConfig, store models.Storage, auth auth.AuthConfig) URLHandler {
	var h URLHandler
	h.BaseURL = cfg.BaseURL
	h.Storage = store
	h.DSN = cfg.DSN
	h.Auth = auth
	return h
}

// ShortURLCalc computes a short URL identifier from an original URL.
// It uses SHA1 hashing and Base64 encoding to generate a concise string.
func ShortURLCalc(originalURL string) string {
	hasher := sha1.New()
	hasher.Write([]byte(originalURL))
	return base64.URLEncoding.EncodeToString(hasher.Sum(nil))
}

// Save adds a new record to the storage. It associates the URL with the
// user ID stored in the handler's Auth field.
// It calculates the short URL, creates the URL model, and passes it to the storage layer.
func (h *URLHandler) Save(originalURL string, correlationID string) (string, error) {
	shortURL := ShortURLCalc(originalURL)
	var baseURL = models.URL{
		CorrelationID: correlationID,
		OriginalURL:   originalURL,
		ShortURL:      shortURL,
		UserID:        h.Auth.UserID,
	}
	shortURL, err := h.Storage.Save(&baseURL)
	return shortURL, err
}

// Load retrieves the original URL from storage using its short URL identifier.
// It delegates the call to the Load method of the configured Storage.
func (h *URLHandler) Load(shortURL string) (string, error) {
	return h.Storage.Load(shortURL)
}
