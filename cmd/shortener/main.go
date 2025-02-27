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
		req.Get(`/{shortURL}`, middleware.WithLogging(h, "GetShortUrlText"))
		req.Post(`/`, middleware.WithLogging(h, "PostRootText"))
		req.Post(`/api/shorten`, middleware.WithLogging(h, "PostApiShortenJson"))
	})

	err := http.ListenAndServe(h.Cfg.ServerURL, req)
	if err != nil {
		log.Fatal(err)
	}
}
