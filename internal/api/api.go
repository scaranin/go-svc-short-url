package api

import (
	"github.com/go-chi/chi"
	"github.com/scaranin/go-svc-short-url/internal/handlers"
	"github.com/scaranin/go-svc-short-url/internal/middleware"
)

func InitRoute(h *handlers.URLHandler) *chi.Mux {
	req := chi.NewRouter()
	req.Use(middleware.WithLogging, middleware.GzipMiddleware)

	req.Route("/", func(req chi.Router) {
		req.Get("/ping", h.PingHandle)
		req.Get("/{shortURL}", h.GetHandle)
		req.Post("/", h.PostHandle)
		req.Post("/api/shorten", h.PostHandleJSON)
	})

	return req
}
