package services

import (
	"sync"

	"agent/internal/repositories"
)

// URLService resolves short URL codes to full URLs
type URLService struct {
	urlRepo  *repositories.URLShortenerRepository
	memCache map[string]string
	mu       sync.RWMutex
}

// NewURLService creates a new URL service
func NewURLService(urlRepo *repositories.URLShortenerRepository) *URLService {
	return &URLService{
		urlRepo:  urlRepo,
		memCache: make(map[string]string),
	}
}

// ResolveURL looks up the full URL from a short code
func (s *URLService) ResolveURL(code string) string {
	s.mu.RLock()
	if url, ok := s.memCache[code]; ok {
		s.mu.RUnlock()
		return url
	}
	s.mu.RUnlock()

	if s.urlRepo == nil {
		return ""
	}

	url, err := s.urlRepo.Lookup(code)
	if err != nil || url == "" {
		return ""
	}

	s.mu.Lock()
	s.memCache[code] = url
	s.mu.Unlock()

	return url
}

// StoreURL caches a URL mapping (called from CacheTools for cross-instance sync)
func (s *URLService) StoreURL(code, url string) {
	s.mu.Lock()
	s.memCache[code] = url
	s.mu.Unlock()
}
