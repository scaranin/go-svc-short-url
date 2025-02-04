package main

import (
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"io"
	"net/http"
)

const contentTypeTextPlain string = "text/plain"

var mapURL map[string]string

func getHandle(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("content-type", contentTypeTextPlain)

	shortURL, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
	}

	url := getURL(string(shortURL))
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
	}
	res.Header().Set("Location", url)
	res.WriteHeader(http.StatusTemporaryRedirect)
	res.Write([]byte(url))
}

func postHandle(res http.ResponseWriter, req *http.Request) {
	contentType := req.Header.Get("content-type")
	if !trings.Contains(contentType, contentTypeTextPlain) {
		http.Error(res, contentType+" not supported", http.StatusBadRequest)
		return
	}

	var err error
	res.Header().Set("content-type", contentTypeTextPlain)

	url, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
	}

	shortURL, err := addShortURL(string(url))
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
	}
	res.WriteHeader(http.StatusCreated)
	res.Write([]byte("http://localhost:8080/" + shortURL))
}

func addShortURL(url string) (string, error) {
	if len([]rune(url)) == 0 {
		return nil, errors.New("Wrong URL")
	} else {
		hasher := sha1.New()
		hasher.Write([]byte(url))
		shortURL := base64.URLEncoding.EncodeToString(hasher.Sum(nil))
		mapURL[shortURL] = url
		return shortURL, nil
	}

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
