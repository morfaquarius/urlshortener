package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestShortenHandler(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		wantStatus int
	}{
		{name: "valid URL", body: `{"url":"https://example.com/a?b=c"}`, wantStatus: http.StatusOK},
		{name: "malformed JSON", body: `{"url":`, wantStatus: http.StatusBadRequest},
		{name: "invalid URL", body: `{"url":"not-a-url"}`, wantStatus: http.StatusBadRequest},
		{name: "multiple JSON values", body: `{"url":"https://example.com"} {}`, wantStatus: http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewHandler(NewURLShortener())
			request := httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader(tt.body))
			response := httptest.NewRecorder()

			handler.ServeHTTP(response, request)
			if response.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body: %s", response.Code, tt.wantStatus, response.Body)
			}

			if tt.wantStatus != http.StatusOK {
				return
			}
			var body shortenResponse
			if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if body.OriginalURL != "https://example.com/a?b=c" {
				t.Errorf("original_url = %q", body.OriginalURL)
			}
			if len(body.ShortURL) < 6 || len(body.ShortURL) > 8 {
				t.Errorf("short_url length = %d, want 6-8", len(body.ShortURL))
			}
		})
	}
}

func TestRedirectHandler(t *testing.T) {
	shortener := NewURLShortener()
	shortID, err := shortener.Shorten("https://example.com/long/path?q=1")
	if err != nil {
		t.Fatalf("Shorten() error = %v", err)
	}
	handler := NewHandler(shortener)

	tests := []struct {
		name       string
		path       string
		wantStatus int
		wantURL    string
	}{
		{name: "existing short URL", path: "/" + shortID, wantStatus: http.StatusFound, wantURL: "https://example.com/long/path?q=1"},
		{name: "missing short URL", path: "/missing", wantStatus: http.StatusNotFound},
		{name: "nested path", path: "/missing/extra", wantStatus: http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, tt.path, nil)
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)

			if response.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", response.Code, tt.wantStatus)
			}
			if gotURL := response.Header().Get("Location"); gotURL != tt.wantURL {
				t.Errorf("Location = %q, want %q", gotURL, tt.wantURL)
			}
		})
	}
}
