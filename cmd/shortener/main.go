package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/scaranin/go-svc-short-url/internal/handlers"
	"github.com/scaranin/go-svc-short-url/internal/middleware"
)

func main() {
	h := handlers.CreateConfig()

	req := chi.NewRouter()

	req.Route(`/`, func(req chi.Router) {
		req.Get(`/`, middleware.WithLogging(h, http.MethodGet))
		req.Get(`/{shortURL}`, middleware.WithLogging(h, http.MethodGet))
		req.Post(`/`, middleware.WithLogging(h, http.MethodPost))
	})

	err := http.ListenAndServe(h.Cfg.ServerURL, req)
	if err != nil {
		log.Fatal(err)
	}
}
