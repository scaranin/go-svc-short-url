package handlers

import (
	"encoding/json"
	"log"
	"net/http"
)

// DelBatch processes short URLs for deletion from a channel. It acts as a worker
// that consumes short URL strings from the provided channel. For each URL received,
// it calls the `DeleteBulk` method of the storage layer, associating the deletion
// with the `UserID` from the handler's context. This function is designed to run
// in a separate goroutine to handle deletion tasks asynchronously.
func (h *URLHandler) DelBatch(shortURLchan <-chan string) {
	var shortURLs []string
	for req := range shortURLchan {
		shortURLs = append(shortURLs, req)

		h.Storage.DeleteBulk(h.Auth.UserID, shortURLs)
		shortURLs = nil
	}
}

// DeleteHandle is an HTTP handler for asynchronously deleting a batch of user-owned URLs.
// It expects a JSON request body containing an array of short URL strings to be deleted.
// The handler sets up a worker to process deletions in the background. It reads the
// short URLs from the request, places them onto a channel, and launches a goroutine
// (`DelBatch`) to perform the deletion.
//
// Crucially, it responds immediately with an HTTP 202 Accepted status, indicating that
// the deletion requests have been received and will be processed without blocking the client.
// It also handles user authentication cookies as part of the request-response cycle.
func (h *URLHandler) DeleteHandle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", contentTypeApJSON)
	var (
		cookieW *http.Cookie
		err     error
	)
	cookieR, err := r.Cookie(h.Auth.CookieName)
	if err != nil {
		log.Print(err.Error())
	}

	cookieW, err = h.Auth.FillUserReturnCookie(cookieR)
	if err != nil {
		log.Fatal(err)
	}
	finalCh := make(chan string, 1024)

	var ShortURLs []string
	if err := json.NewDecoder(r.Body).Decode(&ShortURLs); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	go func(ShortURLs []string) {
		for _, shortURL := range ShortURLs {
			finalCh <- shortURL

		}
	}(ShortURLs)

	go h.DelBatch(finalCh)

	http.SetCookie(w, cookieW)
	w.WriteHeader(http.StatusAccepted)
}
