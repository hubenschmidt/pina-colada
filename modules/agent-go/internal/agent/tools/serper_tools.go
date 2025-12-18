package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

// JobServiceInterface defines the methods needed from JobService for filtering
type JobServiceInterface interface {
	GetLeads(statusNames []string, tenantID *int64) ([]JobInfo, error)
}

// JobInfo holds minimal job data needed for filtering
type JobInfo struct {
	JobTitle string
	Account  string
}

// SerperTools holds Serper API-based tools
type SerperTools struct {
	apiKey     string
	jobService JobServiceInterface
	cacheTools *CacheTools

	// Cache for applied jobs to avoid repeated DB queries within a turn
	appliedJobsCache     []JobInfo
	appliedJobsCacheTime time.Time
	cacheMu              sync.RWMutex

	// Semaphore to limit concurrent job searches
	searchSem chan struct{}

	// Per-turn call limit to prevent runaway searches
	callCount      int
	callCountMu    sync.Mutex
	totalResults   int      // Track total results across all searches in turn
	searchQueries  []string // Track queries used in this turn
}

const appliedJobsCacheTTL = 30 * time.Second
const maxConcurrentSearches = 3
const maxSearchesPerTurn = 3

// NewSerperTools creates Serper tools with API key and optional job service for filtering
func NewSerperTools(apiKey string, jobService JobServiceInterface, cacheTools *CacheTools) *SerperTools {
	return &SerperTools{
		apiKey:     apiKey,
		jobService: jobService,
		cacheTools: cacheTools,
		searchSem:  make(chan struct{}, maxConcurrentSearches),
	}
}

// ResetCallCount resets the per-turn call counter (call at start of each agent turn)
func (t *SerperTools) ResetCallCount() {
	t.callCountMu.Lock()
	t.callCount = 0
	t.totalResults = 0
	t.searchQueries = nil
	t.callCountMu.Unlock()
}

func (t *SerperTools) loadAppliedJobs() []JobInfo {
	if t.jobService == nil {
		return nil
	}

	// Check cache first (read lock)
	t.cacheMu.RLock()
	if time.Since(t.appliedJobsCacheTime) < appliedJobsCacheTTL && t.appliedJobsCache != nil {
		cached := t.appliedJobsCache
		t.cacheMu.RUnlock()
		return cached
	}
	t.cacheMu.RUnlock()

	// Cache miss - load from DB (write lock)
	t.cacheMu.Lock()
	defer t.cacheMu.Unlock()

	// Double-check after acquiring write lock
	if time.Since(t.appliedJobsCacheTime) < appliedJobsCacheTTL && t.appliedJobsCache != nil {
		return t.appliedJobsCache
	}

	jobs, err := t.jobService.GetLeads([]string{"applied", "do_not_apply"}, nil)
	if err != nil {
		log.Printf("Warning: could not load applied jobs for filtering: %v", err)
		return nil
	}

	t.appliedJobsCache = jobs
	t.appliedJobsCacheTime = time.Now()
	log.Printf("Loaded %d applied/do_not_apply jobs for filtering", len(jobs))
	return jobs
}

// --- Tool Parameter Structs ---

// JobSearchParams defines parameters for job search
type JobSearchParams struct {
	Query string `json:"query" jsonschema:"Job search query (e.g., 'Senior Software Engineer NYC startups')"`
}

// JobSearchResult is the result of a job search
type JobSearchResult struct {
	Results string `json:"results"`
	Count   int    `json:"count"`
}

// JobListing represents a single job for compact cache storage
// Uses short field names to minimize JSON size
type JobListing struct {
	C string `json:"c"` // Company
	T string `json:"t"` // Title
	U string `json:"u"` // URL
}

// --- Serper API Types ---

type serperRequest struct {
	Q   string `json:"q"`
	Num int    `json:"num,omitempty"` // Number of results to request (default 10)
}

type serperResponse struct {
	Organic []serperOrganicResult `json:"organic"`
}

type serperOrganicResult struct {
	Title   string `json:"title"`
	Link    string `json:"link"`
	Snippet string `json:"snippet"`
}

// --- Tool Functions ---

// JobSearchCtx searches for jobs using Serper API with domain exclusions.
// Filters out jobs already applied to or marked as do_not_apply.
// Limits to maxSearchesPerTurn calls per request and maxConcurrentSearches concurrent.
func (t *SerperTools) JobSearchCtx(ctx context.Context, params JobSearchParams) (*JobSearchResult, error) {
	// Check per-turn call limit first
	t.callCountMu.Lock()
	if t.callCount >= maxSearchesPerTurn {
		totalResults := t.totalResults
		queries := t.searchQueries
		t.callCountMu.Unlock()
		log.Printf("🔍 job_search: LIMIT REACHED (%d/%d calls) - model has %d total results from queries: %v",
			maxSearchesPerTurn, maxSearchesPerTurn, totalResults, queries)
		return &JobSearchResult{
			Results: fmt.Sprintf("Search limit reached (%d searches). You have %d results to choose from.", maxSearchesPerTurn, totalResults),
			Count:   0,
		}, nil
	}
	t.callCount++
	myCallNum := t.callCount // This search's number (1, 2, or 3)
	t.searchQueries = append(t.searchQueries, params.Query)
	t.callCountMu.Unlock()
	log.Printf("🔍 job_search [%d/%d]: STARTING query: %s", myCallNum, maxSearchesPerTurn, params.Query)

	// Acquire semaphore (limit concurrent searches)
	select {
	case t.searchSem <- struct{}{}:
		defer func() { <-t.searchSem }()
	case <-ctx.Done():
		return &JobSearchResult{Results: "Search cancelled: context deadline exceeded"}, nil
	}

	if t.apiKey == "" {
		log.Printf("SERPER_API_KEY not configured")
		return &JobSearchResult{Results: "Job search not configured. SERPER_API_KEY required."}, nil
	}

	// Check research cache first
	if cached := t.checkCache(params.Query); cached != nil {
		log.Printf("🔍 job_search [%d/%d]: CACHE HIT for query: %s", myCallNum, maxSearchesPerTurn, params.Query)
		return cached, nil
	}

	// Load applied/do_not_apply jobs for filtering
	appliedJobs := t.loadAppliedJobs()

	// Enhance query with exclusions
	enhancedQuery := enhanceJobQuery(params.Query)
	log.Printf("job_search query: %s", enhancedQuery)

	// Call Serper API - request 20 results (max allowed on standard tier)
	reqBody := serperRequest{Q: enhancedQuery, Num: 20}
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return &JobSearchResult{Results: fmt.Sprintf("Error encoding request: %v", err)}, nil
	}

	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequest("POST", "https://google.serper.dev/search", bytes.NewBuffer(jsonBody))
	if err != nil {
		return &JobSearchResult{Results: fmt.Sprintf("Error creating request: %v", err)}, nil
	}

	req.Header.Set("X-API-KEY", t.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Serper API error: %v", err)
		return &JobSearchResult{Results: fmt.Sprintf("Search failed: %v", err)}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("Serper API error: %d - %s", resp.StatusCode, string(body))
		return &JobSearchResult{Results: fmt.Sprintf("Search failed: HTTP %d", resp.StatusCode)}, nil
	}

	var serperResp serperResponse
	if err := json.NewDecoder(resp.Body).Decode(&serperResp); err != nil {
		return &JobSearchResult{Results: fmt.Sprintf("Error parsing response: %v", err)}, nil
	}

	// Extract structured listings (filtered)
	listings := extractListings(serperResp.Organic, 10, appliedJobs)
	listingCount := len(listings)

	// Update total results count
	t.callCountMu.Lock()
	t.totalResults += listingCount
	totalNow := t.totalResults
	t.callCountMu.Unlock()

	log.Printf("🔍 job_search [%d/%d]: COMPLETE - found %d results (running total: %d)",
		myCallNum, maxSearchesPerTurn, listingCount, totalNow)

	if listingCount == 0 {
		return &JobSearchResult{
			Results: "All job postings found have already been applied to or marked as 'do not apply'. Try a different search.",
			Count:   0,
		}, nil
	}

	// Store structured data in cache
	t.storeListingsCache(params.Query, listings)

	// Format for display
	return &JobSearchResult{
		Results: formatListings(listings),
		Count:   listingCount,
	}, nil
}

// enhanceJobQuery adds keywords and minimal exclusions to find company career pages
// Only exclude top job boards in query (Google ignores too many -site: operators)
func enhanceJobQuery(query string) string {
	topJobBoards := []string{
		"linkedin.com",
		"indeed.com",
		"glassdoor.com",
		"ziprecruiter.com",
		"monster.com",
		"dice.com",
		"builtin.com",
		"jobright.ai",
		"motionrecruitment.com",
	}

	var exclusions []string
	for _, domain := range topJobBoards {
		exclusions = append(exclusions, fmt.Sprintf("-site:%s", domain))
	}

	return fmt.Sprintf(`%s careers OR jobs %s`, query, strings.Join(exclusions, " "))
}

// extractListings extracts structured job listings from organic results
func extractListings(organic []serperOrganicResult, maxResults int, appliedJobs []JobInfo) []JobListing {
	var listings []JobListing

	for _, item := range organic {
		if len(listings) >= maxResults {
			return listings
		}
		listing := extractListing(item, appliedJobs)
		if listing != nil {
			listings = append(listings, *listing)
		}
	}

	return listings
}

func extractListing(item serperOrganicResult, appliedJobs []JobInfo) *JobListing {
	company, title := extractCompanyFromTitle(item.Title)
	if company == "" {
		company = extractCompanyFromURL(item.Link)
	}
	if matchesAppliedJob(company, title, appliedJobs) {
		log.Printf("   Filtered out applied job: %s at %s", title, company)
		return nil
	}
	return &JobListing{C: company, T: title, U: item.Link}
}

// formatListings converts structured listings to display string
func formatListings(listings []JobListing) string {
	if len(listings) == 0 {
		return ""
	}
	var lines []string
	for i, l := range listings {
		lines = append(lines, fmt.Sprintf("%d. %s - %s - %s", i+1, l.C, l.T, l.U))
	}
	header := fmt.Sprintf("Found %d job listings. Return ALL of these to the user:\n", len(listings))
	return header + "\n" + strings.Join(lines, "\n")
}

// matchesAppliedJob checks if a search result matches any applied/do_not_apply job
func matchesAppliedJob(company, title string, appliedJobs []JobInfo) bool {
	companyLower := strings.ToLower(company)
	titleLower := strings.ToLower(title)

	for _, job := range appliedJobs {
		jobCompany := strings.ToLower(job.Account)
		jobTitle := strings.ToLower(job.JobTitle)

		// Match if company contains the job's company (or vice versa) AND title contains job title
		companyMatch := strings.Contains(companyLower, jobCompany) || strings.Contains(jobCompany, companyLower)
		titleMatch := strings.Contains(titleLower, jobTitle) || strings.Contains(jobTitle, titleLower)

		if companyMatch && titleMatch && jobCompany != "" {
			return true
		}
	}
	return false
}

type titlePattern struct {
	delimiter    string
	companyFirst bool   // true = company is parts[0], false = company is parts[last]
	jobTitle     string // "first", "last", or "original"
}

var titlePatterns = []titlePattern{
	{" at ", false, "first"},
	{" @ ", false, "first"},
	{" | ", true, "last"},
	{" – ", true, "last"},
	{" - ", true, "original"},
}

// extractCompanyFromTitle extracts company name from title string
func extractCompanyFromTitle(title string) (string, string) {
	for _, p := range titlePatterns {
		company, jobTitle := tryExtractByDelimiter(title, p)
		if company != "" {
			return company, jobTitle
		}
	}
	return "", title
}

func tryExtractByDelimiter(title string, p titlePattern) (string, string) {
	if !strings.Contains(title, p.delimiter) {
		return "", ""
	}
	parts := strings.Split(title, p.delimiter)
	if len(parts) < 2 {
		return "", ""
	}
	company := strings.TrimSpace(parts[0])
	if !p.companyFirst {
		company = strings.TrimSpace(parts[len(parts)-1])
	}
	if looksLikeJobTitle(company) {
		return "", ""
	}
	jobTitle := getJobTitle(parts, p.jobTitle, title)
	return company, jobTitle
}

func getJobTitle(parts []string, mode, original string) string {
	if mode == "first" {
		return strings.TrimSpace(parts[0])
	}
	if mode == "last" {
		return strings.TrimSpace(parts[len(parts)-1])
	}
	return original
}

// looksLikeJobTitle returns true if the string appears to be a job title or generic page name
func looksLikeJobTitle(s string) bool {
	lower := strings.ToLower(s)
	keywords := []string{
		// Job titles
		"engineer", "developer", "manager", "director", "analyst",
		"architect", "designer", "specialist", "consultant", "lead",
		"senior", "junior", "staff", "principal", "head of",
		"vp ", "vice president", "chief", "officer", "coordinator",
		"administrator", "intern", "associate", "scientist", "researcher",
		// Generic page names (not company names)
		"careers", "jobs", "openings", "opportunities", "hiring",
	}
	for _, kw := range keywords {
		if strings.Contains(lower, kw) {
			return true
		}
	}
	return false
}

// extractCompanyFromURL extracts company name from URL domain as fallback
func extractCompanyFromURL(url string) string {
	// Remove protocol
	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimPrefix(url, "http://")

	// Get domain part
	parts := strings.Split(url, "/")
	if len(parts) == 0 {
		return ""
	}
	domain := parts[0]

	// Remove common prefixes
	domain = strings.TrimPrefix(domain, "www.")
	domain = strings.TrimPrefix(domain, "careers.")
	domain = strings.TrimPrefix(domain, "jobs.")

	// Get base domain (remove TLD)
	domainParts := strings.Split(domain, ".")
	if len(domainParts) == 0 {
		return ""
	}

	company := domainParts[0]

	// Capitalize first letter
	if len(company) > 0 {
		company = strings.ToUpper(string(company[0])) + company[1:]
	}

	return company
}

// checkCache looks up cached job listings and formats them for display
func (t *SerperTools) checkCache(query string) *JobSearchResult {
	if t.cacheTools == nil {
		return nil
	}

	cached := t.cacheTools.LookupCache("job_search", query)
	if cached == nil {
		return nil
	}

	var listings []JobListing
	if err := json.Unmarshal([]byte(cached.Data), &listings); err != nil {
		log.Printf("[ResearchCache] Failed to unmarshal listings: %v", err)
		return nil
	}

	return &JobSearchResult{Results: formatListings(listings), Count: len(listings)}
}

// storeListingsCache stores structured job listings in cache
func (t *SerperTools) storeListingsCache(query string, listings []JobListing) {
	if t.cacheTools == nil {
		return
	}

	data, err := json.Marshal(listings)
	if err != nil {
		log.Printf("Failed to marshal listings for cache: %v", err)
		return
	}

	t.cacheTools.StoreCache("job_search", query, string(data), len(listings))
}

