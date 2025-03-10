package main

import (
	"crypto/sha1"
	"encoding/base64"
	"flag"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/caarlos0/env/v6"
	"github.com/go-chi/chi"
	"github.com/scaranin/go-svc-short-url/internal/config"
)

const contentTypeTextPlain string = "text/plain"

var (
	mapURL    = make(map[string]string)
	serverURL string
	baseURL   string
)

type envConfig struct {
	serverURL string `env:"SERVER_ADDRESS"`
	baseURL   string `env:"BASE_URL"`
}

func getHandle(res http.ResponseWriter, req *http.Request) {

	res.Header().Set("content-type", contentTypeTextPlain)

	shortURL := strings.TrimPrefix(req.URL.Path, "/")

	var url string
	if len([]rune(shortURL)) != 0 {
		url = getURL(shortURL)
	} else {
		http.Error(res, "Empty value", http.StatusBadRequest)
		return
	}
	res.Header().Set("Location", url)
	res.WriteHeader(http.StatusTemporaryRedirect)
}

func postHandle(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("content-type", contentTypeTextPlain)

	contentType := req.Header.Get("content-type")

	if !strings.Contains(contentType, contentTypeTextPlain) {
		http.Error(res, contentType+" not supported", http.StatusBadRequest)
		return
	}

	var err error
	url, err := io.ReadAll(req.Body)

	defer req.Body.Close()

	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	if len(url) == 0 {
		http.Error(res, "Empty value", http.StatusBadRequest)
		return
	}

	shortURL := addShortURL(string(url))

	res.WriteHeader(http.StatusCreated)
	res.Write([]byte(baseURL + shortURL))
}

func addShortURL(url string) string {
	hasher := sha1.New()

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

func setParams() {
	var cfg envConfig
	err := env.Parse(&cfg)

	if err != nil {
		log.Fatal(err)
	}

	netCfg := config.New()

	flag.StringVar(&netCfg.ServerURL, "a", "localhost:8080", "Server URL")
	flag.StringVar(&netCfg.BaseURL, "b", "http://localhost:8080", "Base URL")
	flag.Parse()

	if len([]rune(cfg.serverURL)) != 0 {
		serverURL = cfg.serverURL
	} else {
		serverURL = netCfg.ServerURL
	}

	if len([]rune(cfg.baseURL)) != 0 {
		baseURL = cfg.baseURL + "/"
	} else {
		baseURL = netCfg.BaseURL + "/"
	}
}

func main() {
	setParams()

	req := chi.NewRouter()

	req.Route(`/`, func(req chi.Router) {
		req.Get(`/`, getHandle)
		req.Post(`/`, postHandle)
	})

	mux := http.NewServeMux()
	mux.HandleFunc(`/`, routeHandle)

	err := http.ListenAndServe(serverURL, mux)
	if err != nil {
		panic(err)
	}
}
