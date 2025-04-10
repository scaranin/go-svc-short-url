package main

import (
	"log"
	"net/http"

	"github.com/scaranin/go-svc-short-url/internal/api"
	"github.com/scaranin/go-svc-short-url/internal/config"
	"github.com/scaranin/go-svc-short-url/internal/handlers"
)

func main() {

	cfg := config.CreateConfig()

	store, _ := config.CreateStore(cfg)

	h := handlers.CreateHandle(cfg, store)

	req := api.InitRoute(&h)

	err := http.ListenAndServe(cfg.ServerURL, req)
	if err != nil {
		log.Fatal(err)
	}
}
