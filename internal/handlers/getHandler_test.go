package handlers

import (
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/scaranin/go-svc-short-url/internal/auth"
	"github.com/scaranin/go-svc-short-url/internal/config"
	"github.com/scaranin/go-svc-short-url/internal/storage"
	"github.com/stretchr/testify/assert"
)

func TestURLHandler_GetHandle(t *testing.T) {
	type want struct {
		statusCode  int
		request     string
		location    string
		contentType string
	}
	tests := []struct {
		name string
		want want
	}{
		{
			name: "get handle negative test #1",
			want: want{
				statusCode:  http.StatusBadRequest,
				request:     "http://localhost:8080",
				location:    "",
				contentType: "text/plain",
			},
		},

		{
			name: "get handle positive test #1",
			want: want{
				statusCode:  http.StatusTemporaryRedirect,
				request:     "http://localhost:8080/pkmdI_i-nYcS6P7hSfjTtWUmfcA=",
				location:    "https://practicum.yandex.ru/",
				contentType: "text/plain",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := config.CreateConfig()
			if err != nil {
				return
			}
			store, err := storage.CreateStoreFile(cfg.FileStoragePath)
			if err != nil {
				return
			}
			defer store.Close()
			h1 := CreateHandle(cfg, store, auth.NewAuthConfig())

			reqPost := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.want.location))
			reqPost.Header.Set("Content-Type", tt.want.contentType)
			recPost := httptest.NewRecorder()
			h1.PostHandle(recPost, reqPost)

			reader := strings.NewReader(``)
			client := &http.Client{}
			req := httptest.NewRequest(http.MethodGet, tt.want.request, reader)

			req.Header.Add("Content-Type", tt.want.contentType)

			res, err := client.Do(req)
			if err != nil {
				return
			}
			defer res.Body.Close()
			assert.Equal(t, tt.want.statusCode, res.StatusCode)
			assert.Equal(t, tt.want.location, res.Header.Get("Location"))
		})
	}

}

func BenchmarkGetHandle(b *testing.B) {
	cfg, err := config.CreateConfig()
	if err != nil {
		log.Fatal(err)
	}

	store, err := config.CreateStore(cfg)
	if err != nil {
		log.Fatal(err)
	}

	auth := auth.NewAuthConfig()

	h := CreateHandle(cfg, store, auth)

	req := httptest.NewRequest(http.MethodGet, h.BaseURL, nil)

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
		h.post(w, req, "text/plain")
	}

}

func ExampleURLHandler_GetHandle() {
	reader := strings.NewReader(``)
	client := &http.Client{}
	req := httptest.NewRequest(http.MethodGet, "http://localhost:8080/pkmdI_i-nYcS6P7hSfjTtWUmfcA=", reader)

	req.Header.Add("Content-Type", "text/plain")

	res, err := client.Do(req)
	if err != nil {
		return
	}
	defer res.Body.Close()
}
