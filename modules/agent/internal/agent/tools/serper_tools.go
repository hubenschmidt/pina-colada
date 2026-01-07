package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
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
	"alignerr.com",
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
	ATSMode    bool   `json:"ats_mode,omitempty" jsonschema:"Set true to search only ATS platforms (Lever, Greenhouse, Ashby) for direct application links - best for startup jobs"`
	TimeFilter string `json:"time_filter,omitempty" jsonschema:"REQUIRED when user specifies recency. Filter by: 'day' (24h), 'week' (7 days), 'month' (30 days). Map user request: 'last 7 days'/'this week' → 'week', 'last 24h'/'today' → 'day', 'last 2 weeks'/'last 30 days' → 'month'."`
	Location   string `json:"location,omitempty" jsonschema:"Geographic location for search results (e.g., 'United States', 'New York, NY', 'San Francisco, CA')"`
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
	Results         string   `json:"results"`
	Count           int      `json:"count"`
	RelatedSearches []string `json:"related_searches,omitempty"`
}

// JobListing represents a single job for compact cache storage
// Uses short field names to minimize JSON size
type JobListing struct {
	C      string    `json:"c"`           // Company
	T      string    `json:"t"`           // Title
	U      string    `json:"u"`           // URL
	Posted time.Time `json:"p,omitempty"` // Verified posting date (zero if unknown)
}

// --- Serper API Types ---

type serperRequest struct {
	Q        string `json:"q"`
	Tbs      string `json:"tbs,omitempty"`      // Time-based search filter (e.g., qdr:w = past week)
	Location string `json:"location,omitempty"` // Geographic location (e.g., "United States", "New York, NY")
}

type serperResponse struct {
	Organic         []serperOrganicResult `json:"organic"`
	RelatedSearches []relatedSearch       `json:"relatedSearches,omitempty"`
}

type relatedSearch struct {
	Query string `json:"query"`
}

type serperOrganicResult struct {
	Title   string `json:"title"`
	Link    string `json:"link"`
	Snippet string `json:"snippet"`
	Date    string `json:"date,omitempty"` // e.g., "Feb 12, 2017", "Jul 4, 2023"
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

	// Call Serper API (num parameter deprecated by Google Sept 2025)
	reqBody := serperRequest{Q: enhancedQuery, Tbs: tbs, Location: params.Location}
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

	// Extract related searches for query suggestions (free alternative to LLM)
	var relatedSearches []string
	for _, rs := range serperResp.RelatedSearches {
		if rs.Query != "" {
			relatedSearches = append(relatedSearches, rs.Query)
		}
	}
	if len(relatedSearches) > 0 {
		log.Printf("🔍 Found %d related searches from Serper API", len(relatedSearches))
	}

	// Extract structured listings (filtered by existing jobs and seen URLs)
	seenURLs := t.getSeenURLs()
	listings := extractListings(serperResp.Organic, existingJobs, seenURLs)

	// Verify posting dates if time filter is set, or auto-verify in ATS mode
	var maxAge time.Duration
	switch params.TimeFilter {
	case "day":
		maxAge = 24 * time.Hour
	case "week":
		maxAge = 7 * 24 * time.Hour
	case "month":
		maxAge = 30 * 24 * time.Hour
	default:
		// Auto-verify in ATS mode (default to 30 days)
		if params.ATSMode {
			maxAge = 30 * 24 * time.Hour
		}
	}
	if maxAge > 0 && len(listings) > 0 {
		log.Printf("🔍 Verifying posting dates (max age: %v, time_filter: %q)...", maxAge, params.TimeFilter)
		// Strict filter when user explicitly requests a valid time range (exclude unknown dates)
		// Only include unknown dates when using ATSMode default (no explicit time filter)
		strictFilter := params.TimeFilter == "day" || params.TimeFilter == "week" || params.TimeFilter == "month"
		listings = verifyPostingDates(listings, maxAge, strictFilter)
	}

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
			Results:         msg,
			Count:           0,
			RelatedSearches: relatedSearches,
		}, nil
	}

	// Format for display (show dates only when verification was attempted)
	return &JobSearchResult{
		Results:         formatListings(listings, maxAge > 0),
		Count:           listingCount,
		RelatedSearches: relatedSearches,
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
func extractListings(organic []serperOrganicResult, existingJobs []JobInfo, seenURLs map[string]bool) []JobListing {
	var listings []JobListing

	for _, item := range organic {
		listing := extractListing(item, existingJobs, seenURLs)
		if listing != nil {
			listings = append(listings, *listing)
		}
	}

	return listings
}

// parseSerperDate parses date strings from Serper API (e.g., "Feb 12, 2017", "Jul 4, 2023")
func parseSerperDate(dateStr string) time.Time {
	if dateStr == "" {
		return time.Time{}
	}
	// Try common formats returned by Serper
	formats := []string{
		"Jan 2, 2006",  // "Feb 12, 2017"
		"Jan 02, 2006", // "Feb 02, 2017"
		"2 Jan 2006",   // "12 Feb 2017"
		"02 Jan 2006",  // "02 Feb 2017"
	}
	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t
		}
	}
	return time.Time{}
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

	listing := &JobListing{C: company, T: title, U: item.Link}

	// Parse date from Serper API if available (avoids page fetch)
	if item.Date != "" {
		listing.Posted = parseSerperDate(item.Date)
		if !listing.Posted.IsZero() {
			log.Printf("   📅 Using API date for %s: %s", company, item.Date)
		}
	}

	return listing
}

// formatListings converts structured listings to display string
// showDates should be true when date verification was attempted
func formatListings(listings []JobListing, showDates bool) string {
	if len(listings) == 0 {
		return ""
	}
	var lines []string
	for _, l := range listings {
		line := fmt.Sprintf("- %s - %s - [url](%s)", l.C, l.T, l.U)
		if showDates {
			if l.Posted.IsZero() {
				line += " (date unknown)"
			} else {
				line += fmt.Sprintf(" (posted: %s)", l.Posted.Format("Jan 2"))
			}
		}
		lines = append(lines, line)
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
func extractCompanyFromURL(urlStr string) string {
	// Remove protocol
	urlStr = strings.TrimPrefix(urlStr, "https://")
	urlStr = strings.TrimPrefix(urlStr, "http://")

	// Get domain and path parts
	parts := strings.Split(urlStr, "/")
	if len(parts) == 0 {
		return ""
	}
	domain := parts[0]

	// For ATS platforms, extract company from path (e.g., job-boards.greenhouse.io/nyiso/jobs/...)
	if strings.Contains(domain, "greenhouse.io") || strings.Contains(domain, "lever.co") || strings.Contains(domain, "ashbyhq.com") {
		if len(parts) >= 2 && parts[1] != "" && parts[1] != "jobs" {
			company := parts[1]
			if len(company) > 0 {
				company = strings.ToUpper(string(company[0])) + company[1:]
			}
			return company
		}
	}

	// Remove common prefixes
	domain = strings.TrimPrefix(domain, "www.")
	domain = strings.TrimPrefix(domain, "careers.")
	domain = strings.TrimPrefix(domain, "jobs.")
	domain = strings.TrimPrefix(domain, "job-boards.")
	domain = strings.TrimPrefix(domain, "boards.")

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

	reqBody := serperRequest{Q: params.Query}
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

// --- Posting Date Verification ---

// Regex patterns for extracting posting dates from ATS platforms
var (
	// JSON-LD datePosted pattern (Lever, Greenhouse, etc.)
	jsonLDDatePattern = regexp.MustCompile(`"datePosted"\s*:\s*"([^"]+)"`)
	// Relative date patterns
	daysAgoPattern   = regexp.MustCompile(`(?i)posted\s+(\d+)\s+days?\s+ago`)
	weeksAgoPattern  = regexp.MustCompile(`(?i)posted\s+(\d+)\s+weeks?\s+ago`)
	monthsAgoPattern = regexp.MustCompile(`(?i)posted\s+(\d+)\s+months?\s+ago`)
	todayPattern     = regexp.MustCompile(`(?i)posted\s+today`)
	yesterdayPattern = regexp.MustCompile(`(?i)posted\s+yesterday`)
)

// fetchPostingDate fetches a job URL and extracts the posting date
// Returns the date, success boolean, and source description
func fetchPostingDate(url string) (time.Time, bool, string) {
	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return time.Time{}, false, ""
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; JobSearchBot/1.0)")

	resp, err := client.Do(req)
	if err != nil {
		return time.Time{}, false, ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return time.Time{}, false, ""
	}

	// Read body (limit to 500KB to avoid huge pages)
	body, err := io.ReadAll(io.LimitReader(resp.Body, 500*1024))
	if err != nil {
		return time.Time{}, false, ""
	}
	html := string(body)

	return parsePostingDate(html)
}

// parsePostingDate extracts posting date from HTML content
// Returns: date, success, source description
func parsePostingDate(html string) (time.Time, bool, string) {
	now := time.Now()

	// Try JSON-LD datePosted first (most reliable)
	if matches := jsonLDDatePattern.FindStringSubmatch(html); len(matches) > 1 {
		if t, err := time.Parse("2006-01-02", matches[1]); err == nil {
			return t, true, fmt.Sprintf("JSON-LD datePosted: %s", matches[1])
		}
		if t, err := time.Parse(time.RFC3339, matches[1]); err == nil {
			return t, true, fmt.Sprintf("JSON-LD datePosted: %s", matches[1])
		}
	}

	// Try "posted today"
	if todayPattern.MatchString(html) {
		return now.Truncate(24 * time.Hour), true, "text: 'posted today'"
	}

	// Try "posted yesterday"
	if yesterdayPattern.MatchString(html) {
		return now.Add(-24 * time.Hour).Truncate(24 * time.Hour), true, "text: 'posted yesterday'"
	}

	// Try "posted X days ago"
	if matches := daysAgoPattern.FindStringSubmatch(html); len(matches) > 1 {
		days := 0
		fmt.Sscanf(matches[1], "%d", &days)
		return now.Add(-time.Duration(days) * 24 * time.Hour).Truncate(24 * time.Hour), true, fmt.Sprintf("text: 'posted %s days ago'", matches[1])
	}

	// Try "posted X weeks ago"
	if matches := weeksAgoPattern.FindStringSubmatch(html); len(matches) > 1 {
		weeks := 0
		fmt.Sscanf(matches[1], "%d", &weeks)
		return now.Add(-time.Duration(weeks*7) * 24 * time.Hour).Truncate(24 * time.Hour), true, fmt.Sprintf("text: 'posted %s weeks ago'", matches[1])
	}

	// Try "posted X months ago"
	if matches := monthsAgoPattern.FindStringSubmatch(html); len(matches) > 1 {
		months := 0
		fmt.Sscanf(matches[1], "%d", &months)
		return now.Add(-time.Duration(months*30) * 24 * time.Hour).Truncate(24 * time.Hour), true, fmt.Sprintf("text: 'posted %s months ago'", matches[1])
	}

	return time.Time{}, false, ""
}

// verifyPostingDates filters listings by posting date
// maxAge is the maximum age for listings (e.g., 7*24*time.Hour for 1 week)
// strictFilter: if true, exclude listings with unknown dates; if false, include them at the end
// Listings with API dates are checked directly; others are fetched concurrently
func verifyPostingDates(listings []JobListing, maxAge time.Duration, strictFilter bool) []JobListing {
	if len(listings) == 0 {
		return listings
	}

	cutoff := time.Now().Add(-maxAge)
	var verified []JobListing
	var unknown []JobListing
	var needsFetch []int // indices of listings that need page fetch

	// First pass: check listings with API dates, collect those needing fetch
	for i, listing := range listings {
		if !listing.Posted.IsZero() {
			// Already have date from API - apply cutoff directly
			if listing.Posted.Before(cutoff) {
				log.Printf("   📅 Filtered out old posting (%s): %s - %s [API date]", listing.Posted.Format("Jan 2"), listing.C, listing.T)
				continue
			}
			verified = append(verified, listing)
			log.Printf("   📅 Verified fresh posting (%s): %s - %s [API date]", listing.Posted.Format("Jan 2"), listing.C, listing.T)
			continue
		}
		needsFetch = append(needsFetch, i)
	}

	// If no listings need fetching, return early
	if len(needsFetch) == 0 {
		return append(verified, unknown...)
	}

	log.Printf("   📅 %d listings have API dates, %d need page fetch", len(listings)-len(needsFetch), len(needsFetch))

	// Second pass: fetch pages concurrently for listings without API dates
	type result struct {
		index  int
		posted time.Time
		ok     bool
		source string
	}

	results := make(chan result, len(needsFetch))
	sem := make(chan struct{}, 5) // Limit to 5 concurrent requests

	var wg sync.WaitGroup
	for _, idx := range needsFetch {
		wg.Add(1)
		go func(i int, url string) {
			defer wg.Done()
			sem <- struct{}{}        // Acquire
			defer func() { <-sem }() // Release

			posted, ok, source := fetchPostingDate(url)
			results <- result{index: i, posted: posted, ok: ok, source: source}
		}(idx, listings[idx].U)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	postingDates := make(map[int]result)
	for r := range results {
		postingDates[r.index] = r
	}

	// Filter fetched listings
	for _, idx := range needsFetch {
		listing := listings[idx]
		r := postingDates[idx]
		if !r.ok {
			if strictFilter {
				log.Printf("   Filtered out (date unknown, strict filter): %s - %s", listing.C, listing.T)
				continue
			}
			unknown = append(unknown, listing)
			log.Printf("   Date unknown for: %s - %s (no parseable date found)", listing.C, listing.T)
			continue
		}

		listing.Posted = r.posted
		if r.posted.Before(cutoff) {
			log.Printf("   Filtered out old posting (%s): %s - %s [%s]", r.posted.Format("Jan 2"), listing.C, listing.T, r.source)
			continue
		}

		verified = append(verified, listing)
		log.Printf("   Verified fresh posting (%s): %s - %s [%s]", r.posted.Format("Jan 2"), listing.C, listing.T, r.source)
	}

	// Append unknown dates at the end (with warning)
	return append(verified, unknown...)
}

