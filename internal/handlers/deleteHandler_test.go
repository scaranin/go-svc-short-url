package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestURLHandler_DeleteHandle(t *testing.T) {
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
			name: "delete handle positive test #1",
			want: want{
				statusCode:  http.StatusBadRequest,
				request:     "http://localhost:8080/api/user/urls",
				location:    "",
				contentType: "text/plain",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(``)
			client := &http.Client{}
			req := httptest.NewRequest(http.MethodDelete, tt.want.request, reader)

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

func ExampleURLHandler_DeleteHandle() {
	reader := strings.NewReader(``)
	client := &http.Client{}
	req := httptest.NewRequest(http.MethodDelete, "http://localhost:8080/api/user/urls", reader)

	req.Header.Add("Content-Type", "text/plain")

	res, err := client.Do(req)
	if err != nil {
		return
	}
	defer res.Body.Close()
}
