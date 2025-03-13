package storage

import (
	"io"
	"log"

	"github.com/scaranin/go-svc-short-url/internal/models"
)

type FileStorageJSON struct {
	Producer *models.Producer
	Consumer *models.Consumer
	URLMap   map[string]string
	INMemory bool
}

func (fs FileStorageJSON) Save(URL *models.URL) (string, error) {
	var err error
	if !fs.INMemory {
		err = fs.Producer.AddURL(URL)
	}
	fs.URLMap[URL.ShortURL] = URL.OriginalURL
	return URL.ShortURL, err
}

func (fs FileStorageJSON) Load(shortURL string) (string, error) {
	originalURL, _ := fs.URLMap[shortURL]

	return originalURL, nil
}

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

func CreateStoreFile(fileStoragePath string) (FileStorageJSON, error) {
	var fs FileStorageJSON
	fs.URLMap = make(map[string]string)
	if len(fileStoragePath) == 0 {
		fs.INMemory = true
		return fs, nil
	}
	Producer, err := models.NewProducer(fileStoragePath)
	if err != nil {
		return fs, err
	}
	fs.Producer = Producer

	Consumer, err := models.NewConsumer(fileStoragePath)
	if err != nil {
		return fs, err
	}
	fs.Consumer = Consumer
	fs.URLMap = GetDataFromFile(Consumer)
	return fs, err
}

func (fs *FileStorageJSON) Close() {
	fs.Producer.Close()
	fs.Consumer.Close()
}
