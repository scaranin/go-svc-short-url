package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/scaranin/go-svc-short-url/internal/handlers"
	"github.com/scaranin/go-svc-short-url/internal/loggers"
)

func main() {
	h := handlers.CreateConfig()

	req := chi.NewRouter()

	req.Route(`/`, func(req chi.Router) {
		req.Get(`/{shortURL}`, loggers.WithLogger(h.GetHandle))
		req.Post(`/`, h.PostHandle)
	})

	err := http.ListenAndServe(h.Cfg.ServerURL, req)
	if err != nil {
		log.Fatal(err)
	}
}
