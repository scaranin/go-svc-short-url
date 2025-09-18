package models

import (
	"context"
	"encoding/json"
	"os"
)

// Storage defines the interface for URL persistence layers.
// It abstracts the underlying data store (e.g., in-memory map, file, database),
// allowing different implementations to be used interchangeably.
type Storage interface {
	// Save takes a URL object and persists it. It returns the short URL identifier
	// and an error if the operation fails (e.g., a conflict on a unique URL).
	Save(URL *URL) (string, error)
	// Load retrieves the original URL corresponding to a given short URL identifier.
	// It returns an error if the short URL is not found or has been marked as deleted.
	Load(shortURL string) (string, error)
	// GetUserURLList retrieves a list of all URLs created by a specific user.
	// It returns a slice of URLUserList objects and an error if the query fails.
	GetUserURLList(UserID string) ([]URLUserList, error)
	// DeleteBulk marks a batch of URLs for deletion for a specific user.
	// This is typically a "soft delete" operation.
	DeleteBulk(UserID string, ShortURLs []string) error
	// Ping storage.
	Ping(ctx context.Context) error
	// GetStats return full storage statistic: users and URLs count.
	GetStats() (Statistic, error)
}

// Request represents the JSON structure for a single URL shortening request.
type Request struct {
	// URL is the original URL to be shortened.
	URL string `json:"url"`
}

// Response represents the JSON structure for a single URL shortening response.
type Response struct {
	// Result contains the generated short URL.
	Result string `json:"result"`
}

// URL represents the core data model for a shortened URL, linking the
// original and short versions, and associating it with a user.
type URL struct {
	// CorrelationID is an optional identifier used in batch operations.
	// The `json:"-"` tag prevents it from being serialized into JSON.
	CorrelationID string `json:"-"`
	// OriginalURL is the original, full-length URL.
	OriginalURL string `json:"url"`
	// ShortURL is the generated short URL identifier.
	ShortURL string `json:"shorturl"`
	// UserID is the identifier of the user who owns this URL.
	// The `json:"-"` tag prevents it from being serialized into JSON.
	UserID string `json:"-"`
}

// PairRequest represents a single item in a batch shortening request.
type PairRequest struct {
	// CorrelationID is a client-provided ID to track this specific request.
	CorrelationID string `json:"correlation_id"`
	// OriginalURL is the URL to be shortened for this item.
	OriginalURL string `json:"original_url"`
}

// PairResponse represents a single item in a batch shortening response.
type PairResponse struct {
	// CorrelationID is the same ID from the request to correlate the response.
	CorrelationID string `json:"correlation_id"`
	// ShortURL is the fully-formed short URL for the corresponding original URL.
	ShortURL string `json:"short_url"`
}

// URLUserList is a data transfer object representing a single URL entry
// returned in the list of a user's URLs.
type URLUserList struct {
	// ShortURL is the full, publicly-accessible short URL.
	ShortURL string `json:"short_url"`
	// OriginalURL is the original URL that was shortened.
	OriginalURL string `json:"original_url"`
}

// Producer is responsible for writing URL data to a file in a streaming JSON format.
type Producer struct {
	file    *os.File
	encoder *json.Encoder
}

// Statistic represents storage statistics including the total number of URLs and users.
// It's used for marshaling/unmarshaling JSON data, with fields tagged for JSON output.
type Statistic struct {
	URLs  int `json:"urls"`
	Users int `json:"users"`
}

// NewProducer creates a new Producer for writing to the specified file.
// The caller is responsible for calling Close() on the producer to release resources.
func NewProducer(filename string) (*Producer, error) {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}
	return &Producer{
		file:    file,
		encoder: json.NewEncoder(file),
	}, nil
}

// AddURL encodes the given URL object as JSON and writes it to the file.
func (p *Producer) AddURL(url *URL) error {
	return p.encoder.Encode(url)
}

// Close closes the underlying file handle.
func (p *Producer) Close() error {
	return p.file.Close()
}

// Consumer is responsible for reading URL data from a file that was written
// in a streaming JSON format.
type Consumer struct {
	file    *os.File
	decoder *json.Decoder
}

// NewConsumer creates a new Consumer for reading from the specified file.
// The caller is responsible for calling Close() on the consumer to release resources.
func NewConsumer(filename string) (*Consumer, error) {
	file, err := os.OpenFile(filename, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}
	return &Consumer{
		file:    file,
		decoder: json.NewDecoder(file),
	}, nil
}

// GetURL reads and decodes the next JSON object from the file into a URL struct.
// It returns an `io.EOF` error when there are no more objects to read.
func (c *Consumer) GetURL() (*URL, error) {
	sURL := &URL{}
	if err := c.decoder.Decode(&sURL); err != nil {
		return nil, err
	}
	return sURL, nil
}

// Close closes the underlying file handle.
func (c *Consumer) Close() error {
	return c.file.Close()
}
