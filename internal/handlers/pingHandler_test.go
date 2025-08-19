package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestURLHandler_PingHandle(t *testing.T) {
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
			name: "ping handle positive test #1",
			want: want{
				statusCode:  http.StatusOK,
				request:     "http://localhost:8080/ping",
				location:    "",
				contentType: "text/plain",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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

		})
	}
}

func BenchmarkPingHandle(b *testing.B) {
	reader := strings.NewReader(``)
	client := &http.Client{}
	req := httptest.NewRequest(http.MethodGet, "http://localhost:8080/ping", reader)

	req.Header.Add("Content-Type", "text/plain")

	for i := 0; i < b.N; i++ {
		res, err := client.Do(req)
		if err != nil {
			return
		}
		defer res.Body.Close()
	}

}
