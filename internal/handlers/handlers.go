package handlers

import (
	"crypto/sha1"
	"encoding/base64"
	"io"
	"net/http"
	"strings"
)

const contentTypeTextPlain string = "text/plain"

var (
	mapURL    = make(map[string]string)
	serverURL string
	baseURL   string
)

type URLHandler struct {
	urlMap  map[string]string
	baseURL string
}

func (h *URLHandler) PostHandle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", contentTypeTextPlain)

	contentType := r.Header.Get("content-type")

	if !strings.Contains(contentType, contentTypeTextPlain) {
		http.Error(w, contentType+" not supported", http.StatusBadRequest)
		return
	}

	var err error
	url, err := io.ReadAll(r.Body)

	defer r.Body.Close()

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(url) == 0 {
		http.Error(w, "Empty value", http.StatusBadRequest)
		return
	}

	shortURL := addShortURL(string(url))

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(baseURL + shortURL))
}

func addShortURL(url string) string {
	hasher := sha1.New()

	hasher.Write([]byte(url))

	shortURL := base64.URLEncoding.EncodeToString(hasher.Sum(nil))
	mapURL[shortURL] = url

	return shortURL
}

func (h *URLHandler) GetHandle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", contentTypeTextPlain)

	shortURL := strings.TrimPrefix(r.URL.Path, "/")

	var url string
	if len(shortURL) != 0 {
		url = getURL(shortURL)
	} else {
		http.Error(w, "Empty value", http.StatusBadRequest)
		return
	}
	w.Header().Set("Location", url)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func getURL(shortURL string) string {
	return mapURL[shortURL]
}
