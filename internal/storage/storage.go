package storage

import (
	"io"
	"log"

	"github.com/scaranin/go-svc-short-url/internal/models"
)

type BaseFileJSON struct {
	Producer *models.Producer
	Consumer *models.Consumer
}

type Storage interface {
	Save(URL string)
	Load(shortURL string) (string, bool)
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
		urlMap[mURL.ShortURL] = mURL.URL
	}
	return urlMap
}

func CreateStore(fileStoragePath string) BaseFileJSON {
	var bfj BaseFileJSON
	Producer, err := models.NewProducer(fileStoragePath)
	if err != nil {
		log.Fatal(err)
	}
	bfj.Producer = Producer

	Consumer, err := models.NewConsumer(fileStoragePath)
	if err != nil {
		log.Fatal(err)
	}
	bfj.Consumer = Consumer
	return bfj
}

func (h *BaseFileJSON) Close() {
	h.Producer.Close()
	h.Consumer.Close()
}
