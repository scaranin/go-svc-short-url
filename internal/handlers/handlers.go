package handlers

import (
	"crypto/sha1"
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/caarlos0/env"
	"github.com/go-chi/chi"
	"github.com/scaranin/go-svc-short-url/internal/config"
)

const contentTypeTextPlain string = "text/plain"

type EnvConfig struct {
	ServerURL string `env:"SERVER_ADDRESS"`
	BaseURL   string `env:"BASE_URL"`
}

type URLHandler struct {
	urlMap map[string]string
	Cfg    EnvConfig
}

func CreateConfig() URLHandler {
	var cfg EnvConfig
	var h URLHandler
	h.urlMap = make(map[string]string)

	err := env.Parse(&cfg)

	if err != nil {
		log.Fatal(err)
	}

	netCfg := config.New()

	if flag.Lookup("a") == nil {
		flag.StringVar(&netCfg.ServerURL, "a", "localhost:8080", "Server URL")
	}
	if flag.Lookup("b") == nil {
		flag.StringVar(&netCfg.BaseURL, "b", "http://localhost:8080/", "Base URL")
	}
	flag.Parse()

	if len(cfg.ServerURL) == 0 {
		cfg.ServerURL = netCfg.ServerURL
	}

	if len(cfg.BaseURL) == 0 {
		cfg.BaseURL = netCfg.BaseURL + "/"
	}

	h.Cfg = cfg
	return h
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

	shortURL := h.addShortURL(string(url))

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(h.Cfg.BaseURL + shortURL))
}

func (h *URLHandler) addShortURL(url string) string {
	hasher := sha1.New()

	hasher.Write([]byte(url))

	shortURL := base64.URLEncoding.EncodeToString(hasher.Sum(nil))
	h.urlMap[shortURL] = url

	return shortURL
}

func (h *URLHandler) GetHandle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", contentTypeTextPlain)

	shortURL := chi.URLParam(r, "shortURL")
	fmt.Println("shortURL", shortURL)

	var url string
	if len(shortURL) != 0 {
		url = h.getURL(shortURL)
	} else {
		http.Error(w, "Empty value", http.StatusBadRequest)
		return
	}
	w.Header().Set("Location", url)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func (h *URLHandler) getURL(shortURL string) string {
	return h.urlMap[shortURL]
}
