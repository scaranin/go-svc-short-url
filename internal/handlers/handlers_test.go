package handlers

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			name: "get handle positive test #1",
			want: want{
				statusCode:  http.StatusTemporaryRedirect,
				request:     "http://localhost:8080/pkmdI_i-nYcS6P7hSfjTtWUmfcA=",
				location:    "https://practicum.yandex.ru/",
				contentType: "text/plain",
			},
		},
		{
			name: "get handle negative test #1",
			want: want{
				statusCode:  http.StatusBadRequest,
				request:     "http://localhost:8080/",
				location:    "",
				contentType: "text/plain",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := CreateConfig()
			reqPost := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.want.location))
			reqPost.Header.Set("content-type", "text/plain")
			recPost := httptest.NewRecorder()
			h.PostHandle(recPost, reqPost)

			req := httptest.NewRequest(http.MethodGet, tt.want.request, nil)
			req.Header.Set("content-type", "text/plain")
			rec := httptest.NewRecorder()
			h.GetHandle(rec, req)

			res := rec.Result()
			assert.Equal(t, tt.want.statusCode, res.StatusCode)
			defer res.Body.Close()

			assert.Equal(t, tt.want.location, res.Header.Get("Location"))
		})
	}
}

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
		{
			name: "post handle positive test #1",
			want: want{
				statusCode:  http.StatusCreated,
				request:     "https://practicum.yandex.ru/",
				response:    "http://localhost:8080/pkmdI_i-nYcS6P7hSfjTtWUmfcA=",
				contentType: "text/plain",
			},
		},
		{
			name: "post handle negative test #1",
			want: want{
				statusCode:  http.StatusBadRequest,
				request:     "",
				response:    "Empty value\n",
				contentType: "text/plain",
			},
		},
	}
	h := CreateConfig()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.want.request))
			req.Header.Set("content-type", "text/plain")
			rec := httptest.NewRecorder()
			h.PostHandle(rec, req)

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
