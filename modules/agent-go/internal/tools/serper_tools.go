package tools

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/functiontool"
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
	Q string `json:"q"`
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

// JobSearch searches for jobs using Serper API with domain exclusions
func (t *SerperTools) JobSearch(ctx tool.Context, params JobSearchParams) (*JobSearchResult, error) {
	if t.apiKey == "" {
		log.Printf("SERPER_API_KEY not configured")
		return &JobSearchResult{Results: "Job search not configured. SERPER_API_KEY required."}, nil
	}

	// Enhance query with exclusions
	enhancedQuery := enhanceJobQuery(params.Query)
	log.Printf("job_search query: %s", enhancedQuery)

	// Call Serper API
	reqBody := serperRequest{Q: enhancedQuery}
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

// enhanceJobQuery adds keywords and exclusions to find direct company career pages
// Excludes job boards (LinkedIn, Indeed, etc.) but allows ATS platforms (Greenhouse, Lever, Ashby)
func enhanceJobQuery(query string) string {
	// Job boards to exclude (recruiter spam, aggregators)
	excludedDomains := []string{
		"linkedin.com",
		"indeed.com",
		"glassdoor.com",
		"ziprecruiter.com",
		"monster.com",
		"careerbuilder.com",
		"dice.com",
		"simplyhired.com",
		"snagajob.com",
		"wellfound.com", // formerly AngelList Talent
		"builtin.com",
		"otta.com",
	}

	// Build exclusion string
	var exclusions []string
	for _, domain := range excludedDomains {
		exclusions = append(exclusions, fmt.Sprintf("-site:%s", domain))
	}

	return fmt.Sprintf(`%s "careers" OR "jobs" %s`, query, strings.Join(exclusions, " "))
}

// formatSerperResults formats organic results into readable job listings
func formatSerperResults(organic []serperOrganicResult, maxResults int) string {
	if len(organic) == 0 {
		return "No job listings found for this search."
	}

	limit := maxResults
	if len(organic) < limit {
		limit = len(organic)
	}

	var lines []string
	for i, item := range organic[:limit] {
		company, title := extractCompanyFromTitle(item.Title)
		lines = append(lines, fmt.Sprintf("%d. %s - %s", i+1, company, title))
		if item.Link != "" {
			lines = append(lines, fmt.Sprintf("   URL: %s", item.Link))
		}
		if item.Snippet != "" {
			snippet := item.Snippet
			if len(snippet) > 150 {
				snippet = snippet[:150] + "..."
			}
			lines = append(lines, fmt.Sprintf("   %s", snippet))
		}
		lines = append(lines, "")
	}

	return strings.Join(lines, "\n")
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

// BuildTools returns ADK tool.Tool instances for Serper operations
func (t *SerperTools) BuildTools() ([]tool.Tool, error) {
	jobSearchTool, err := functiontool.New(
		functiontool.Config{
			Name:        "job_search",
			Description: "Search for job listings. Returns direct company career pages and ATS-hosted pages (Greenhouse, Lever, Ashby). Excludes job boards like LinkedIn, Indeed, Glassdoor. Use this for any job search request.",
		},
		t.JobSearch,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create job_search tool: %w", err)
	}

	return []tool.Tool{jobSearchTool}, nil
}
