package handlers

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
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
	Storage      storage.Storage
}

func CreateHandle(cfg config.ShortenerConfig, store storage.Storage) URLHandler {
	var h URLHandler
	//h.URLMap = storage.GetDataFromFile(cfg.FileStoragePath)
	h.BaseURL = cfg.BaseURL
	//h.FileProducer = store.Producer
	h.DSN = cfg.DSN
	h.Storage = store
	return h
}

func (h *URLHandler) post(w http.ResponseWriter, r *http.Request, postKind string) {
	var (
		url  []byte
		err  error
		req  models.Request
		resp []byte
	)
	w.Header().Set("Content-Type", postKind)
	defer r.Body.Close()
	var buf bytes.Buffer
	_, err = buf.ReadFrom(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if postKind == contentTypeTextPlain {
		url = buf.Bytes()
	} else if postKind == contentTypeApJSON {
		if err = json.Unmarshal(buf.Bytes(), &req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		url = []byte(req.URL)
	}

	if len(url) == 0 {
		w.WriteHeader(http.StatusCreated)
		return
	}

	shortURL := h.Save(string(url), h.Storage)

	if postKind == contentTypeTextPlain {
		resp = []byte(h.BaseURL + shortURL)
	} else if postKind == contentTypeApJSON {
		var response models.Response
		response.Result = h.BaseURL + shortURL
		resp, err = json.Marshal(response)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", postKind)
	w.WriteHeader(http.StatusCreated)
	w.Write(resp)
}

func (h *URLHandler) PostHandle(w http.ResponseWriter, r *http.Request) {
	h.post(w, r, contentTypeTextPlain)
}

func (h *URLHandler) PostHandleJSON(w http.ResponseWriter, r *http.Request) {
	h.post(w, r, contentTypeApJSON)
}

func (h *URLHandler) Save(originalURL string, store storage.Storage) string {
	hasher := sha1.New()

	hasher.Write([]byte(originalURL))

	shortURL := base64.URLEncoding.EncodeToString(hasher.Sum(nil))

	if _, found := store.Load(shortURL); !found {
		fmt.Println(shortURL, originalURL)
		h.URLMap[shortURL] = originalURL
		var baseURL = models.URL{OriginalURL: originalURL, ShortURL: h.BaseURL + "/" + shortURL}
		store.Save(&baseURL)
	}

	return shortURL
}

func (h *URLHandler) GetHandle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", contentTypeTextPlain)
	shortURL := chi.URLParam(r, "shortURL")

	var originalURL string
	if len(shortURL) != 0 {
		originalURL = h.Load(shortURL, h.Storage)
	} else {
		http.Error(w, "Empty value", http.StatusBadRequest)
		return
	}
	w.Header().Set("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func (h *URLHandler) Load(shortURL string, store storage.Storage) string {
	if res, found := store.Load(shortURL); found {
		return res
	} else {
		return ""
	}

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
