package handlers

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/scaranin/go-svc-short-url/internal/models"
)

// post is an internal helper function that handles the logic for creating a single short URL.
// It is designed to be called by public-facing handlers like PostHandle and PostHandleJSON.
// It orchestrates user authentication, request parsing based on the `postKind` content type,
// saving the URL, and formatting the response.
//
// A key feature is its ability to handle database conflicts: if a unique constraint
// violation occurs, it returns an HTTP 409 Conflict status. Otherwise, it returns
// HTTP 201 Created on success.
func (h *URLHandler) post(w http.ResponseWriter, r *http.Request, postKind string) {
	var (
		url  []byte
		err  error
		req  models.Request
		resp []byte
		buf  bytes.Buffer
	)
	cookieR, err := r.Cookie(h.Auth.CookieName)
	if err != nil {
		log.Print(err.Error())
	}
	cookieW, err := h.Auth.FillUserReturnCookie(cookieR)
	if err != nil {
		log.Print(err.Error())
	}
	http.SetCookie(w, cookieW)

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

// PostHandle handles requests to create a short URL from a plain text body.
// It delegates the core logic to the `post` helper, specifying `contentTypeTextPlain`.
func (h *URLHandler) PostHandle(w http.ResponseWriter, r *http.Request) {
	h.post(w, r, contentTypeTextPlain)
}

// PostHandleJSON handles requests to create a short URL from a JSON request body.
// The expected JSON format is `{"url":"<your_url>"}`.
// It delegates the core logic to the `post` helper, specifying `contentTypeApJSON`.
func (h *URLHandler) PostHandleJSON(w http.ResponseWriter, r *http.Request) {
	h.post(w, r, contentTypeApJSON)
}

// PostHandleJSONBatch handles requests to shorten multiple URLs in a single batch operation.
// It expects a JSON array of objects, each with a `correlation_id` and an `original_url`.
// It authenticates the user, processes each URL, and returns a JSON array of corresponding
// objects with the `correlation_id` and the new `short_url`.
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
	if err != nil {
		log.Print(err.Error())
	}
	cookieW, err := h.Auth.FillUserReturnCookie(cookieR)
	if err != nil {
		log.Print(err.Error())
	}
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
