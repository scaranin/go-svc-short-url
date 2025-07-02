package handlers

import (
	"log"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PingHandle serves as an HTTP handler to check the health of the database connection.
// It attempts to create a new connection pool using the DSN from the handler
// and then pings the database to verify connectivity.
//
// On success, it responds with an HTTP 200 OK status. If creating the connection pool
// or the ping fails, it responds with an HTTP 500 Internal Server Error.
func (h *URLHandler) PingHandle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", contentTypeTextPlain)
	pool, err := pgxpool.New(r.Context(), h.DSN)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}
	defer pool.Close()

	err = pool.Ping(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}

	w.WriteHeader(http.StatusOK)
}
