package utils

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// URL extraction regex - matches URLs in text
var urlRegex = regexp.MustCompile(`https?://[^\s<>"{}|\\^` + "`" + `\[\]]+`)

// ExtractURLs extracts all URLs from the given text
func ExtractURLs(text string) []string {
	return urlRegex.FindAllString(text, -1)
}

// ValidateURLStructure checks if a URL has valid structure:
// - Valid scheme (http or https)
// - Non-empty host
// - Host contains a dot (valid domain)
func ValidateURLStructure(rawURL string) error {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}

	// Check scheme
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("invalid scheme '%s': must be http or https", parsed.Scheme)
	}

	// Check host
	if parsed.Host == "" {
		return fmt.Errorf("missing host")
	}

	// Check that host has a dot (valid domain)
	if !strings.Contains(parsed.Host, ".") {
		return fmt.Errorf("invalid domain '%s': must contain a dot", parsed.Host)
	}

	return nil
}

// ValidateURLReachable checks if a URL is reachable via HTTP HEAD request
// Returns nil if the URL returns a status code < 400
func ValidateURLReachable(ctx context.Context, rawURL string) error {
	// Create a client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Allow up to 10 redirects
			if len(via) >= 10 {
				return fmt.Errorf("too many redirects")
			}
			return nil
		},
	}

	// Create HEAD request
	req, err := http.NewRequestWithContext(ctx, "HEAD", rawURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set a user agent to avoid being blocked
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; PinaColada/1.0)")

	// Execute request
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode >= 400 {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	return nil
}

// URLValidationResult holds the result of URL validation
type URLValidationResult struct {
	URL            string
	StructureValid bool
	StructureError string
	Reachable      bool
	ReachableError string
	StatusCode     int
}

// ValidateURLs validates a list of URLs for both structure and reachability
func ValidateURLs(ctx context.Context, urls []string) []URLValidationResult {
	results := make([]URLValidationResult, len(urls))
	for i, rawURL := range urls {
		results[i] = validateSingleURL(ctx, rawURL)
	}
	return results
}

func validateSingleURL(ctx context.Context, rawURL string) URLValidationResult {
	result := URLValidationResult{URL: rawURL}

	result.StructureValid = true
	if err := ValidateURLStructure(rawURL); err != nil {
		result.StructureValid = false
		result.StructureError = err.Error()
		return result
	}

	result.Reachable = true
	if err := ValidateURLReachable(ctx, rawURL); err != nil {
		result.Reachable = false
		result.ReachableError = err.Error()
	}

	return result
}

// CalculateValidityRate calculates the percentage of valid (reachable) URLs
func CalculateValidityRate(results []URLValidationResult) float64 {
	if len(results) == 0 {
		return 0.0
	}

	valid := 0
	for _, r := range results {
		if r.StructureValid && r.Reachable {
			valid++
		}
	}

	return float64(valid) / float64(len(results))
}
