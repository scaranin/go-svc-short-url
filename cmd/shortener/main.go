package main

import (
	"log"
	"net/http"

	"github.com/scaranin/go-svc-short-url/internal/api"
	"github.com/scaranin/go-svc-short-url/internal/config"
	"github.com/scaranin/go-svc-short-url/internal/handlers"
	"github.com/scaranin/go-svc-short-url/internal/storage"
)

func main() {

	cfg := config.CreateConfig()

	store := storage.CreateStore(cfg.FileStoragePath)
	defer store.Close()
	h := handlers.CreateHandle(cfg, store)

	req := api.InitRoute(&h)

	err := http.ListenAndServe(cfg.ServerURL, req)
	if err != nil {
		log.Fatal(err)
	}
}
