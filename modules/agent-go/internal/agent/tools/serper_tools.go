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
}

// NewSerperTools creates Serper tools with API key and optional job service for filtering
func NewSerperTools(apiKey string, jobService JobServiceInterface) *SerperTools {
	return &SerperTools{apiKey: apiKey, jobService: jobService}
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
// Filters out jobs already applied to or marked as do_not_apply.
func (t *SerperTools) JobSearchCtx(ctx context.Context, params JobSearchParams) (*JobSearchResult, error) {
	if t.apiKey == "" {
		log.Printf("SERPER_API_KEY not configured")
		return &JobSearchResult{Results: "Job search not configured. SERPER_API_KEY required."}, nil
	}

	// Load applied/do_not_apply jobs for filtering
	var appliedJobs []JobInfo
	if t.jobService != nil {
		jobs, err := t.jobService.GetLeads([]string{"applied", "do_not_apply"}, nil)
		if err != nil {
			log.Printf("Warning: could not load applied jobs for filtering: %v", err)
		} else {
			appliedJobs = jobs
			log.Printf("Loaded %d applied/do_not_apply jobs for filtering", len(appliedJobs))
		}
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

	// Format and filter results
	results := formatSerperResultsWithFilter(serperResp.Organic, 10, appliedJobs)
	log.Printf("job_search found %d results", len(serperResp.Organic))

	if results == "" {
		return &JobSearchResult{
			Results: "All job postings found have already been applied to or marked as 'do not apply'. Try a different search.",
			Count:   0,
		}, nil
	}

	return &JobSearchResult{
		Results: results,
		Count:   len(serperResp.Organic),
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

// formatSerperResultsWithFilter formats organic results and filters against applied jobs
func formatSerperResultsWithFilter(organic []serperOrganicResult, maxResults int, appliedJobs []JobInfo) string {
	if len(organic) == 0 {
		return "No job listings found for this search."
	}

	var lines []string
	count := 0

	for _, item := range organic {
		if count >= maxResults {
			break
		}
		company, title := extractCompanyFromTitle(item.Title)

		// Fallback to URL-based company extraction
		if company == "" {
			company = extractCompanyFromURL(item.Link)
		}

		// Skip if matches an applied/do_not_apply job
		if matchesAppliedJob(company, title, appliedJobs) {
			log.Printf("   Filtered out applied job: %s at %s", title, company)
			continue
		}

		count++
		lines = append(lines, fmt.Sprintf("%d. %s - %s - %s", count, company, title, item.Link))
	}

	if len(lines) == 0 {
		return ""
	}

	header := fmt.Sprintf("Found %d job listings. Return ALL of these to the user:\n", len(lines))
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

// extractCompanyFromTitle extracts company name from title string
func extractCompanyFromTitle(title string) (string, string) {
	// Try " at " pattern (e.g., "Software Engineer at Google")
	if strings.Contains(title, " at ") {
		parts := strings.Split(title, " at ")
		if len(parts) >= 2 {
			company := strings.TrimSpace(parts[len(parts)-1])
			if !looksLikeJobTitle(company) {
				return company, strings.TrimSpace(parts[0])
			}
		}
	}

	// Try " @ " pattern (e.g., "Senior Software Engineer @ Crosby")
	if strings.Contains(title, " @ ") {
		parts := strings.Split(title, " @ ")
		if len(parts) >= 2 {
			company := strings.TrimSpace(parts[len(parts)-1])
			if !looksLikeJobTitle(company) {
				return company, strings.TrimSpace(parts[0])
			}
		}
	}

	// Try " | " pattern (e.g., "AI Agent Security | Senior Software Engineer")
	if strings.Contains(title, " | ") {
		parts := strings.Split(title, " | ")
		if len(parts) >= 2 {
			company := strings.TrimSpace(parts[0])
			if !looksLikeJobTitle(company) {
				return company, strings.TrimSpace(parts[len(parts)-1])
			}
		}
	}

	// Try em-dash " – " pattern (e.g., "Senior AI Engineer – Core AI")
	if strings.Contains(title, " – ") {
		parts := strings.Split(title, " – ")
		if len(parts) >= 2 {
			company := strings.TrimSpace(parts[0])
			if !looksLikeJobTitle(company) {
				return company, strings.TrimSpace(parts[len(parts)-1])
			}
		}
	}

	// Try " - " pattern (e.g., "Google - Software Engineer")
	if strings.Contains(title, " - ") {
		parts := strings.Split(title, " - ")
		if len(parts) >= 2 {
			company := strings.TrimSpace(parts[0])
			if !looksLikeJobTitle(company) {
				return company, title
			}
		}
	}

	return "", title
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

