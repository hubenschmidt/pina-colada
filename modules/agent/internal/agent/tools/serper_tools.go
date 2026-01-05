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

	// Session-scoped seen URLs to avoid duplicate results across turns
	currentSessionID string
	seenJobURLs      map[string]map[string]bool // sessionID -> URL -> true
	seenMu           sync.RWMutex
}

const appliedJobsCacheTTL = 30 * time.Second
const maxConcurrentSearches = 3
const maxSearchesPerTurn = 3 // Allow multiple searches for different job titles

// Company name suffixes to remove for normalization
var companySuffixes = []string{
	" inc.", " inc", " llc", " corp.", " corp", " ltd.", " ltd",
	" co.", " co", " company", " incorporated", " corporation",
	" limited", " holdings", " group", " technologies", " technology",
}

// Seniority modifiers to remove for fuzzy title matching
var seniorityModifiers = []string{
	"senior ", "sr ", "sr. ", "lead ", "staff ", "principal ",
	"junior ", "jr ", "jr. ", "associate ", "entry level ", "mid-level ",
}

// Aggregator URL patterns to filter out
var aggregatorPatterns = []string{
	"hnhiring", "news.ycombinator.com", "facebook.com/groups",
	"theladders.com", "jobtoday.com", "zippia.com", "whatjobs.com",
	"appinventiv.com", "diffblog", "dynamodb", "instagram.com",
	"asian-jobs", "womenforhire", "snagajob", "getclera.com/job",
}

// normalizeCompanyName removes common suffixes and normalizes for matching
func normalizeCompanyName(name string) string {
	result := strings.ToLower(strings.TrimSpace(name))
	for _, suffix := range companySuffixes {
		result = strings.TrimSuffix(result, suffix)
	}
	return strings.TrimSpace(result)
}

// normalizeTitleForMatch removes seniority modifiers for fuzzy matching
func normalizeTitleForMatch(title string) string {
	result := strings.ToLower(strings.TrimSpace(title))
	for _, modifier := range seniorityModifiers {
		result = strings.TrimPrefix(result, modifier)
	}
	return strings.TrimSpace(result)
}

// isAggregatorURL checks if URL matches known aggregator patterns
func isAggregatorURL(url string) bool {
	urlLower := strings.ToLower(url)
	for _, pattern := range aggregatorPatterns {
		if strings.Contains(urlLower, pattern) {
			return true
		}
	}
	return false
}

// NewSerperTools creates Serper tools with API key and optional job service for filtering
func NewSerperTools(apiKey string, jobService JobServiceInterface) *SerperTools {
	return &SerperTools{
		apiKey:      apiKey,
		jobService:  jobService,
		searchSem:   make(chan struct{}, maxConcurrentSearches),
		seenJobURLs: make(map[string]map[string]bool),
	}
}

// SetSession sets the current session ID for tracking seen URLs across turns
func (t *SerperTools) SetSession(sessionID string) {
	t.seenMu.Lock()
	defer t.seenMu.Unlock()
	t.currentSessionID = sessionID
	if t.seenJobURLs[sessionID] == nil {
		t.seenJobURLs[sessionID] = make(map[string]bool)
	}
}

// getSeenURLs returns a copy of seen URLs for the current session (safe for concurrent use)
func (t *SerperTools) getSeenURLs() map[string]bool {
	t.seenMu.RLock()
	defer t.seenMu.RUnlock()
	original := t.seenJobURLs[t.currentSessionID]
	if original == nil {
		return nil
	}
	// Return a copy to avoid concurrent map access
	copied := make(map[string]bool, len(original))
	for k, v := range original {
		copied[k] = v
	}
	return copied
}

// markURLsAsSeen adds URLs to the seen set for the current session
func (t *SerperTools) markURLsAsSeen(urls []string) {
	t.seenMu.Lock()
	defer t.seenMu.Unlock()
	if t.seenJobURLs[t.currentSessionID] == nil {
		t.seenJobURLs[t.currentSessionID] = make(map[string]bool)
	}
	for _, url := range urls {
		t.seenJobURLs[t.currentSessionID][url] = true
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

// loadExistingJobs loads ALL jobs from database for deduplication filtering
func (t *SerperTools) loadExistingJobs() []JobInfo {
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

	// Get ALL jobs regardless of status for deduplication
	jobs, err := t.jobService.GetLeads(nil, nil)
	if err != nil {
		log.Printf("Warning: could not load existing jobs for filtering: %v", err)
		return nil
	}

	t.appliedJobsCache = jobs
	t.appliedJobsCacheTime = time.Now()
	log.Printf("Loaded %d existing jobs for filtering", len(jobs))
	return jobs
}

// --- Tool Parameter Structs ---

// JobSearchParams defines parameters for job search
type JobSearchParams struct {
	Query      string `json:"query" jsonschema:"Job search query (e.g., 'Senior Software Engineer NYC startups')"`
	MaxResults int    `json:"max_results,omitempty" jsonschema:"Number of results to return (default 10, max 20)"`
	ATSMode    bool   `json:"ats_mode,omitempty" jsonschema:"Set true to search only ATS platforms (Lever, Greenhouse, Ashby) for direct application links - best for startup jobs"`
	TimeFilter string `json:"time_filter,omitempty" jsonschema:"Filter by recency: 'day' (24h), 'week' (7 days), 'month' (30 days). Use based on user's request (e.g., 'last 14 days' → 'month', 'recent' → 'week'). Default: no filter."`
}

// WebSearchParams defines parameters for general web search
type WebSearchParams struct {
	Query string `json:"query" jsonschema:"Search query (e.g., 'John Smith LinkedIn')"`
}

// WebSearchResult is the result of a general web search
type WebSearchResult struct {
	Results string `json:"results"`
	Count   int    `json:"count"`
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
	Tbs string `json:"tbs,omitempty"` // Time-based search filter (e.g., qdr:w = past week)
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

	// Load existing jobs from DB for deduplication
	existingJobs := t.loadExistingJobs()

	// Enhance query - use ATS mode for direct application links, otherwise normal mode
	enhancedQuery := enhanceJobQuery(params.Query)
	if params.ATSMode {
		enhancedQuery = buildATSQuery(params.Query)
	}
	log.Printf("job_search query (ats=%v): %s", params.ATSMode, enhancedQuery)

	// Determine max results (default 10, max 20 - Serper limit)
	maxResults := params.MaxResults
	if maxResults <= 0 {
		maxResults = 10
	}
	if maxResults > 20 {
		maxResults = 20
	}

	// Map time filter to Serper tbs parameter
	var tbs string
	switch params.TimeFilter {
	case "day":
		tbs = "qdr:d"
	case "week":
		tbs = "qdr:w"
	case "month":
		tbs = "qdr:m"
	}

	// Call Serper API - always request 20 (max allowed)
	reqBody := serperRequest{Q: enhancedQuery, Num: 20, Tbs: tbs}
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

	// Extract structured listings (filtered by existing jobs and seen URLs)
	seenURLs := t.getSeenURLs()
	listings := extractListings(serperResp.Organic, maxResults, existingJobs, seenURLs)
	listingCount := len(listings)

	// Mark returned URLs as seen for this session
	var urls []string
	for _, l := range listings {
		urls = append(urls, l.U)
	}
	t.markURLsAsSeen(urls)

	// Update total results count
	t.callCountMu.Lock()
	t.totalResults += listingCount
	totalNow := t.totalResults
	t.callCountMu.Unlock()

	log.Printf("🔍 job_search [%d/%d]: COMPLETE - found %d new results, %d seen URLs filtered (running total: %d)",
		myCallNum, maxSearchesPerTurn, listingCount, len(seenURLs), totalNow)

	if listingCount == 0 {
		msg := "No new job postings found. All results were either already shown, applied to, or marked as 'do not apply'. Try a different search query."
		return &JobSearchResult{
			Results: msg,
			Count:   0,
		}, nil
	}

	// Format for display
	return &JobSearchResult{
		Results: formatListings(listings),
		Count:   listingCount,
	}, nil
}

// buildATSQuery creates an ATS-biased query for direct company job pages
// Use this for startup/tech company searches to get direct application links
func buildATSQuery(baseQuery string) string {
	// ATS platforms where companies post directly
	ats := "(site:lever.co OR site:greenhouse.io OR site:jobs.ashbyhq.com OR site:boards.greenhouse.io)"
	return baseQuery + " " + ats
}

// enhanceJobQuery adds career focus and minimal exclusions
func enhanceJobQuery(query string) string {
	// Add "careers" to bias toward company pages, exclude job boards and aggregators
	exclusions := []string{
		// Major job boards
		"-site:linkedin.com", "-site:indeed.com", "-site:glassdoor.com",
		"-site:ziprecruiter.com", "-site:dice.com", "-site:monster.com",
		"-site:simplyhired.com", "-site:careerbuilder.com",
		// Tech job boards
		"-site:builtinnyc.com", "-site:builtin.com", "-site:jobright.ai",
		"-site:wellfound.com", "-site:roberthalf.com", "-site:hired.com",
		"-site:angel.co", "-site:stackoverflow.com/jobs",
		// Remote job boards
		"-site:remoteok.com", "-site:remoterocketship.com", "-site:weworkremotely.com",
		"-site:remote.co", "-site:flexjobs.com", "-site:dailyremote.com",
		// AI/ML specific job boards
		"-site:aijobs.com", "-site:aijobs.net", "-site:mljobs.com",
		// Aggregators
		"-site:jobleads.com", "-site:jobgether.com", "-site:jooble.org",
		"-site:neuvoo.com", "-site:talent.com", "-site:getwork.com",
		"-site:codingjobboard.com", "-site:snagajob.com", "-site:jobget.com",
		"-site:jobtarget.com", "-site:harnham.com", "-site:glocomms.com",
		"-site:hiringagents.com", "-site:funded.club",
	}
	return query + " careers " + strings.Join(exclusions, " ")
}

// extractListings extracts structured job listings from organic results
func extractListings(organic []serperOrganicResult, maxResults int, existingJobs []JobInfo, seenURLs map[string]bool) []JobListing {
	var listings []JobListing

	for _, item := range organic {
		if len(listings) >= maxResults {
			return listings
		}
		listing := extractListing(item, existingJobs, seenURLs)
		if listing != nil {
			listings = append(listings, *listing)
		}
	}

	return listings
}

func extractListing(item serperOrganicResult, existingJobs []JobInfo, seenURLs map[string]bool) *JobListing {
	// Filter out already-seen URLs first
	if seenURLs != nil && seenURLs[item.Link] {
		log.Printf("   Filtered out seen URL: %s", item.Link)
		return nil
	}

	// Filter out aggregator URLs
	if isAggregatorURL(item.Link) {
		log.Printf("   Filtered out aggregator URL: %s", item.Link)
		return nil
	}

	company, title := extractCompanyFromTitle(item.Title)
	if company == "" {
		company = extractCompanyFromURL(item.Link)
	}
	if matchesAppliedJob(company, title, existingJobs) {
		log.Printf("   Filtered out existing job: %s at %s", title, company)
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
	for _, l := range listings {
		lines = append(lines, fmt.Sprintf("- %s - %s - [url](%s)", l.C, l.T, l.U))
	}
	return strings.Join(lines, "\n")
}

// matchesAppliedJob checks if a search result matches any existing job in database
// Uses normalized company names and fuzzy title matching (ignoring seniority modifiers)
func matchesAppliedJob(company, title string, existingJobs []JobInfo) bool {
	searchCompany := normalizeCompanyName(company)
	searchTitle := normalizeTitleForMatch(title)

	// Require minimum 4 characters to avoid false matches
	if len(searchCompany) < 4 {
		return false
	}

	for _, job := range existingJobs {
		jobCompany := normalizeCompanyName(job.Account)
		jobTitle := normalizeTitleForMatch(job.JobTitle)

		// Skip empty company names
		if jobCompany == "" || len(jobCompany) < 4 {
			continue
		}

		// Company match: exact or substring match
		companyMatch := searchCompany == jobCompany ||
			strings.Contains(searchCompany, jobCompany) ||
			strings.Contains(jobCompany, searchCompany)

		// Title match: normalized title comparison
		titleMatch := searchTitle == jobTitle ||
			strings.Contains(searchTitle, jobTitle) ||
			strings.Contains(jobTitle, searchTitle)

		if companyMatch && titleMatch {
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

// WebSearchCtx performs a general web search without job-specific filtering.
// Use for background research on people, companies, or general information.
func (t *SerperTools) WebSearchCtx(ctx context.Context, params WebSearchParams) (*WebSearchResult, error) {
	if t.apiKey == "" {
		return &WebSearchResult{Results: "Web search not configured. SERPER_API_KEY required."}, nil
	}

	log.Printf("🔍 web_search: query=%s", params.Query)

	reqBody := serperRequest{Q: params.Query, Num: 10}
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return &WebSearchResult{Results: fmt.Sprintf("Error encoding request: %v", err)}, nil
	}

	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequest("POST", "https://google.serper.dev/search", bytes.NewBuffer(jsonBody))
	if err != nil {
		return &WebSearchResult{Results: fmt.Sprintf("Error creating request: %v", err)}, nil
	}

	req.Header.Set("X-API-KEY", t.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Serper API error: %v", err)
		return &WebSearchResult{Results: fmt.Sprintf("Search failed: %v", err)}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("Serper API error: %d - %s", resp.StatusCode, string(body))
		return &WebSearchResult{Results: fmt.Sprintf("Search failed: HTTP %d", resp.StatusCode)}, nil
	}

	var serperResp serperResponse
	if err := json.NewDecoder(resp.Body).Decode(&serperResp); err != nil {
		return &WebSearchResult{Results: fmt.Sprintf("Error parsing response: %v", err)}, nil
	}

	return &WebSearchResult{
		Results: formatWebResults(serperResp.Organic),
		Count:   len(serperResp.Organic),
	}, nil
}

// formatWebResults formats general web search results for display
func formatWebResults(organic []serperOrganicResult) string {
	if len(organic) == 0 {
		return "No results found."
	}

	var lines []string
	for i, item := range organic {
		lines = append(lines, fmt.Sprintf("%d. **%s**\n   %s\n   %s", i+1, item.Title, item.Snippet, item.Link))
	}
	return strings.Join(lines, "\n\n")
}

