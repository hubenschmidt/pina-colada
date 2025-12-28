package tools

import (
	"crypto/sha256"
	"log"
	"strings"
	"sync"
	"time"
)

// URLShortenerInterface defines the interface for URL shortening
type URLShortenerInterface interface {
	Store(code, fullURL string, ttl time.Duration) error
	Lookup(code string) (string, error)
}

// URLTools handles URL shortening for job listings
type URLTools struct {
	urlRepo URLShortenerInterface

	// In-memory cache for URL lookups
	memCache   map[string]string
	memCacheMu sync.RWMutex
}

// NewURLTools creates a new URLTools instance
func NewURLTools(urlRepo URLShortenerInterface) *URLTools {
	return &URLTools{
		urlRepo:  urlRepo,
		memCache: make(map[string]string),
	}
}

// --- URL Shortening ---

const base62Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

func base62Encode(data []byte) string {
	var result strings.Builder
	for _, b := range data {
		result.WriteByte(base62Chars[int(b)%62])
	}
	return result.String()
}

func generateShortCode(url string) string {
	hash := sha256.Sum256([]byte(url))
	return base62Encode(hash[:6])
}

// ShortenURL generates a short code and stores the URL mapping
func (t *URLTools) ShortenURL(fullURL string) string {
	code := generateShortCode(fullURL)

	t.memCacheMu.Lock()
	t.memCache[code] = fullURL
	t.memCacheMu.Unlock()

	go t.storeURLInDB(code, fullURL)
	return code
}

func (t *URLTools) storeURLInDB(code, fullURL string) {
	if t.urlRepo == nil {
		return
	}

	if err := t.urlRepo.Store(code, fullURL, 168*time.Hour); err != nil {
		log.Printf("URL shortener DB error: %v", err)
	}
}

// ResolveURL looks up the full URL from a short code
func (t *URLTools) ResolveURL(code string) string {
	t.memCacheMu.RLock()
	if url, ok := t.memCache[code]; ok {
		t.memCacheMu.RUnlock()
		return url
	}
	t.memCacheMu.RUnlock()

	if t.urlRepo == nil {
		return ""
	}

	url, err := t.urlRepo.Lookup(code)
	if err != nil || url == "" {
		return ""
	}

	t.memCacheMu.Lock()
	t.memCache[code] = url
	t.memCacheMu.Unlock()

	return url
}
