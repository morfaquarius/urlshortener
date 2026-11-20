package main

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"sync"
)

var (
	ErrInvalidURL       = errors.New("invalid URL: expected an absolute HTTP or HTTPS URL")
	ErrShortURLNotFound = errors.New("short URL not found")
)

type URLShortener struct {
	urls map[string]string
	mu   sync.RWMutex
}

func NewURLShortener() *URLShortener {
	return &URLShortener{
		urls: make(map[string]string),
	}
}

// Shorten validates and stores a URL under a unique short identifier.
func (us *URLShortener) Shorten(originalURL string) (string, error) {
	if !isValidURL(originalURL) {
		return "", ErrInvalidURL
	}

	us.mu.Lock()
	defer us.mu.Unlock()

	for {
		shortID := generateShortID()
		if _, exists := us.urls[shortID]; exists {
			continue
		}
		us.urls[shortID] = originalURL
		return shortID, nil
	}
}

// GetOriginal returns the URL associated with a short identifier.
func (us *URLShortener) GetOriginal(shortID string) (string, error) {
	us.mu.RLock()
	defer us.mu.RUnlock()

	originalURL, exists := us.urls[shortID]
	if !exists {
		return "", fmt.Errorf("%w: %q", ErrShortURLNotFound, shortID)
	}
	return originalURL, nil
}

// generateShortID returns eight URL-safe characters from six random bytes.
func generateShortID() string {
	randomBytes := make([]byte, 6)
	if _, err := rand.Read(randomBytes); err != nil {
		panic(fmt.Errorf("generate short URL identifier: %w", err))
	}
	return base64.RawURLEncoding.EncodeToString(randomBytes)
}

func isValidURL(value string) bool {
	if value == "" || strings.TrimSpace(value) != value {
		return false
	}

	parsed, err := url.Parse(value)
	if err != nil {
		return false
	}
	return parsed.IsAbs() &&
		(strings.EqualFold(parsed.Scheme, "http") || strings.EqualFold(parsed.Scheme, "https")) &&
		parsed.Hostname() != ""
}
