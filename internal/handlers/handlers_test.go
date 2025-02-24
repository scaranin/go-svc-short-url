package handlers

import (
	"net/http"
	"testing"
)

func TestURLHandler_PostHandle(t *testing.T) {
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name string
		h    *URLHandler
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.h.PostHandle(tt.args.w, tt.args.r)
		})
	}
}

func TestURLHandler_GetHandle(t *testing.T) {
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name string
		h    *URLHandler
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.h.GetHandle(tt.args.w, tt.args.r)
		})
	}
}
