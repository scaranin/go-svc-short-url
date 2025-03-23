package main

import (
	"log"
	"net/http"

	"github.com/scaranin/go-svc-short-url/internal/api"
	"github.com/scaranin/go-svc-short-url/internal/config"
	"github.com/scaranin/go-svc-short-url/internal/handlers"
)

func main() {

	cfg, err := config.CreateConfig()
	if err != nil {
		log.Fatal(err)
	}

	store, err := config.CreateStore(cfg)
	if err != nil {
		log.Fatal(err)
	}

	h := handlers.CreateHandle(cfg, store)

	mux := api.InitRoute(&h)

	err = http.ListenAndServe(cfg.ServerURL, mux)
	if err != nil {
		log.Fatal(err)
	}
}
