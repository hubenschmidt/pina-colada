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
	"time"
)

// SerperTools holds Serper API-based tools
type SerperTools struct {
	apiKey string
}

// NewSerperTools creates Serper tools with API key
func NewSerperTools(apiKey string) *SerperTools {
	return &SerperTools{apiKey: apiKey}
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
func (t *SerperTools) JobSearchCtx(ctx context.Context, params JobSearchParams) (*JobSearchResult, error) {
	if t.apiKey == "" {
		log.Printf("SERPER_API_KEY not configured")
		return &JobSearchResult{Results: "Job search not configured. SERPER_API_KEY required."}, nil
	}

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

	// Format results
	results := formatSerperResults(serperResp.Organic, 10)
	log.Printf("job_search found %d results", len(serperResp.Organic))

	return &JobSearchResult{
		Results: results,
		Count:   len(serperResp.Organic),
	}, nil
}

// enhanceJobQuery adds keywords and minimal exclusions to find company career pages
// Only exclude top job boards in query (Google ignores too many -site: operators)
// Post-filtering handles the rest
func enhanceJobQuery(query string) string {
	// Only top 8 job boards - Google ignores too many -site: operators
	// Post-filtering catches the rest
	topJobBoards := []string{
		"linkedin.com",
		"indeed.com",
		"glassdoor.com",
		"ziprecruiter.com",
		"monster.com",
		"dice.com",
		"wellfound.com",
		"builtin.com",
	}

	var exclusions []string
	for _, domain := range topJobBoards {
		exclusions = append(exclusions, fmt.Sprintf("-site:%s", domain))
	}

	return fmt.Sprintf(`%s careers OR jobs %s`, query, strings.Join(exclusions, " "))
}

// formatSerperResults formats organic results into readable job listings
// Post-filters to exclude job board domains that made it through query exclusions
func formatSerperResults(organic []serperOrganicResult, maxResults int) string {
	if len(organic) == 0 {
		return "No job listings found for this search."
	}

	// Filter out job board domains that slipped through query exclusions
	filtered := filterJobBoardResults(organic)
	if len(filtered) == 0 {
		return "No direct company career pages found. Try a different search."
	}

	limit := maxResults
	if len(filtered) < limit {
		limit = len(filtered)
	}

	var lines []string
	lines = append(lines, fmt.Sprintf("Found %d job listings. Return ALL of these to the user:", limit))
	lines = append(lines, "")

	for i, item := range filtered[:limit] {
		company, title := extractCompanyFromTitle(item.Title)
		// Simple format: number. Company - Title - URL
		lines = append(lines, fmt.Sprintf("%d. %s - %s - %s", i+1, company, title, item.Link))
	}

	return strings.Join(lines, "\n")
}

// filterJobBoardResults removes results from job board domains
// NOTE: ATS platforms (Lever, Greenhouse, Ashby) are KEPT - they are direct application links
func filterJobBoardResults(results []serperOrganicResult) []serperOrganicResult {
	// Job boards and aggregators to exclude
	// ATS platforms (lever.co, greenhouse.io, ashbyhq.com) are ALLOWED - they're direct company apps
	excludedDomains := []string{
		// Major job boards
		"linkedin.com", "indeed.com", "glassdoor.com", "ziprecruiter.com",
		"monster.com", "careerbuilder.com", "dice.com", "simplyhired.com",
		"snagajob.com",
		// Startup/tech aggregators (not direct applications)
		"wellfound.com", "angellist.com", "angel.co",
		"builtin.com", "builtinnyc.com", "builtinla.com", "builtinchicago.com",
		"builtinaustin.com", "builtinboston.com", "builtincolorado.com",
		"builtinsf.com", "builtinseattle.com",
		"ycombinator.com", "workatastartup.com",
		"triplebyte.com", "hired.com", "otta.com",
		// Recruiter/staffing aggregators
		"jobright.ai", "technyc.org", "getclera.com", "jobtarget.com",
		"startup.jobs", "levels.fyi", "remoterocketship.com", "standout.work",
		"roberthalf.com", "motionrecruitment.com", "theladders.com",
		"hello.cv", "gem.com", "tallo.com", "talent.com", "careerjet.com",
		"jooble.org", "neuvoo.com", "lensa.com", "jobcase.com",
	}

	var filtered []serperOrganicResult
	for _, result := range results {
		linkLower := strings.ToLower(result.Link)
		excluded := false
		for _, domain := range excludedDomains {
			if strings.Contains(linkLower, domain) {
				log.Printf("   Filtered out job board: %s", result.Link)
				excluded = true
				break
			}
		}
		if !excluded {
			filtered = append(filtered, result)
		}
	}

	log.Printf("Post-filter: %d/%d results kept", len(filtered), len(results))
	return filtered
}

// extractCompanyFromTitle extracts company name from title string
func extractCompanyFromTitle(title string) (string, string) {
	company := "Unknown Company"

	// Try " at " pattern (e.g., "Software Engineer at Google")
	if strings.Contains(title, " at ") {
		parts := strings.Split(title, " at ")
		if len(parts) >= 2 {
			company = strings.TrimSpace(parts[len(parts)-1])
			title = strings.TrimSpace(parts[0])
		}
	} else if strings.Contains(title, " - ") {
		// Try " - " pattern (e.g., "Google - Software Engineer")
		parts := strings.Split(title, " - ")
		if len(parts) >= 2 {
			company = strings.TrimSpace(parts[0])
		}
	}

	return company, title
}

