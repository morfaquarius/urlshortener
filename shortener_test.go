package main

import (
	"errors"
	"testing"
)

func TestURLShortener_Shorten(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{name: "valid HTTP URL", url: "http://example.com", wantErr: false},
		{name: "valid HTTPS URL", url: "https://google.com/search?q=test", wantErr: false},
		{name: "URL with port and fragment", url: "https://example.com:8443/path#section", wantErr: false},
		{name: "missing scheme", url: "example.com/path", wantErr: true},
		{name: "unsupported scheme", url: "ftp://example.com/file", wantErr: true},
		{name: "missing host", url: "https:///path", wantErr: true},
		{name: "malformed URL", url: "http://[::1", wantErr: true},
		{name: "leading or trailing whitespace", url: " https://example.com ", wantErr: true},
		{name: "empty URL", url: "", wantErr: true},
	}

	shortener := NewURLShortener()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shortID, err := shortener.Shorten(tt.url)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Shorten() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				if !errors.Is(err, ErrInvalidURL) {
					t.Errorf("Shorten() error = %v, want ErrInvalidURL", err)
				}
				return
			}

			if len(shortID) < 6 || len(shortID) > 8 {
				t.Errorf("short ID length = %d, want 6-8: %q", len(shortID), shortID)
			}
			gotURL, err := shortener.GetOriginal(shortID)
			if err != nil {
				t.Fatalf("GetOriginal() error = %v", err)
			}
			if gotURL != tt.url {
				t.Errorf("GetOriginal() = %q, want %q", gotURL, tt.url)
			}
		})
	}
}

func TestURLShortener_GeneratesUniqueIDs(t *testing.T) {
	shortener := NewURLShortener()
	ids := make(map[string]struct{})

	for i := 0; i < 100; i++ {
		id, err := shortener.Shorten("https://example.com")
		if err != nil {
			t.Fatalf("Shorten() error = %v", err)
		}
		if _, exists := ids[id]; exists {
			t.Fatalf("duplicate short ID generated: %q", id)
		}
		ids[id] = struct{}{}
	}
}

func TestURLShortener_GetOriginalNotFound(t *testing.T) {
	_, err := NewURLShortener().GetOriginal("missing")
	if !errors.Is(err, ErrShortURLNotFound) {
		t.Fatalf("GetOriginal() error = %v, want ErrShortURLNotFound", err)
	}
}
