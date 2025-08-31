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
	h.handleCookies(w, r)

	w.Header().Set("Content-Type", postKind)
	defer r.Body.Close()

	url, err := h.parseRequestBody(r, postKind)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(url) == 0 {
		w.WriteHeader(http.StatusCreated)
		return
	}

	resp, statusCode, err := h.saveURLAndBuildResponse(url, postKind)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(statusCode)
	_, _ = w.Write(resp)
}

// handleCookies reads the cookie from the request, updates it via Auth,
// and sets it in the response. Errors are logged but do not interrupt execution.
func (h *URLHandler) handleCookies(w http.ResponseWriter, r *http.Request) {
	cookieR, err := r.Cookie(h.Auth.CookieName)
	if err != nil {
		log.Print(err.Error())
	}
	cookieW, err := h.Auth.FillUserReturnCookie(cookieR)
	if err != nil {
		log.Print(err.Error())
	}
	http.SetCookie(w, cookieW)
}

// parseRequestBody reads and parses the HTTP request body,
// returning the URL as a byte slice.
// Supports two content types:
// - contentTypeTextPlain: returns the raw request body;
// - contentTypeApJSON: parses JSON and extracts the URL field.
// Returns an error if parsing fails.
func (h *URLHandler) parseRequestBody(r *http.Request, postKind string) ([]byte, error) {
	var buf bytes.Buffer
	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		return nil, err
	}

	if postKind == contentTypeTextPlain {
		return buf.Bytes(), nil
	} else if postKind == contentTypeApJSON {
		var req models.Request
		if err := json.Unmarshal(buf.Bytes(), &req); err != nil {
			return nil, err
		}
		return []byte(req.URL), nil
	}
	return nil, nil
}

// saveURLAndBuildResponse saves the URL in storage and builds the HTTP response body.
// Returns the response body, HTTP status code, and an error if any occurs.
// Handles database unique constraint violations by returning HTTP 409 Conflict.
// Response format depends on postKind:
// - contentTypeTextPlain: returns the short URL as plain text;
// - contentTypeApJSON: returns JSON containing the short URL in the "Result" field.
func (h *URLHandler) saveURLAndBuildResponse(url []byte, postKind string) ([]byte, int, error) {
	shortURL, pgErr := h.Save(string(url), "")

	var resp []byte
	if postKind == contentTypeTextPlain {
		resp = []byte(h.BaseURL + shortURL)
	} else if postKind == contentTypeApJSON {
		response := models.Response{Result: h.BaseURL + shortURL}
		var err error
		resp, err = json.Marshal(response)
		if err != nil {
			return nil, http.StatusInternalServerError, err
		}
	}

	if pgErr != nil {
		if pgError, ok := pgErr.(*pgconn.PgError); ok && pgError.Code == pgerrcode.UniqueViolation {
			return resp, http.StatusConflict, nil
		}
	}

	return resp, http.StatusCreated, nil
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
