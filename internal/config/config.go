package config

import (
	"flag"
	"log"

	"github.com/caarlos0/env"
)

type ShortenerConfig struct {
	ServerURL       string `env:"SERVER_ADDRESS"`
	BaseURL         string `env:"BASE_URL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	DSN             string `env:"DATABASE_DSN"`
}

func New() *ShortenerConfig {
	return &ShortenerConfig{ServerURL: "localhost:8080", BaseURL: "http://localhost:8080", FileStoragePath: "BaseFile.json", DSN: "postgres://postgres:admin@localhost:5432/postgres"}

}

func CreateConfig() ShortenerConfig {
	var Cfg ShortenerConfig

	err := env.Parse(&Cfg)

	if err != nil {
		log.Fatal(err)
	}

	NetCfg := New()

	if flag.Lookup("a") == nil {
		flag.StringVar(&NetCfg.ServerURL, "a", "localhost:8080", "Server URL")
	}
	if flag.Lookup("b") == nil {
		flag.StringVar(&NetCfg.BaseURL, "b", "http://localhost:8080", "Base URL")
	}
	if flag.Lookup("f") == nil {
		flag.StringVar(&NetCfg.FileStoragePath, "f", "BaseFile.json", "Base URL")
	}
	if flag.Lookup("d") == nil {
		flag.StringVar(&NetCfg.DSN, "d", "postgres://postgres:admin@localhost:5432/postgres", "DataBase DSN")

	}
	flag.Parse()

	if len(Cfg.ServerURL) == 0 {
		Cfg.ServerURL = NetCfg.ServerURL
	}

	if len(Cfg.BaseURL) == 0 {
		Cfg.BaseURL = NetCfg.BaseURL
	}

	Cfg.BaseURL += "/"

	if len(Cfg.FileStoragePath) == 0 {
		Cfg.FileStoragePath = NetCfg.FileStoragePath
	}

	if len(Cfg.DSN) == 0 {
		Cfg.DSN = NetCfg.DSN
	}

	return Cfg
}
