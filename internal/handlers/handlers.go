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
const contentTypeApJSON string = "application/json"

type EnvConfig struct {
	ServerURL       string `env:"SERVER_ADDRESS"`
	BaseURL         string `env:"BASE_URL"`
	FileStorageParh string `env:"FILE_STORAGE_PATH"`
}

type BaseFileJSON struct {
	Producer *models.Producer
	Consumer *models.Consumer
}

type URLHandler struct {
	urlMap   map[string]string
	Cfg      EnvConfig
	BaseFile BaseFileJSON
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
	if flag.Lookup("f") == nil {
		flag.StringVar(&netCfg.FileStorageParh, "f", "BaseFile.json", "Base URL")
	}
	flag.Parse()

	if len(cfg.ServerURL) == 0 {
		cfg.ServerURL = netCfg.ServerURL
	}

	if len(cfg.BaseURL) == 0 {
		cfg.BaseURL = netCfg.BaseURL
	}

	cfg.BaseURL += "/"

	if len(cfg.FileStorageParh) == 0 {
		cfg.FileStorageParh = netCfg.FileStorageParh
	}

	Producer, err := models.NewProducer(cfg.FileStorageParh)
	if err != nil {
		log.Fatal(err)
	}
	h.BaseFile.Producer = Producer

	Consumer, err := models.NewConsumer(cfg.FileStorageParh)
	if err != nil {
		log.Fatal(err)
	}
	h.BaseFile.Consumer = Consumer

	for {
		mURL, err := h.BaseFile.Consumer.GetURL()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		h.urlMap[mURL.ShortURL] = mURL.URL
	}

	h.Cfg = cfg
	return h
}

func (h *URLHandler) PostHandle(w http.ResponseWriter, r *http.Request) {
	var (
		url []byte
		err error
	)
	w.Header().Set("Content-Type", contentTypeTextPlain)

	url, err = io.ReadAll(r.Body)
	if len(url) == 0 {
		w.WriteHeader(http.StatusCreated)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	shortURL := h.addShortURL(string(url))

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(h.Cfg.BaseURL + shortURL))

	defer r.Body.Close()

}

func (h *URLHandler) PostHandleJSON(w http.ResponseWriter, r *http.Request) {
	var (
		url []byte
		err error
	)

	var req models.Request
	var buf bytes.Buffer
	_, err = buf.ReadFrom(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err = json.Unmarshal(buf.Bytes(), &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	url = []byte(req.URL)

	defer r.Body.Close()

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
	w.Header().Set("Content-Type", contentTypeApJSON)
	w.WriteHeader(http.StatusCreated)
	w.Write(resp)
}

func (h *URLHandler) Close() {
	h.BaseFile.Producer.Close()
	h.BaseFile.Consumer.Close()
}

func (h *URLHandler) addShortURL(url string) string {
	hasher := sha1.New()

	hasher.Write([]byte(url))

	shortURL := base64.URLEncoding.EncodeToString(hasher.Sum(nil))
	if _, found := h.urlMap[shortURL]; !found {
		h.urlMap[shortURL] = url
		var baseURL = models.URL{URL: url, ShortURL: h.Cfg.BaseURL + "/" + shortURL}
		h.BaseFile.Producer.AddURL(&baseURL)
	}

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
