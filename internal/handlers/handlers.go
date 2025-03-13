package handlers

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/scaranin/go-svc-short-url/internal/config"
	"github.com/scaranin/go-svc-short-url/internal/models"
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
	Storage      models.Storage
}

func CreateHandle(cfg config.ShortenerConfig, store models.Storage) URLHandler {
	var h URLHandler
	h.BaseURL = cfg.BaseURL
	h.Storage = store
	h.DSN = cfg.DSN
	return h
}

func (h *URLHandler) post(w http.ResponseWriter, r *http.Request, postKind string) {
	var (
		url  []byte
		err  error
		req  models.Request
		resp []byte
		buf  bytes.Buffer
	)
	w.Header().Set("Content-Type", postKind)
	defer r.Body.Close()
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
	shortURL, pgErr := h.Save(string(url), "")

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
	pgError, ok := pgErr.(*pgconn.PgError)
	if ok && pgError.Code == pgerrcode.UniqueViolation {
		w.WriteHeader(http.StatusCreated)
	} else {
		w.WriteHeader(http.StatusCreated)
	}
	w.Write(resp)
}

func (h *URLHandler) PostHandle(w http.ResponseWriter, r *http.Request) {
	h.post(w, r, contentTypeTextPlain)
}

func (h *URLHandler) PostHandleJSON(w http.ResponseWriter, r *http.Request) {
	h.post(w, r, contentTypeApJSON)
}

func (h *URLHandler) PostHandleJSONBatch(w http.ResponseWriter, r *http.Request) {
	var (
		data         []byte
		err          error
		pairRequest  []models.PairRequest
		pairResponse []models.PairResponse
		resp         []byte
		buf          bytes.Buffer
	)
	_, err = buf.ReadFrom(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	data = buf.Bytes()

	if err := json.Unmarshal(data, &pairRequest); err != nil {
		log.Fatal("Error parsing JSON:", err)
	}

	for _, pair := range pairRequest {
		sourtURL, _ := h.Save(pair.OriginalURL, pair.CorrelationID)
		newPair := models.PairResponse{
			CorrelationID: pair.CorrelationID,
			ShortURL:      h.BaseURL + sourtURL,
		}
		pairResponse = append(pairResponse, newPair)
		var URL = models.URL{CorrelationID: pair.CorrelationID, OriginalURL: pair.OriginalURL, ShortURL: ShortURLCalc(pair.OriginalURL)}
		h.Storage.Save(&URL)
	}

	fmt.Println("pairResponse", pairResponse)
	resp, err = json.Marshal(pairResponse)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", contentTypeApJSON)
	w.WriteHeader(http.StatusCreated)
	w.Write(resp)

}

func ShortURLCalc(originalURL string) string {
	hasher := sha1.New()
	hasher.Write([]byte(originalURL))
	return base64.URLEncoding.EncodeToString(hasher.Sum(nil))
}

func (h *URLHandler) Save(originalURL string, correlationID string) (string, error) {
	shortURL := ShortURLCalc(originalURL)
	var baseURL = models.URL{CorrelationID: correlationID, OriginalURL: originalURL, ShortURL: shortURL}
	shortURL, err := h.Storage.Save(&baseURL)

	return shortURL, err
}

func (h *URLHandler) GetHandle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", contentTypeTextPlain)
	shortURL := chi.URLParam(r, "shortURL")
	var originalURL string
	var err error
	if len(shortURL) != 0 {
		originalURL, err = h.Load(shortURL)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		http.Error(w, "Empty value", http.StatusBadRequest)
		return
	}
	w.Header().Add("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func (h *URLHandler) Load(shortURL string) (string, error) {
	return h.Storage.Load(shortURL)

}

func (h *URLHandler) PingHandle(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", contentTypeTextPlain)
	pool, err := pgxpool.New(r.Context(), h.DSN)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Fatal(err)
		return
	}

	err = pool.Ping(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Fatal(err)
		return
	}
	defer pool.Close()

	w.WriteHeader(http.StatusOK)

}
