package filtering

import (
	"net/url"
	"strings"
)

var internationalTLDs = map[string]bool{
	".co.uk":  true,
	".co.in":  true,
	".co.nz":  true,
	".co.jp":  true,
	".com.au": true,
	".com.br": true,
	".com.sg": true,
	".de":     true,
	".fr":     true,
	".ca":     true,
	".ie":     true,
	".nl":     true,
	".sg":     true,
	".in":     true,
	".uk":     true,
	".au":     true,
	".nz":     true,
	".jp":     true,
	".br":     true,
	".mx":     true,
	".za":     true,
	".se":     true,
	".ch":     true,
	".at":     true,
	".be":     true,
	".dk":     true,
	".fi":     true,
	".no":     true,
	".it":     true,
	".es":     true,
	".pt":     true,
	".pl":     true,
}

var internationalSubdomains = map[string]bool{
	"uk":  true,
	"de":  true,
	"fr":  true,
	"in":  true,
	"ca":  true,
	"au":  true,
	"nz":  true,
	"jp":  true,
	"br":  true,
	"mx":  true,
	"sg":  true,
	"ie":  true,
	"nl":  true,
	"za":  true,
	"se":  true,
	"ch":  true,
	"at":  true,
	"be":  true,
	"it":  true,
	"es":  true,
	"pt":  true,
	"pl":  true,
}

var internationalPathPrefixes = map[string]bool{
	"en-gb": true,
	"en-au": true,
	"en-in": true,
	"en-ca": true,
	"en-nz": true,
	"de":    true,
	"fr":    true,
	"uk":    true,
	"in":    true,
	"au":    true,
	"ca":    true,
	"nz":    true,
	"jp":    true,
	"br":    true,
}

// FilterInternationalResults drops results that appear non-US based on URL heuristics.
func FilterInternationalResults(results []JobResult) []JobResult {
	if len(results) == 0 {
		return results
	}

	kept := make([]JobResult, 0, len(results))
	for _, job := range results {
		if isInternationalURL(job.URL) {
			continue
		}
		kept = append(kept, job)
	}
	return kept
}

func isInternationalURL(rawURL string) bool {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return false
	}

	host := strings.ToLower(parsed.Hostname())

	if hasInternationalTLD(host) {
		return true
	}

	if hasInternationalSubdomain(host) {
		return true
	}

	return hasInternationalPathPrefix(parsed.Path)
}

func hasInternationalTLD(host string) bool {
	for tld := range internationalTLDs {
		if strings.HasSuffix(host, tld) {
			return true
		}
	}
	return false
}

func hasInternationalSubdomain(host string) bool {
	parts := strings.Split(host, ".")
	if len(parts) < 3 {
		return false
	}
	return internationalSubdomains[parts[0]]
}

func hasInternationalPathPrefix(path string) bool {
	trimmed := strings.TrimPrefix(path, "/")
	seg, _, _ := strings.Cut(trimmed, "/")
	return internationalPathPrefixes[strings.ToLower(seg)]
}
