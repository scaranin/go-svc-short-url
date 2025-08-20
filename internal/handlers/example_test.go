package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
)

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

func ExampleURLHandler_PingHandle() {
	reader := strings.NewReader(``)
	client := &http.Client{}
	req := httptest.NewRequest(http.MethodGet, "http://localhost:8080/ping", reader)

	req.Header.Add("Content-Type", "text/plain")

	res, err := client.Do(req)
	if err != nil {
		return
	}
	defer res.Body.Close()
}
