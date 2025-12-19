package services

import (
	"encoding/json"
	"sync"

	"github.com/pina-colada-co/agent-go/internal/repositories"
)

// URLService resolves short URL codes to full URLs
type URLService struct {
	cacheRepo *repositories.ResearchCacheRepository
	memCache  map[string]string
	mu        sync.RWMutex
}

// NewURLService creates a new URL service
func NewURLService(cacheRepo *repositories.ResearchCacheRepository) *URLService {
	return &URLService{
		cacheRepo: cacheRepo,
		memCache:  make(map[string]string),
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

	if s.cacheRepo == nil {
		return ""
	}

	cached, err := s.cacheRepo.Lookup(0, "url:"+code)
	if err != nil || cached == nil {
		return ""
	}

	var url string
	if err := json.Unmarshal(cached.ResultData, &url); err != nil {
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
