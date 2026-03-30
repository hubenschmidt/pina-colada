package filtering

import (
	"net/url"
	"strings"
)

// FilterBlacklistedDomains removes results whose URL hostname matches any blacklisted domain.
func FilterBlacklistedDomains(results []JobResult, blacklist []string) []JobResult {
	if len(blacklist) == 0 {
		return results
	}

	blocked := make(map[string]bool, len(blacklist))
	for _, domain := range blacklist {
		blocked[strings.ToLower(strings.TrimSpace(domain))] = true
	}

	var kept []JobResult
	for _, job := range results {
		if isBlacklisted(job.URL, blocked) {
			continue
		}
		kept = append(kept, job)
	}
	return kept
}

func isBlacklisted(rawURL string, blocked map[string]bool) bool {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	host := strings.ToLower(parsed.Hostname())
	if blocked[host] {
		return true
	}
	// Strip leading "www." and check again
	return blocked[strings.TrimPrefix(host, "www.")]
}
