package tools

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/pina-colada-co/agent-go/internal/repositories"
)

// CacheRepositoryInterface defines the interface for cache operations
type CacheRepositoryInterface interface {
	Lookup(tenantID int64, cacheKey string) (*repositories.ResearchCacheDTO, error)
	Upsert(input repositories.CacheUpsertInput) error
	ListRecent(tenantID int64, cacheType string, limit int) ([]repositories.ResearchCacheDTO, error)
}

// MemCacheEntry holds an in-memory cached result
type MemCacheEntry struct {
	Data  string
	Count int
}

type CacheTools struct {
	cacheRepo CacheRepositoryInterface

	// Per-request dedup
	seenKeys   map[string]bool
	seenKeysMu sync.Mutex

	// In-memory cache (session-scoped, checked before DB)
	memCache   map[string]*MemCacheEntry
	memCacheMu sync.RWMutex

	// Context for tenant/user
	tenantID int64
	userID   int64
}

func NewCacheTools(cacheRepo CacheRepositoryInterface) *CacheTools {
	return &CacheTools{
		cacheRepo: cacheRepo,
		seenKeys:  make(map[string]bool),
		memCache:  make(map[string]*MemCacheEntry),
	}
}

// SetContext sets the tenant and user context for cache operations
func (t *CacheTools) SetContext(tenantID, userID int64) {
	t.tenantID = tenantID
	t.userID = userID
}

// ResetRequestState clears per-request dedup state
func (t *CacheTools) ResetRequestState() {
	t.seenKeysMu.Lock()
	t.seenKeys = make(map[string]bool)
	t.seenKeysMu.Unlock()
}

// MarkSeen marks a key as seen for this request, returns false if already seen
func (t *CacheTools) MarkSeen(key string) bool {
	t.seenKeysMu.Lock()
	defer t.seenKeysMu.Unlock()

	if t.seenKeys[key] {
		return false
	}
	t.seenKeys[key] = true
	return true
}

// getFromMemory retrieves a cached entry from in-memory cache
func (t *CacheTools) getFromMemory(cacheKey string) *MemCacheEntry {
	t.memCacheMu.RLock()
	defer t.memCacheMu.RUnlock()
	return t.memCache[cacheKey]
}

// storeInMemory stores an entry in the in-memory cache
func (t *CacheTools) storeInMemory(cacheKey string, data string, count int) {
	t.memCacheMu.Lock()
	defer t.memCacheMu.Unlock()
	t.memCache[cacheKey] = &MemCacheEntry{Data: data, Count: count}
}

// LookupCache checks memory first, then DB. Returns nil if not found.
func (t *CacheTools) LookupCache(cacheType, query string) *MemCacheEntry {
	cacheKey := MakeCacheKey(cacheType, query)

	// Check memory first
	if mem := t.getFromMemory(cacheKey); mem != nil {
		preview := truncatePreview(mem.Data, 200)
		log.Printf("[ResearchCache] MEM HIT type=%s query=%q count=%d preview=%s", cacheType, query, mem.Count, preview)
		return mem
	}

	// Fall back to database
	if t.cacheRepo == nil {
		log.Printf("[ResearchCache] MISS type=%s query=%q (no DB)", cacheType, query)
		return nil
	}

	cached, err := t.cacheRepo.Lookup(t.tenantID, cacheKey)
	if err != nil {
		log.Printf("[ResearchCache] DB LOOKUP ERROR: %v", err)
		return nil
	}
	if cached == nil {
		log.Printf("[ResearchCache] MISS type=%s query=%q", cacheType, query)
		return nil
	}

	// Found in DB - unmarshal JSON string back to plain string
	var data string
	if err := json.Unmarshal(cached.ResultData, &data); err != nil {
		log.Printf("[ResearchCache] DB UNMARSHAL ERROR: %v", err)
		return nil
	}

	// Store in memory for faster subsequent lookups
	t.storeInMemory(cacheKey, data, cached.ResultCount)

	age := formatAge(cached.CreatedAt)
	preview := truncatePreview(data, 200)
	log.Printf("[ResearchCache] DB HIT type=%s query=%q count=%d age=%s preview=%s", cacheType, query, cached.ResultCount, age, preview)

	return &MemCacheEntry{Data: data, Count: cached.ResultCount}
}

// StoreCache stores in memory and DB
func (t *CacheTools) StoreCache(cacheType, query, data string, count int) {
	cacheKey := MakeCacheKey(cacheType, query)

	// Always store in memory
	t.storeInMemory(cacheKey, data, count)

	// Store in DB for persistence
	if t.cacheRepo == nil {
		log.Printf("[ResearchCache] MEM STORED type=%s query=%q count=%d (no DB)", cacheType, query, count)
		return
	}

	queryParams, _ := json.Marshal(map[string]string{"query": query, "cache_type": cacheType})
	dataJSON, _ := json.Marshal(data) // Encode string as valid JSON
	input := repositories.CacheUpsertInput{
		TenantID:    t.tenantID,
		CacheKey:    cacheKey,
		CacheType:   cacheType,
		QueryParams: queryParams,
		ResultData:  dataJSON,
		ResultCount: count,
		CreatedBy:   &t.userID,
		ExpiresAt:   time.Now().Add(getTTL(cacheType, 0)),
	}

	if err := t.cacheRepo.Upsert(input); err != nil {
		log.Printf("[ResearchCache] DB STORE ERROR: %v", err)
		return
	}

	log.Printf("[ResearchCache] STORED type=%s query=%q count=%d", cacheType, query, count)
}

// --- Tool Parameters and Results ---

type ResearchCacheParams struct {
	Action    string `json:"action" jsonschema:"required,enum=lookup|store|list_recent,description=Action: lookup (check cache), store (save results), list_recent (show recent searches)"`
	CacheType string `json:"cache_type" jsonschema:"description=Type of cache: job_search, web_fetch, company_research"`
	Query     string `json:"query" jsonschema:"description=Search query or URL to lookup/store"`
	Data      string `json:"data" jsonschema:"description=JSON data to store (required for store action)"`
	TTLHours  int    `json:"ttl_hours" jsonschema:"description=Time-to-live in hours (default: 24 for job_search, 1 for web_fetch, 168 for company_research)"`
}

type ResearchCacheResult struct {
	Results string `json:"results"`
	Found   bool   `json:"found"`
	Count   int    `json:"count"`
}

// --- Tool Method ---

func (t *CacheTools) ResearchCacheCtx(ctx context.Context, params ResearchCacheParams) (*ResearchCacheResult, error) {
	if t.cacheRepo == nil {
		return &ResearchCacheResult{Results: "Cache not available"}, nil
	}

	switch params.Action {
	case "lookup":
		return t.lookup(params)
	case "store":
		return t.store(params)
	case "list_recent":
		return t.listRecent(params)
	default:
		return &ResearchCacheResult{Results: "Invalid action. Use: lookup, store, or list_recent"}, nil
	}
}

func (t *CacheTools) lookup(params ResearchCacheParams) (*ResearchCacheResult, error) {
	if params.Query == "" {
		return &ResearchCacheResult{Results: "Query is required for lookup"}, nil
	}

	log.Printf("[ResearchCache] LOOKUP type=%s query=%q", params.CacheType, params.Query)

	cacheKey := MakeCacheKey(params.CacheType, params.Query)
	cached, err := t.cacheRepo.Lookup(t.tenantID, cacheKey)
	if err != nil {
		log.Printf("[ResearchCache] LOOKUP ERROR: %v", err)
		return &ResearchCacheResult{Results: fmt.Sprintf("Cache lookup error: %v", err)}, nil
	}

	if cached == nil {
		log.Printf("[ResearchCache] MISS type=%s query=%q", params.CacheType, params.Query)
		return &ResearchCacheResult{
			Results: fmt.Sprintf("No cached results for '%s' (type: %s)", params.Query, params.CacheType),
			Found:   false,
			Count:   0,
		}, nil
	}

	// Unmarshal JSON string back to plain string
	var data string
	if err := json.Unmarshal(cached.ResultData, &data); err != nil {
		return &ResearchCacheResult{Results: fmt.Sprintf("Cache data error: %v", err)}, nil
	}

	age := formatAge(cached.CreatedAt)
	preview := truncatePreview(data, 200)
	log.Printf("[ResearchCache] HIT type=%s query=%q count=%d age=%s preview=%s", params.CacheType, params.Query, cached.ResultCount, age, preview)

	return &ResearchCacheResult{
		Results: fmt.Sprintf("Found %d cached results (%s ago):\n%s", cached.ResultCount, age, data),
		Found:   true,
		Count:   cached.ResultCount,
	}, nil
}

func (t *CacheTools) store(params ResearchCacheParams) (*ResearchCacheResult, error) {
	if params.Query == "" {
		return &ResearchCacheResult{Results: "Query is required for store"}, nil
	}
	if params.Data == "" {
		return &ResearchCacheResult{Results: "Data is required for store"}, nil
	}

	cacheKey := MakeCacheKey(params.CacheType, params.Query)
	ttl := getTTL(params.CacheType, params.TTLHours)

	queryParams, _ := json.Marshal(map[string]string{
		"query":      params.Query,
		"cache_type": params.CacheType,
	})

	count := countResults(params.Data)
	dataJSON, _ := json.Marshal(params.Data) // Encode as valid JSON

	input := repositories.CacheUpsertInput{
		TenantID:    t.tenantID,
		CacheKey:    cacheKey,
		CacheType:   params.CacheType,
		QueryParams: queryParams,
		ResultData:  dataJSON,
		ResultCount: count,
		CreatedBy:   &t.userID,
		ExpiresAt:   time.Now().Add(ttl),
	}

	err := t.cacheRepo.Upsert(input)
	if err != nil {
		return &ResearchCacheResult{Results: fmt.Sprintf("Failed to store cache: %v", err)}, nil
	}

	return &ResearchCacheResult{
		Results: fmt.Sprintf("Stored %d results for '%s' (expires in %v)", count, params.Query, ttl),
		Found:   true,
		Count:   count,
	}, nil
}

func (t *CacheTools) listRecent(params ResearchCacheParams) (*ResearchCacheResult, error) {
	limit := 10
	caches, err := t.cacheRepo.ListRecent(t.tenantID, params.CacheType, limit)
	if err != nil {
		return &ResearchCacheResult{Results: fmt.Sprintf("Failed to list cache: %v", err)}, nil
	}

	if len(caches) == 0 {
		return &ResearchCacheResult{
			Results: "No recent cached searches",
			Found:   false,
			Count:   0,
		}, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Recent cached searches (%d):\n", len(caches)))
	for _, c := range caches {
		age := formatAge(c.CreatedAt)
		sb.WriteString(fmt.Sprintf("- [%s] %d results (%s ago)\n", c.CacheType, c.ResultCount, age))
	}

	return &ResearchCacheResult{
		Results: sb.String(),
		Found:   true,
		Count:   len(caches),
	}, nil
}

// --- Helpers ---

func MakeCacheKey(cacheType, query string) string {
	normalized := strings.ToLower(strings.TrimSpace(query))
	hash := sha256.Sum256([]byte(cacheType + ":" + normalized))
	return hex.EncodeToString(hash[:])
}

func getTTL(cacheType string, customHours int) time.Duration {
	if customHours > 0 {
		return time.Duration(customHours) * time.Hour
	}

	switch cacheType {
	case "web_fetch":
		return 24 * time.Hour
	default: // job_search, company_research, and others
		return 168 * time.Hour // 7 days
	}
}

func formatAge(t time.Time) string {
	d := time.Since(t)

	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh", int(d.Hours()))
	}
	return fmt.Sprintf("%dd", int(d.Hours()/24))
}

func countResults(data string) int {
	var arr []interface{}
	if err := json.Unmarshal([]byte(data), &arr); err == nil {
		return len(arr)
	}

	var obj map[string]interface{}
	if err := json.Unmarshal([]byte(data), &obj); err == nil {
		return 1
	}

	return 0
}

func truncatePreview(s string, maxLen int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\t", " ")
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
