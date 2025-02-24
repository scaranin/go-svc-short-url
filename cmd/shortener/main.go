package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/scaranin/go-svc-short-url/internal/handlers"
)

func main() {
	h := handlers.CreateConfig()

	req := chi.NewRouter()

	req.Route(`/`, func(req chi.Router) {
		req.Get(`/`, h.GetHandle)
		req.Post(`/`, h.PostHandle)
	})

	mux := http.NewServeMux()
	mux.HandleFunc(`/`, h.RouteHandle)

	err := http.ListenAndServe(h.Cfg.ServerURL, mux)
	if err != nil {
		log.Fatal(err)
	}
}
