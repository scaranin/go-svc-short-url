package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_postHandle(t *testing.T) {
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
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.want.request))
			req.Header.Set("content-type", "text/plain")
			rec := httptest.NewRecorder()
			postHandle(rec, req)

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

func Test_routeHandle(t *testing.T) {
	type want struct {
		statusCode  int
		response    string
		contentType string
		method      string
	}
	tests := []struct {
		name string
		want want
	}{
		{
			name: "route handle negative test #1",
			want: want{
				statusCode:  http.StatusBadRequest,
				response:    "Only post and get requests are allowed!\n",
				contentType: "text/plain",
				method:      http.MethodPut,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.want.method, "/", nil)

			req.Header.Set("content-type", "text/plain")

			rec := httptest.NewRecorder()

			routeHandle(rec, req)

			res := rec.Result()

			defer res.Body.Close()
			assert.Equal(t, tt.want.statusCode, res.StatusCode)

			resBody, err := io.ReadAll(res.Body)

			require.NoError(t, err)
			assert.Equal(t, tt.want.response, string(resBody))
			assert.Contains(t, res.Header.Get("content-type"), tt.want.contentType)
		})
	}
}
