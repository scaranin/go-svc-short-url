package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
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
	if err := h.authAndSetCookie(w, r); err != nil {
		log.Print(err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	urlStr, err := h.extractURLFromBody(r, postKind)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if urlStr == "" {
		w.WriteHeader(http.StatusCreated)
		return
	}

	shortURL, err := h.Save(urlStr, "")
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == pgerrcode.UniqueViolation {
			http.Error(w, "URL already exists", http.StatusConflict)
			return
		}
		log.Printf("Database error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if err := h.writeResponse(w, postKind, h.BaseURL+shortURL); err != nil {
		log.Printf("Failed to write response: %v", err)
	}
}

// authAndSetCookie validates the authentication cookie and updates it in the response.
// It retrieves the existing cookie, fills it with user data, and sets the updated cookie.
//
// Parameters:
//   - w: http.ResponseWriter to set the cookie.
//   - r: *http.Request to read the existing cookie.
//
// Returns:
//   - error: nil if successful, otherwise an error describing the failure (e.g., missing cookie).
func (h *URLHandler) authAndSetCookie(w http.ResponseWriter, r *http.Request) error {
	cookieR, err := r.Cookie(h.Auth.CookieName)
	if err != nil {
		return fmt.Errorf("missing auth cookie: %w", err)
	}

	cookieW, err := h.Auth.FillUserReturnCookie(cookieR)
	if err != nil {
		return fmt.Errorf("failed to fill user cookie: %w", err)
	}

	http.SetCookie(w, cookieW)
	return nil
}

// extractURLFromBody reads the request body and extracts the URL based on the content type.
// It supports text/plain and application/json formats.
//
// Parameters:
//   - r: *http.Request containing the body to read.
//   - postKind: string indicating the content type.
//
// Returns:
//   - string: the extracted URL.
//   - error: nil if successful, otherwise an error.
func (h *URLHandler) extractURLFromBody(r *http.Request, postKind string) (string, error) {
	defer r.Body.Close()

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r.Body); err != nil {
		return "", fmt.Errorf("failed to read request body: %w", err)
	}

	switch postKind {
	case contentTypeTextPlain:
		return buf.String(), nil
	case contentTypeApJSON:
		var req models.Request
		if err := json.Unmarshal(buf.Bytes(), &req); err != nil {
			return "", fmt.Errorf("invalid JSON: %w", err)
		}
		return req.URL, nil
	default:
		return "", fmt.Errorf("unsupported Content-Type: %s", postKind)
	}
}

// writeResponse formats and sends the HTTP response in the specified content type.
// It sets the appropriate headers and writes the result to the response writer.
//
// Parameters:
//   - w: http.ResponseWriter to write the response.
//   - postKind: string indicating the content type.
//   - result: string containing the response data.
//
// Returns:
//   - error: nil if successful, otherwise an error.
func (h *URLHandler) writeResponse(w http.ResponseWriter, postKind, result string) error {
	w.Header().Set("Content-Type", postKind)
	w.WriteHeader(http.StatusCreated)

	switch postKind {
	case contentTypeTextPlain:
		_, err := w.Write([]byte(result))
		return err
	case contentTypeApJSON:
		resp := models.Response{Result: result}
		data, err := json.Marshal(resp)
		if err != nil {
			http.Error(w, "failed to marshal response", http.StatusInternalServerError)
			return err
		}
		_, err = w.Write(data)
		return err
	default:
		http.Error(w, "unsupported Content-Type", http.StatusUnsupportedMediaType)
		return fmt.Errorf("unsupported Content-Type")
	}
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
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
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
