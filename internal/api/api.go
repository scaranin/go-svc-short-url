package api

import (
	"github.com/go-chi/chi"
	"github.com/scaranin/go-svc-short-url/internal/handlers"
	"github.com/scaranin/go-svc-short-url/internal/middleware"
)

func InitRoute(h *handlers.URLHandler) *chi.Mux {
	mux := chi.NewRouter()
	mux.Use(middleware.WithLogging, middleware.GzipMiddleware)

	mux.Route("/", func(mux chi.Router) {
		mux.Post("/", h.PostHandle)
		mux.Post("/api/shorten", h.PostHandleJSON)
		mux.Post("/api/shorten/batch", h.PostHandleJSONBatch)
		mux.Get("/api/user/urls", h.GetUserURLs)
		mux.Get("/ping", h.PingHandle)
		mux.Get("/{shortURL}", h.GetHandle)
	})

	return mux
}
