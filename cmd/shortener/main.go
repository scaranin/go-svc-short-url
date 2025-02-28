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
	defer h.Close()

	req := chi.NewRouter()

	req.Use(middleware.WithLogging, middleware.GzipMiddleware)

	h.BaseFile.Consumer.GetURL()

	req.Route(`/`, func(req chi.Router) {
		req.Get(`/{shortURL}`, h.GetHandle)
		req.Post(`/`, h.PostHandle)
		req.Post(`/api/shorten`, h.PostHandleJSON)
	})

	err := http.ListenAndServe(h.Cfg.ServerURL, req)
	if err != nil {
		log.Fatal(err)
	}
}
