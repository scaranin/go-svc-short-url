package api

import (
	"net/http/pprof"

	"github.com/go-chi/chi"
	"github.com/scaranin/go-svc-short-url/internal/handlers"
	"github.com/scaranin/go-svc-short-url/internal/middleware"
)

// InitRoute initializes and configures the router with all application routes and middleware.
// It sets up:
// - Logging and compression middleware
// - Core URL shortening routes (JSON and plaintext)
// - User-specific routes
// - Health check endpoint
// - Debug/profiling endpoints
// Returns a configured chi.Mux router ready for use.
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
		mux.Get("/api/internal/stats", h.GetStats)
		mux.Delete("/api/user/urls", h.DeleteHandle)

		mux.Get("/debug/pprof", pprof.Index)
		mux.Get("/debug/profile", pprof.Profile)
		mux.Get("/debug/symbol", pprof.Symbol)
		mux.Get("/debug/trace", pprof.Trace)
	})

	return mux
}
