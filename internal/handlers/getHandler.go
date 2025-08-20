package handlers

import (
	"net/http"

	"github.com/go-chi/chi"

	"encoding/json"
	"log"
)

// GetHandle handles GET requests for short URLs, redirecting clients to the original URL.
// It extracts the `shortURL` from the path parameter.
//   - On success, it performs an HTTP 307 Temporary Redirect to the original URL.
//   - If the storage indicates the URL was deleted (by returning a "ROW_IS_DELETED" error),
//     it responds with an HTTP 410 Gone status.
//   - For any other lookup errors, it returns an HTTP 500 Internal Server Error.
//   - If the `shortURL` parameter is missing, it returns an HTTP 400 Bad Request.
func (h *URLHandler) GetHandle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", contentTypeTextPlain)
	shortURL := chi.URLParam(r, "shortURL")
	var originalURL string
	var err error
	if len(shortURL) != 0 {
		originalURL, err = h.Load(shortURL)
		if err != nil {
			if err.Error() == "ROW_IS_DELETED" {
				w.WriteHeader(http.StatusGone)
				return
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
	} else {
		http.Error(w, "Empty value", http.StatusBadRequest)
		return
	}
	w.Header().Add("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

// GetUserURLs is an HTTP handler that retrieves all URLs created by the currently authenticated user.
// It authenticates the user via a cookie. If the user is not authenticated, it responds with
// an appropriate status (e.g., 401 Unauthorized). If the user is authenticated but has no URLs,
// it responds with HTTP 204 No Content.
// On success, it prepends the service's BaseURL to each short URL identifier, marshals the
// list into a JSON array, and sends it back to the client with an HTTP 200 OK status.
func (h *URLHandler) GetUserURLs(w http.ResponseWriter, r *http.Request) {
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
	if err == http.ErrNoCookie {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	URLList, err := h.Storage.GetUserURLList(h.Auth.UserID)

	if err != nil || len(URLList) == 0 {
		http.SetCookie(w, cookieW)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	for i := range URLList {
		URLList[i].ShortURL = h.BaseURL + URLList[i].ShortURL
	}

	URLUserListJSON, err := json.Marshal(URLList)
	if err != nil {
		http.SetCookie(w, cookieW)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	http.SetCookie(w, cookieW)
	w.WriteHeader(http.StatusOK)
	w.Write(URLUserListJSON)

}
