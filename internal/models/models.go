package models

import (
	"encoding/json"
	"os"
)

type Storage interface {
	Save(URL *URL) (string, error)
	Load(shortURL string) (string, error)
	GetUserURLList(UserID string) ([]URLUserList, error)
}

type Request struct {
	URL string `json:"url"`
}

type Response struct {
	Result string `json:"result"`
}

type URL struct {
	CorrelationID string `json:"-"`
	OriginalURL   string `json:"url"`
	ShortURL      string `json:"shorturl"`
	UserID        string `json:"-"`
}

type PairRequest struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type PairResponse struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

type URLUserList struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type Producer struct {
	file    *os.File
	encoder *json.Encoder
}

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

func (p *Producer) AddURL(url *URL) error {
	return p.encoder.Encode(url)
}

func (p *Producer) Close() error {
	return p.file.Close()
}

type Consumer struct {
	file    *os.File
	decoder *json.Decoder
}

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

func (c *Consumer) GetURL() (*URL, error) {
	sURL := &URL{}
	if err := c.decoder.Decode(&sURL); err != nil {
		return nil, err
	}
	return sURL, nil
}

func (c *Consumer) Close() error {
	return c.file.Close()
}
