package main

import (
	"flag"
	"log"
	"net/http"
	"strings"

	"github.com/caarlos0/env/v6"
	"github.com/go-chi/chi"
	"github.com/scaranin/go-svc-short-url/internal/config"
	"github.com/scaranin/go-svc-short-url/internal/handlers"
)

var (
	mapURL    = make(map[string]string)
	serverURL string
	baseURL   string
)

type envConfig struct {
	ServerURL string `env:"SERVER_ADDRESS"`
	BaseURL   string `env:"BASE_URL"`
}

func getHandle(res http.ResponseWriter, req *http.Request) {

	res.Header().Set("content-type", contentTypeTextPlain)

	shortURL := strings.TrimPrefix(req.URL.Path, "/")

	var url string
	if len(shortURL) != 0 {
		url = getURL(shortURL)
	} else {
		http.Error(res, "Empty value", http.StatusBadRequest)
		return
	}
	res.Header().Set("Location", url)
	res.WriteHeader(http.StatusTemporaryRedirect)
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

	if len(cfg.ServerURL) != 0 {
		serverURL = cfg.ServerURL
	} else {
		serverURL = netCfg.ServerURL
	}

	if len(cfg.BaseURL) != 0 {
		baseURL = cfg.BaseURL + "/"
	} else {
		baseURL = netCfg.BaseURL + "/"
	}
}

func main() {
	setParams()

	req := chi.NewRouter()

	h := handlers.URLHandler;
	
	h.GetHandle

	req.Route(`/`, func(req chi.Router) {
		req.Get(`/`, handlers.GetHandle)
		req.Post(`/`, handlers.URLHandler. )
	})

	mux := http.NewServeMux()
	mux.HandleFunc(`/`, routeHandle)

	err := http.ListenAndServe(serverURL, mux)
	if err != nil {
		log.Fatal(err)
	}
}
