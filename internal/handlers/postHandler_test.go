package handlers_test

import (
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/scaranin/go-svc-short-url/internal/auth"
	"github.com/scaranin/go-svc-short-url/internal/config"
	"github.com/scaranin/go-svc-short-url/internal/handlers"
	"github.com/scaranin/go-svc-short-url/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestURLHandler_PostHandle(t *testing.T) {
	type want struct {
		statusCode  int
		request     string
		response    string
		contentType string
	}
	tests := []struct {
		name string
		want want
	}{
		/*
			{
				name: "post handle negative test #1",
				want: want{
					statusCode:  http.StatusBadRequest,
					request:     "",
					response:    "application/json not supported\n",
					contentType: "application/json",
				},
			},
		*/
		{
			name: "post handle positive test #1",
			want: want{
				statusCode:  http.StatusCreated,
				request:     "https://practicum.yandex.ru/",
				response:    "http://localhost:8080/pkmdI_i-nYcS6P7hSfjTtWUmfcA=",
				contentType: "text/plain",
			},
		},
	}

	cfg, err := config.CreateConfig()
	if err != nil {
		return
	}
	store, err := storage.CreateStoreFile(cfg.FileStoragePath)
	if err != nil {
		return
	}
	defer store.Close()
	h := handlers.CreateHandle(cfg, store, auth.NewAuthConfig())
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.want.request))
			req.Header.Set("Content-Type", tt.want.contentType)
			rec := httptest.NewRecorder()
			h.PostHandle(rec, req)

			res := rec.Result()
			assert.Equal(t, tt.want.statusCode, res.StatusCode)
			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)

			require.NoError(t, err)
			assert.Equal(t, tt.want.response, string(resBody))
			assert.Contains(t, res.Header.Get("Content-Type"), "text/plain")
		})
	}
}

func TestURLHandler_PostHandleJson(t *testing.T) {
	type want struct {
		statusCode  int
		request     string
		response    string
		contentType string
	}
	tests := []struct {
		name string
		want want
	}{
		/*{
			name: "post json handle negative test #1",
			want: want{
				statusCode:  http.StatusBadRequest,
				request:     `{"url": "https://practicum.yandex.ru"}`,
				response:    "text/plain not supported\n",
				contentType: "text/plain",
			},
		},*/
		{
			name: "post json handle positive test #1",
			want: want{
				statusCode:  http.StatusCreated,
				request:     `{"url": "https://practicum.yandex.ru"}`,
				response:    `{"result":"http://localhost:8080/7CwAhsKqdvt3oSw8T1fXFwxdMLY="}`,
				contentType: "application/json",
			},
		},
	}
	cfg, err := config.CreateConfig()
	if err != nil {
		return
	}
	store, err := storage.CreateStoreFile(cfg.FileStoragePath)
	if err != nil {
		return
	}
	defer store.Close()
	h := handlers.CreateHandle(cfg, store, auth.NewAuthConfig())
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(tt.want.request))
			req.Header.Set("content-type", tt.want.contentType)
			rec := httptest.NewRecorder()
			h.PostHandleJSON(rec, req)

			res := rec.Result()
			assert.Equal(t, tt.want.statusCode, res.StatusCode)
			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)

			require.NoError(t, err)
			assert.Equal(t, tt.want.response, string(resBody))
			assert.Contains(t, res.Header.Get("content-type"), tt.want.contentType)
		})
	}
}

func BenchmarkPost(b *testing.B) {
	cfg, err := config.CreateConfig()
	if err != nil {
		log.Fatal(err)
	}

	store, err := config.CreateStore(cfg)
	if err != nil {
		log.Fatal(err)
	}

	auth := auth.NewAuthConfig()

	h := handlers.CreateHandle(cfg, store, auth)

	req := httptest.NewRequest(http.MethodPost, h.BaseURL, nil)

	w := httptest.NewRecorder()

	resAuthToken, err := auth.BuildJWTString()
	if err != nil {
		log.Fatal(err)
	}

	cookie := &http.Cookie{
		Name:     auth.CookieName,
		Value:    resAuthToken,
		Expires:  time.Now().Add(auth.TokenExp),
		HttpOnly: true,
		Path:     "/",
	}
	req.AddCookie(cookie)

	for i := 0; i < b.N; i++ {
		h.PostHandleJSON(w, req)
	}

}
