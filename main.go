package main

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strings"
)

type shortenRequest struct {
	URL string `json:"url"`
}

type shortenResponse struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

// NewHandler returns the HTTP API backed by the supplied URL shortener.
func NewHandler(shortener *URLShortener) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /shorten", func(w http.ResponseWriter, r *http.Request) {
		handleShorten(w, r, shortener)
	})
	mux.HandleFunc("GET /{short_url}", func(w http.ResponseWriter, r *http.Request) {
		shortID := r.PathValue("short_url")
		if strings.Contains(shortID, "/") {
			http.NotFound(w, r)
			return
		}
		handleRedirect(w, r, shortener, shortID)
	})
	return mux
}

func handleShorten(w http.ResponseWriter, r *http.Request, shortener *URLShortener) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	decoder := json.NewDecoder(r.Body)

	var request shortenRequest
	if err := decoder.Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON request")
		return
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		writeError(w, http.StatusBadRequest, "request must contain a single JSON object")
		return
	}

	shortID, err := shortener.Shorten(request.URL)
	if errors.Is(err, ErrInvalidURL) {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not shorten URL")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(shortenResponse{
		ShortURL:    shortID,
		OriginalURL: request.URL,
	}); err != nil {
		log.Printf("encode shorten response: %v", err)
	}
}

func handleRedirect(w http.ResponseWriter, r *http.Request, shortener *URLShortener, shortID string) {
	originalURL, err := shortener.GetOriginal(shortID)
	if errors.Is(err, ErrShortURLNotFound) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not retrieve URL")
		return
	}

	http.Redirect(w, r, originalURL, http.StatusFound)
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(map[string]string{"error": message}); err != nil {
		log.Printf("encode error response: %v", err)
	}
}

func main() {
	log.Fatal(http.ListenAndServe(":8080", NewHandler(NewURLShortener())))
}
