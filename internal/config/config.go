package config

import (
	"encoding/json"
	"flag"
	"log"
	"os"

	"github.com/caarlos0/env"
	"github.com/scaranin/go-svc-short-url/internal/models"
	"github.com/scaranin/go-svc-short-url/internal/storage"
)

// ShortenerConfig contains all configuration parameters for the URL shortener service.
// Fields are tagged for environment variable parsing using github.com/caarlos0/env.
type ShortenerConfig struct {
	ServerURL       string `json:"server_address" env:"SERVER_ADDRESS"`
	BaseURL         string `json:"base_url" env:"BASE_URL"`
	FileStoragePath string `json:"file_storage_path" env:"FILE_STORAGE_PATH"`
	DSN             string `json:"database_dsn" env:"DATABASE_DSN"`
	HTTPSMode       string `json:"enable_https" env:"ENABLE_HTTPS"`
}

// New creates a new ShortenerConfig with default values.
// Default values:
//   - ServerURL: "localhost:8080"
//   - BaseURL: "http://localhost:8080"
//   - FileStoragePath: "BaseFile.json"
//   - DSN: "postgres://postgres:admin@localhost:5432/postgres"
//   - HTTPSMode: "false"
func New() ShortenerConfig {
	return ShortenerConfig{
		ServerURL:       "localhost:8080",
		BaseURL:         "http://localhost:8080/",
		FileStoragePath: "BaseFile.json",
		DSN:             "postgres://postgres:admin@localhost:5432/postgres",
		HTTPSMode:       "false",
	}

}

func fillConfig(srcCfg *ShortenerConfig, dstCfg *ShortenerConfig) {
	if len(srcCfg.ServerURL) == 0 {
		srcCfg.ServerURL = dstCfg.ServerURL
	}

	if len(srcCfg.BaseURL) == 0 {
		srcCfg.BaseURL = dstCfg.BaseURL
		srcCfg.BaseURL += "/"
	}

	if len(srcCfg.FileStoragePath) == 0 {
		srcCfg.FileStoragePath = dstCfg.FileStoragePath
	}

	if len(srcCfg.DSN) == 0 {
		srcCfg.DSN = dstCfg.DSN
	}

	if len(srcCfg.HTTPSMode) == 0 {
		srcCfg.HTTPSMode = dstCfg.HTTPSMode
	}
}

// CreateConfig loads and initializes application configuration.
// It follows this precedence order:
//  1. Environment variables (highest priority)
//  2. Command-line flags
//  3. JSON config file
//  3. Default values (lowest priority)
//
// Returns:
//   - ShortenerConfig: The populated configuration
//   - error: Any error that occurred during parsing
func CreateConfig() (ShortenerConfig, error) {
	var Cfg ShortenerConfig

	err := env.Parse(&Cfg)

	if err != nil {
		return Cfg, err
	}

	NetCfg := New()

	if flag.Lookup("s") == nil {
		flag.StringVar(&NetCfg.HTTPSMode, "s", "false", "HTTPS mode")
	}

	if flag.Lookup("a") == nil {
		flag.StringVar(&NetCfg.ServerURL, "a", "localhost:8080", "Server URL")
	}
	if flag.Lookup("b") == nil {
		if NetCfg.HTTPSMode == "true" {
			flag.StringVar(&NetCfg.BaseURL, "b", "https://localhost:8080", "Base URL")
		} else {
			flag.StringVar(&NetCfg.BaseURL, "b", "http://localhost:8080", "Base URL")
		}

	}
	if flag.Lookup("f") == nil {
		flag.StringVar(&NetCfg.FileStoragePath, "f", "BaseFile.json", "File storage path")
	}
	if flag.Lookup("d") == nil {
		flag.StringVar(&NetCfg.DSN, "d", "postgres://postgres:admin@localhost:5432/postgres", "DataBase DSN")
	}

	flag.Parse()

	fillConfig(&Cfg, &NetCfg)

	byteFile, err := os.ReadFile("./internal/config/config.json")
	if err != nil {
		return Cfg, err
	}
	json.Unmarshal(byteFile, &NetCfg)

	fillConfig(&Cfg, &NetCfg)

	return Cfg, err
}

// CreateStore initializes the appropriate storage implementation based on configuration.
// Storage selection logic:
//  1. Attempt to use PostgreSQL if DSN is configured
//  2. Fall back to file storage if FileStoragePath is configured
//  3. Fall back to in-memory storage if neither is configured
//
// Returns:
//   - models.Storage: The initialized storage implementation
//   - error: Any error that occurred during initialization
func CreateStore(cfg ShortenerConfig) (models.Storage, error) {
	var store models.Storage
	var err error
	if len(cfg.DSN) > 0 {
		store, err = storage.CreateStoreDB(cfg.DSN)
		if err == nil {
			log.Println("DBStoreMode")
			return store, err
		}
	}
	if len(cfg.FileStoragePath) > 0 {
		log.Println("FileStoreMode")
	} else {
		log.Println("InMemoryMode")
	}

	return storage.CreateStoreFile(cfg.FileStoragePath)

}
