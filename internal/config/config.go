package config

type ShortenerConfig struct {
	ServerURL       string
	BaseURL         string
	FileStorageParh string
}

func New() *ShortenerConfig {
	return &ShortenerConfig{ServerURL: "localhost:8080", BaseURL: "http://localhost:8080", FileStorageParh: "BaseFile.json"}

}
