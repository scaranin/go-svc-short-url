package handlers

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/scaranin/go-svc-short-url/internal/config"
	"github.com/scaranin/go-svc-short-url/internal/models"
	"github.com/scaranin/go-svc-short-url/internal/storage"
)

const (
	contentTypeTextPlain string = "text/plain"
	contentTypeApJSON    string = "application/json"
)

type URLHandler struct {
	URLMap       map[string]string
	BaseURL      string
	FileProducer *models.Producer
	DSN          string
}

func CreateHandle(cfg config.ShortenerConfig, store storage.BaseFileJSON) URLHandler {
	var h URLHandler
	h.URLMap = storage.GetDataFromFile(store.Consumer)
	h.BaseURL = cfg.BaseURL
	h.FileProducer = store.Producer
	h.DSN = cfg.DSN
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
	shortURL := h.Save(string(url))

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(h.BaseURL + shortURL))

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

	shortURL := h.Save(string(url))

	var res models.Response
	res.Result = h.BaseURL + shortURL
	resp, err := json.Marshal(res)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", contentTypeApJSON)
	w.WriteHeader(http.StatusCreated)
	w.Write(resp)
}

func (h *URLHandler) Save(url string) string {
	hasher := sha1.New()

	hasher.Write([]byte(url))

	shortURL := base64.URLEncoding.EncodeToString(hasher.Sum(nil))
	if _, found := h.URLMap[shortURL]; !found {
		h.URLMap[shortURL] = url
		var baseURL = models.URL{URL: url, ShortURL: h.BaseURL + "/" + shortURL}
		h.FileProducer.AddURL(&baseURL)
	}

	return shortURL
}

func (h *URLHandler) GetHandle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", contentTypeTextPlain)
	shortURL := chi.URLParam(r, "shortURL")

	var url string
	if len(shortURL) != 0 {
		url = h.Load(shortURL)
	} else {
		http.Error(w, "Empty value", http.StatusBadRequest)
		return
	}
	w.Header().Set("Location", url)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func (h *URLHandler) Load(shortURL string) string {
	return h.URLMap[shortURL]
}

func (h *URLHandler) PingHandle(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", contentTypeTextPlain)

	pool, err := pgxpool.New(r.Context(), h.DSN)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if pool.Ping(r.Context()) != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer pool.Close()

	w.WriteHeader(http.StatusOK)

}
