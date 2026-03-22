package filtering

import (
	"log"
	"strings"
)

// DuplicateSource returns the source of the duplicate ("Job", "Proposal", "Rejected") or empty if not duplicate
func (d *DedupData) DuplicateSource(job JobResult) string {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if source := d.checkURLMatch(job.URL); source != "" {
		return source
	}

	normalizedCompany := NormalizeCompanyName(job.Company)
	normalizedTitle := NormalizeTitleForMatch(job.Title)

	if len(normalizedCompany) < 4 {
		return ""
	}

	if d.matchesExistingJob(normalizedCompany, normalizedTitle) {
		return "Job"
	}

	if d.matchesPendingProposal(normalizedCompany, normalizedTitle) {
		return "Proposal"
	}

	return ""
}

func (d *DedupData) checkURLMatch(url string) string {
	if url == "" {
		return ""
	}
	if source, ok := d.URLSource[url]; ok {
		return source
	}
	return ""
}

func (d *DedupData) matchesExistingJob(company, title string) bool {
	return d.JobIndex[company+"|"+title]
}

func (d *DedupData) matchesPendingProposal(company, title string) bool {
	return d.ProposalIndex[company+"|"+title]
}

// MarkURL adds a URL to the dedup set (thread-safe)
func (d *DedupData) MarkURL(url, source string) {
	if url == "" {
		return
	}
	d.mu.Lock()
	d.URLSource[url] = source
	d.mu.Unlock()
}

// BuildCompanyTitleKey creates normalized key for hash index
func BuildCompanyTitleKey(company, title string) string {
	normalizedCompany := NormalizeCompanyName(company)
	normalizedTitle := NormalizeTitleForMatch(title)
	if normalizedCompany == "" || len(normalizedCompany) < 4 {
		return ""
	}
	return normalizedCompany + "|" + normalizedTitle
}

// NormalizeCompanyName removes common suffixes and normalizes for comparison
func NormalizeCompanyName(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	suffixes := []string{", inc.", ", inc", " inc.", " inc", ", llc", " llc", ", corp.", " corp.", ", corp", " corp", ", ltd.", " ltd.", ", ltd", " ltd", " co.", " co"}
	for _, suffix := range suffixes {
		name = strings.TrimSuffix(name, suffix)
	}
	return strings.TrimSpace(name)
}

// NormalizeTitleForMatch removes seniority modifiers for fuzzy matching
func NormalizeTitleForMatch(title string) string {
	title = strings.ToLower(strings.TrimSpace(title))
	prefixes := []string{"senior ", "sr. ", "sr ", "lead ", "principal ", "staff ", "junior ", "jr. ", "jr ", "associate ", "entry level ", "entry-level "}
	for _, prefix := range prefixes {
		title = strings.TrimPrefix(title, prefix)
	}
	return strings.TrimSpace(title)
}

// FilterDuplicates removes jobs that already exist in the dedup data
func FilterDuplicates(results []JobResult, dedup *DedupData) []JobResult {
	if dedup == nil {
		return results
	}
	filtered := make([]JobResult, 0, len(results))
	for _, job := range results {
		dupSource := dedup.DuplicateSource(job)
		if dupSource != "" {
			log.Printf("Filtering: duplicate [%s]: %s at %s", dupSource, job.Title, job.Company)
		}
		if dupSource == "" {
			dedup.MarkURL(job.URL, "Pending")
			filtered = append(filtered, job)
		}
	}
	return filtered
}
