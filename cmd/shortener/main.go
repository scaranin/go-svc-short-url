package main

import (
	"crypto/sha1"
	"encoding/base64"
	"io"
	"net/http"
	"strings"
)

const contentTypeTextPlain string = "text/plain"

var mapURL map[string]string

func getHandle(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("content-type", contentTypeTextPlain)

	shortURL, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
	}

	var url string
	if len(shortURL) != 0 {
		url = getURL(string(shortURL)[1:])
	} else {
		url = ""
	}

	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
	}

	res.Header().Set("Location", url)
	res.WriteHeader(http.StatusTemporaryRedirect)
	res.Write([]byte(url))
}

func postHandle(res http.ResponseWriter, req *http.Request) {
	contentType := req.Header.Get("content-type")
	if !strings.Contains(contentType, contentTypeTextPlain) {
		http.Error(res, contentType+" not supported", http.StatusBadRequest)
		return
	}

	var err error
	res.Header().Set("content-type", contentTypeTextPlain)

	url, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
	}

	shortURL := addShortURL(string(url))

	res.WriteHeader(http.StatusCreated)
	res.Write([]byte("http://localhost:8080/" + shortURL))
}

func addShortURL(url string) string {
	hasher := sha1.New()
	if url == "" {
		url = "EmptyCode"
	}
	hasher.Write([]byte(url))
	shortURL := base64.URLEncoding.EncodeToString(hasher.Sum(nil))
	mapURL[shortURL] = url
	return shortURL

}

func getURL(shortURL string) string {
	return mapURL[shortURL]
}

func routeHandle(res http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		getHandle(res, req)
	case http.MethodPost:
		postHandle(res, req)
	default:
		http.Error(res, "Only post and get requests are allowed!", http.StatusBadRequest)
	}
}

func main() {
	mapURL = make(map[string]string)
	mux := http.NewServeMux()
	mux.HandleFunc(`/`, routeHandle)

	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}
