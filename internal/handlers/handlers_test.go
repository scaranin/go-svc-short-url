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
			name: "get handle negative test #1",
			want: want{
				statusCode:  http.StatusBadRequest,
				request:     "http://localhost:8080",
				location:    "",
				contentType: "text/plain",
			},
		},
		/*
			{
				name: "get handle positive test #1",
				want: want{
					statusCode:  http.StatusTemporaryRedirect,
					request:     "http://localhost:8080/pkmdI_i-nYcS6P7hSfjTtWUmfcA=",
					location:    "https://practicum.yandex.ru/",
					contentType: "text/plain",
				},
			},
		*/
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h1 := CreateConfig()
			reqPost := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.want.location))
			reqPost.Header.Set("content-type", tt.want.contentType)
			recPost := httptest.NewRecorder()
			h1.PostHandle(recPost, reqPost)

			req := httptest.NewRequest(http.MethodGet, tt.want.request, nil)
			req.Header.Set("content-type", tt.want.contentType)
			rec := httptest.NewRecorder()
			h1.GetHandle(rec, req)
			res := rec.Result()
			defer res.Body.Close()
			assert.Equal(t, tt.want.statusCode, res.StatusCode)
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
			name: "post handle negative test #1",
			want: want{
				statusCode:  http.StatusBadRequest,
				request:     "",
				response:    "Empty value\n",
				contentType: "text/plain",
			},
		},
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

	h := CreateConfig()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.want.request))
			req.Header.Set("content-type", tt.want.contentType)
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
	}{{
		name: "post json handle negative test #1",
		want: want{
			statusCode:  http.StatusBadRequest,
			request:     `{"url": "https://practicum.yandex.ru"}`,
			response:    "text/plain not supported\n",
			contentType: "text/plain",
		},
	},
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
	h := CreateConfig()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(tt.want.request))
			req.Header.Set("content-type", tt.want.contentType)
			rec := httptest.NewRecorder()
			h.PostHandleJson(rec, req)

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
