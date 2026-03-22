package filtering

import (
	"bytes"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	readability "codeberg.org/readeck/go-readability/v2"
)

// ValidateURLs checks URLs with HEAD requests to filter broken links (concurrent)
func ValidateURLs(results []JobResult, client *http.Client) []JobResult {
	if len(results) == 0 {
		return results
	}

	const maxConcurrent = 5
	sem := make(chan struct{}, maxConcurrent)
	var wg sync.WaitGroup
	var mu sync.Mutex
	valid := make([]JobResult, 0, len(results))

	for _, r := range results {
		wg.Add(1)
		go func(job JobResult) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			if IsURLValid(client, job.URL) {
				mu.Lock()
				valid = append(valid, job)
				mu.Unlock()
			}
		}(r)
	}
	wg.Wait()

	if len(valid) < len(results) {
		log.Printf("Filtering: URL validation filtered %d/%d broken links", len(results)-len(valid), len(results))
	}

	return valid
}

// IsURLValid checks if a URL responds with a non-error status
func IsURLValid(client *http.Client, url string) bool {
	resp, err := client.Head(url)
	if err != nil {
		log.Printf("Filtering: URL validation failed for %s: %v", url, err)
		return false
	}
	resp.Body.Close()

	if resp.StatusCode >= 400 {
		log.Printf("Filtering: skipping broken link %s (status %d)", url, resp.StatusCode)
		return false
	}
	return true
}

// FetchFullText fetches job pages and extracts readable text via go-readability.
// Results that fail to fetch proceed unchanged (graceful degradation).
func FetchFullText(results []JobResult) []JobResult {
	if len(results) == 0 {
		return results
	}

	start := time.Now()
	var wg sync.WaitGroup
	for i := range results {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			text, err := ExtractReadableText(results[idx].URL)
			if err != nil {
				log.Printf("Filtering: full text extraction failed for %s: %v", results[idx].URL, err)
				return
			}
			results[idx].FullText = text
			log.Printf("Filtering: extracted %d chars from %s", len(text), results[idx].URL)
		}(i)
	}
	wg.Wait()

	fetched := 0
	for _, r := range results {
		if r.FullText != "" {
			fetched++
		}
	}
	log.Printf("Filtering: full text extracted for %d/%d results in %s", fetched, len(results), time.Since(start).Round(time.Millisecond))
	return results
}

// ExtractReadableText fetches a page and extracts readable text
func ExtractReadableText(pageURL string) (string, error) {
	article, err := readability.FromURL(pageURL, 10*time.Second, func(r *http.Request) {
		r.Header.Set("User-Agent", "Mozilla/5.0 (compatible; JobSearchBot/1.0)")
	})
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := article.RenderText(&buf); err != nil {
		return "", err
	}

	text := strings.TrimSpace(buf.String())
	if len(text) > 8000 {
		text = text[:8000]
	}
	return text, nil
}
