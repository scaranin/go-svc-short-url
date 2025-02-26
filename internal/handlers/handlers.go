package handlers

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"flag"
	"io"
	"log"
	"net/http"

	"github.com/caarlos0/env"
	"github.com/go-chi/chi"
	"github.com/scaranin/go-svc-short-url/internal/config"
	"github.com/scaranin/go-svc-short-url/internal/models"
)

const contentTypeTextPlain string = "text/plain"
const contentTypeApJson string = "application/json"

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
		flag.StringVar(&netCfg.BaseURL, "b", "http://localhost:8080", "Base URL")
	}
	flag.Parse()

	if len(cfg.ServerURL) == 0 {
		cfg.ServerURL = netCfg.ServerURL
	}

	if len(cfg.BaseURL) == 0 {
		cfg.BaseURL = netCfg.BaseURL
	}

	cfg.BaseURL += "/"

	h.Cfg = cfg
	return h
}

func (h *URLHandler) PostHandle(w http.ResponseWriter, r *http.Request) {
	var (
		url         []byte
		err         error
		contentType string = r.Header.Get("Content-Type")
	)
	switch contentType {
	case contentTypeTextPlain:
		{
			url, err = io.ReadAll(r.Body)
			if len(url) == 0 {
				w.WriteHeader(http.StatusCreated)
				return
			}
		}
	default:
		{
			http.Error(w, contentType+" not supported", http.StatusBadRequest)
			return
		}
	}

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

	w.Header().Set("Content-Type", contentTypeTextPlain)
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(h.Cfg.BaseURL + shortURL))

}

func (h *URLHandler) PostHandleJson(w http.ResponseWriter, r *http.Request) {
	var (
		url         []byte
		err         error
		contentType string = r.Header.Get("content-type")
	)

	switch contentType {
	case contentTypeApJson:
		{
			var req models.Request
			var buf bytes.Buffer
			_, err := buf.ReadFrom(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if err = json.Unmarshal(buf.Bytes(), &req); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			url = []byte(req.Url)
		}
	default:
		{
			http.Error(w, contentType+" not supported", http.StatusBadRequest)
			return
		}
	}

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

	var res models.Response
	res.Result = h.Cfg.BaseURL + shortURL
	resp, err := json.Marshal(res)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", contentTypeApJson)
	w.WriteHeader(http.StatusCreated)
	w.Write(resp)
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
