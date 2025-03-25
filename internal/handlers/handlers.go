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
	"github.com/scaranin/go-svc-short-url/internal/auth"
	"github.com/scaranin/go-svc-short-url/internal/config"
	"github.com/scaranin/go-svc-short-url/internal/models"
)

const (
	contentTypeTextPlain string = "text/plain"
	contentTypeApJSON    string = "application/json"
)

// Основная структура, содержащая настройки сервиса
type URLHandler struct {
	URLMap       map[string]string
	BaseURL      string
	FileProducer *models.Producer
	DSN          string
	Storage      models.Storage
	Auth         auth.AuthConfig
}

// CreateHandle инициализирует структуру URLHandler
func CreateHandle(cfg config.ShortenerConfig, store models.Storage, auth auth.AuthConfig) URLHandler {
	var h URLHandler
	h.BaseURL = cfg.BaseURL
	h.Storage = store
	h.DSN = cfg.DSN
	h.Auth = auth
	return h
}

// post - локальная функция, содержащаю общую логику для post handlers
func (h *URLHandler) post(w http.ResponseWriter, r *http.Request, postKind string) {
	var (
		url  []byte
		err  error
		req  models.Request
		resp []byte
		buf  bytes.Buffer
	)
	cookieR, err := r.Cookie(h.Auth.CookieName)
	cookieW, err := h.Auth.FillUserReturnCookie(cookieR)
	http.SetCookie(w, cookieW)
	fmt.Println(cookieW)
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
		w.WriteHeader(http.StatusConflict)
	} else {
		w.WriteHeader(http.StatusCreated)
	}

	w.Write(resp)
}

// PostHandle принимает originalURL и возвращает shortURL. Формат text
func (h *URLHandler) PostHandle(w http.ResponseWriter, r *http.Request) {
	h.post(w, r, contentTypeTextPlain)
}

// PostHandleJSON принимает originalURL и возвращает shortURL. Формат JSON
func (h *URLHandler) PostHandleJSON(w http.ResponseWriter, r *http.Request) {
	h.post(w, r, contentTypeApJSON)
}

// PostHandleJSONBatch групповая загрузка данных в хранилище
func (h *URLHandler) PostHandleJSONBatch(w http.ResponseWriter, r *http.Request) {
	var (
		data         []byte
		err          error
		pairRequest  []models.PairRequest
		pairResponse []models.PairResponse
		resp         []byte
		buf          bytes.Buffer
	)
	cookieR, err := r.Cookie(h.Auth.CookieName)
	cookieW, err := h.Auth.FillUserReturnCookie(cookieR)
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

	resp, err = json.Marshal(pairResponse)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", contentTypeApJSON)
	http.SetCookie(w, cookieW)
	w.WriteHeader(http.StatusCreated)
	w.Write(resp)
}

// ShortURLCalc функция расчета сокращенного URL
func ShortURLCalc(originalURL string) string {
	hasher := sha1.New()
	hasher.Write([]byte(originalURL))
	return base64.URLEncoding.EncodeToString(hasher.Sum(nil))
}

// Save функция, добавляющая в хранилище новую запись
func (h *URLHandler) Save(originalURL string, correlationID string) (string, error) {

	shortURL := ShortURLCalc(originalURL)
	var baseURL = models.URL{CorrelationID: correlationID, OriginalURL: originalURL, ShortURL: shortURL, UserID: h.Auth.UserID}
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

// PingHandle проверяет доступность подключения к БД
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

func (h *URLHandler) GetUserURLs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", contentTypeApJSON)
	var (
		cookieW *http.Cookie
		err     error
	)
	cookieR, err := r.Cookie(h.Auth.CookieName)
	if err != nil {
		cookieW, err = h.Auth.FillUserReturnCookie(cookieR)
		//http.Error(w, err.Error(), http.StatusUnauthorized)
		//return
	}
	fmt.Println(cookieR.Value)
	cookieW, err = h.Auth.FillUserReturnCookie(cookieR)
	fmt.Println(cookieW)
	if err == http.ErrNoCookie {
		http.Error(w, err.Error(), http.StatusNoContent)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	fmt.Println("userid", h.Auth.UserID)
	URLList, err := h.Storage.GetUserURLList(h.Auth.UserID)
	if err != nil || len(URLList) == 0 {
		http.Error(w, err.Error(), http.StatusNoContent)
		return
	}

	for i := range URLList {
		URLList[i].ShortURL = h.BaseURL + URLList[i].ShortURL
	}

	URLUserListJSON, err := json.Marshal(URLList)
	if err != nil {
		fmt.Println(err)
		http.Error(w, err.Error(), http.StatusNoContent)
		return
	}

	http.SetCookie(w, cookieW)
	w.WriteHeader(http.StatusOK)
	w.Write(URLUserListJSON)

}
