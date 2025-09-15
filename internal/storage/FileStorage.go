package storage

import (
	"context"
	"io"
	"log"

	"github.com/scaranin/go-svc-short-url/internal/models"
)

// FileStorageJSON provides an implementation of the models.Storage interface that
// uses a JSON file for persistence and an in-memory map for fast lookups.
// It can also be configured to operate in a purely in-memory mode.
type FileStorageJSON struct {
	// Producer handles writing new URL entries to the persistence file.
	Producer *models.Producer
	// Consumer handles reading URL entries from the persistence file during startup.
	Consumer *models.Consumer
	// URLMap serves as a cache or the primary in-memory store for all URL lookups.
	URLMap map[string]string
	// INMemory is a flag that, when true, disables all file writing operations,
	// making the storage ephemeral.
	INMemory bool
}

// Save implements the models.Storage interface. It writes the URL to the
// persistence file (if persistence is enabled) and always adds the URL to the
// internal in-memory map for immediate availability.
func (fs FileStorageJSON) Save(URL *models.URL) (string, error) {
	var err error
	if !fs.INMemory {
		err = fs.Producer.AddURL(URL)
	}
	fs.URLMap[URL.ShortURL] = URL.OriginalURL
	return URL.ShortURL, err
}

// Load implements the models.Storage interface. It retrieves the original URL
// by looking it up in the internal in-memory map. It returns an empty string
// if the short URL is not found.
func (fs FileStorageJSON) Load(shortURL string) (string, error) {
	originalURL := fs.URLMap[shortURL]
	return originalURL, nil
}

// GetUserURLList is a stub implementation to satisfy the models.Storage interface.
// This file-based storage type does not support user-specific data,
// so it always returns an empty list and no error.
func (fs FileStorageJSON) GetUserURLList(UserID string) ([]models.URLUserList, error) {
	var URLList []models.URLUserList
	return URLList, nil
}

// DeleteBulk is a stub implementation to satisfy the models.Storage interface.
// In this file-based storage, URLs are not marked as deleted. This method is
// a no-op (no operation) and always returns nil.
func (fs FileStorageJSON) DeleteBulk(UserID string, ShortURLs []string) error {
	return nil
}

// GetStats returns storage statistics including the number of users
// and the number of URLs stored in fs.URLMap. It never returns an error in the current implementation.
//
// Returns:
//   - models.Statistic: a struct containing Users (number of users) and URLs (number of URLs)
func (fs FileStorageJSON) GetStats() (models.Statistic, error) {
	var stat models.Statistic
	stat.Users = 0
	stat.URLs = len(fs.URLMap)
	return stat, nil
}

// GetDataFromFile reads all URL records from the provided consumer and populates an in-memory map.
// It is a helper function used during initialization to load existing data from a file.
func GetDataFromFile(consumer *models.Consumer) map[string]string {
	urlMap := make(map[string]string)
	for {
		mURL, err := consumer.GetURL()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		urlMap[mURL.ShortURL] = mURL.OriginalURL
	}
	return urlMap
}

// CreateStoreFile is a constructor that initializes a FileStorageJSON.
// If fileStoragePath is an empty string, it returns an in-memory-only store.
// Otherwise, it sets up file-based persistence by creating a producer and consumer,
// and then calls GetDataFromFile to pre-load all existing data into the in-memory map.
func CreateStoreFile(fileStoragePath string) (FileStorageJSON, error) {
	var fs FileStorageJSON
	fs.URLMap = make(map[string]string)

	if len(fileStoragePath) == 0 {
		fs.INMemory = true
		return fs, nil
	}

	producer, err := models.NewProducer(fileStoragePath)
	if err != nil {
		return fs, err
	}
	fs.Producer = producer

	consumer, err := models.NewConsumer(fileStoragePath)
	if err != nil {
		producer.Close()
		return fs, err
	}
	fs.Consumer = consumer
	fs.URLMap = GetDataFromFile(consumer)
	return fs, nil
}

// Close releases the resources held by the FileStorageJSON, specifically by
// closing the underlying file handles for the producer and consumer. This should
// be called when the application is shutting down. It safely handles cases where
// persistence is disabled (and thus producer/consumer are nil).
func (fs *FileStorageJSON) Close() {
	if fs.Producer != nil {
		fs.Producer.Close()
	}
	if fs.Consumer != nil {
		fs.Consumer.Close()
	}
}

// Ping return nil for FileStorage
func (fs FileStorageJSON) Ping(ctx context.Context) error {
	return nil
}
