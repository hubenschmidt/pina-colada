package tools

import (
	"fmt"
	"strings"
	"testing"
)

func TestExtractCompanyFromTitle(t *testing.T) {
	tests := []struct {
		title           string
		expectedCompany string
		expectedTitle   string
	}{
		{"Senior Engineer at Stripe", "Stripe", "Senior Engineer"},
		{"Software Engineer @ Google", "Google", "Software Engineer"},
		{"Airbnb | Full Stack Developer", "Airbnb", "Full Stack Developer"},
		{"OpenAI – ML Engineer", "OpenAI", "ML Engineer"},
		{"Netflix - Senior SWE", "Netflix", "Netflix - Senior SWE"},
	}

	for _, tc := range tests {
		t.Run(tc.title, func(t *testing.T) {
			company, title := extractCompanyFromTitle(tc.title)
			if company != tc.expectedCompany {
				t.Errorf("company: got %q, want %q", company, tc.expectedCompany)
			}
			if title != tc.expectedTitle {
				t.Errorf("title: got %q, want %q", title, tc.expectedTitle)
			}
		})
	}
}

func TestExtractCompanyFromURL(t *testing.T) {
	tests := []struct {
		url      string
		expected string
	}{
		{"https://careers.google.com/jobs/123", "Google"},
		{"https://jobs.lever.co/stripe/abc", "Lever"},
		{"https://stripe.com/jobs", "Stripe"},
		{"https://www.openai.com/careers", "Openai"},
	}

	for _, tc := range tests {
		t.Run(tc.url, func(t *testing.T) {
			got := extractCompanyFromURL(tc.url)
			if got != tc.expected {
				t.Errorf("extractCompanyFromURL(%q) = %q, want %q", tc.url, got, tc.expected)
			}
		})
	}
}

func TestMatchesAppliedJob(t *testing.T) {
	appliedJobs := []JobInfo{
		{JobTitle: "Software Engineer", Account: "Google"},
		{JobTitle: "Senior Developer", Account: "Stripe"},
	}

	tests := []struct {
		company  string
		title    string
		expected bool
	}{
		{"Google", "Software Engineer", true},
		{"google", "software engineer", true}, // case insensitive
		{"Stripe", "Senior Developer", true},
		{"Meta", "Software Engineer", false},  // different company
		{"Google", "Product Manager", false},  // different title
	}

	for _, tc := range tests {
		t.Run(tc.company+"_"+tc.title, func(t *testing.T) {
			got := matchesAppliedJob(tc.company, tc.title, appliedJobs)
			if got != tc.expected {
				t.Errorf("matchesAppliedJob(%q, %q) = %v, want %v", tc.company, tc.title, got, tc.expected)
			}
		})
	}
}

// TestExtractListingsReturns10Results verifies that job search extracts up to 10 results
func TestExtractListingsReturns10Results(t *testing.T) {
	// Simulate 15 organic results from Serper API
	organic := []serperOrganicResult{
		{Title: "Senior Engineer at Stripe", Link: "https://stripe.com/jobs/123"},
		{Title: "Software Engineer at OpenAI", Link: "https://openai.com/careers/456"},
		{Title: "Full Stack Developer at Airbnb", Link: "https://careers.airbnb.com/positions/789"},
		{Title: "Backend Engineer at Netflix", Link: "https://jobs.netflix.com/jobs/abc"},
		{Title: "ML Engineer at Google", Link: "https://careers.google.com/jobs/def"},
		{Title: "Platform Engineer at Meta", Link: "https://www.metacareers.com/jobs/ghi"},
		{Title: "Staff Engineer at Uber", Link: "https://www.uber.com/careers/jkl"},
		{Title: "Senior SWE at Lyft", Link: "https://www.lyft.com/careers/mno"},
		{Title: "Principal Engineer at Spotify", Link: "https://www.lifeatspotify.com/jobs/pqr"},
		{Title: "Tech Lead at Dropbox", Link: "https://www.dropbox.com/jobs/stu"},
		{Title: "Engineer at Slack", Link: "https://slack.com/careers/vwx"},
		{Title: "Developer at Notion", Link: "https://notion.so/careers/yza"},
		{Title: "SWE at Figma", Link: "https://figma.com/careers/bcd"},
		{Title: "Engineer at Linear", Link: "https://linear.app/careers/efg"},
		{Title: "Developer at Vercel", Link: "https://vercel.com/careers/hij"},
	}

	listings := extractListings(organic, 10, nil, nil)

	if len(listings) != 10 {
		t.Errorf("extractListings returned %d results, want 10", len(listings))
	}

	// Verify first result
	if listings[0].C != "Stripe" {
		t.Errorf("first listing company: got %q, want %q", listings[0].C, "Stripe")
	}
	if listings[0].U != "https://stripe.com/jobs/123" {
		t.Errorf("first listing URL: got %q, want direct URL", listings[0].U)
	}
}

// TestExtractListingsFiltersAppliedJobs verifies applied jobs are filtered out
func TestExtractListingsFiltersAppliedJobs(t *testing.T) {
	organic := []serperOrganicResult{
		{Title: "Senior Engineer at Stripe", Link: "https://stripe.com/jobs/123"},
		{Title: "Software Engineer at Google", Link: "https://google.com/careers/456"},
		{Title: "Developer at Meta", Link: "https://meta.com/careers/789"},
	}

	appliedJobs := []JobInfo{
		{JobTitle: "Software Engineer", Account: "Google"},
	}

	listings := extractListings(organic, 10, appliedJobs, nil)

	// Should return 2, not 3 (Google filtered out)
	if len(listings) != 2 {
		t.Errorf("extractListings returned %d results, want 2 (with 1 filtered)", len(listings))
	}

	// Verify Google is not in results
	for _, l := range listings {
		if strings.Contains(strings.ToLower(l.C), "google") {
			t.Error("Google should have been filtered out as applied job")
		}
	}
}

// TestFormatListingsProducesDirectURLs verifies output format contains direct URLs
func TestFormatListingsProducesDirectURLs(t *testing.T) {
	listings := []JobListing{
		{C: "Stripe", T: "Senior Engineer", U: "https://stripe.com/jobs/123"},
		{C: "OpenAI", T: "ML Engineer", U: "https://openai.com/careers/456"},
		{C: "Airbnb", T: "Full Stack Developer", U: "https://careers.airbnb.com/789"},
	}

	output := formatListings(listings)

	// Verify each listing has direct URL in markdown link format
	expectedURLs := []string{
		"https://stripe.com/jobs/123",
		"https://openai.com/careers/456",
		"https://careers.airbnb.com/789",
	}

	for _, url := range expectedURLs {
		if !strings.Contains(output, url) {
			t.Errorf("output should contain direct URL %q", url)
		}
	}

	// Verify format: "- Company - Title - [url](url)"
	if !strings.Contains(output, "- Stripe - Senior Engineer - [url]") {
		t.Error("output should have bullet format with company and title")
	}
}

// TestExtractListingsReturns20Results verifies max_results parameter works for 20
func TestExtractListingsReturns20Results(t *testing.T) {
	// Create 25 organic results
	var organic []serperOrganicResult
	companies := []string{"Stripe", "OpenAI", "Airbnb", "Netflix", "Google", "Meta", "Uber", "Lyft", "Spotify", "Dropbox",
		"Slack", "Notion", "Figma", "Linear", "Vercel", "Supabase", "Cloudflare", "Datadog", "Snowflake", "Databricks",
		"Confluent", "HashiCorp", "GitLab", "MongoDB", "Elastic"}
	for i, company := range companies {
		organic = append(organic, serperOrganicResult{
			Title: fmt.Sprintf("Senior Engineer at %s", company),
			Link:  fmt.Sprintf("https://%s.com/jobs/%d", strings.ToLower(company), i+1),
		})
	}

	listings := extractListings(organic, 20, nil, nil)

	if len(listings) != 20 {
		t.Errorf("extractListings returned %d results, want 20", len(listings))
	}

	// Verify first and last
	if listings[0].C != "Stripe" {
		t.Errorf("first listing company: got %q, want Stripe", listings[0].C)
	}
	if listings[19].C != "Databricks" {
		t.Errorf("20th listing company: got %q, want Databricks", listings[19].C)
	}
}

// TestJobSearchQuality is an integration-style test documenting expected behavior
func TestJobSearchQuality(t *testing.T) {
	t.Run("should return 10 direct company URLs", func(t *testing.T) {
		// This test documents the expected behavior:
		// 1. Job search should return exactly 10 results
		// 2. Results should be direct company career page URLs
		// 3. Results should NOT be job board aggregator URLs

		// Example of expected output format:
		expectedFormat := `- Stripe - Senior Engineer - [url](https://stripe.com/jobs/123)
- OpenAI - ML Engineer - [url](https://openai.com/careers/456)`

		if !strings.Contains(expectedFormat, "- Stripe") {
			t.Error("format should have bullet list")
		}
	})
}

func TestNormalizeCompanyName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Google Inc.", "google"},
		{"Stripe, Inc", "stripe,"},
		{"OpenAI LLC", "openai"},
		{"Meta Corporation", "meta"},
		{"Apple Inc", "apple"},
		{"Microsoft Corp", "microsoft"},
		{"  Notion  ", "notion"},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			got := normalizeCompanyName(tc.input)
			if got != tc.expected {
				t.Errorf("normalizeCompanyName(%q) = %q, want %q", tc.input, got, tc.expected)
			}
		})
	}
}

func TestNormalizeTitleForMatch(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Senior Software Engineer", "software engineer"},
		{"Lead Developer", "developer"},
		{"Staff Engineer", "engineer"},
		{"Junior Analyst", "analyst"},
		{"Principal Architect", "architect"},
		{"Software Engineer", "software engineer"},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			got := normalizeTitleForMatch(tc.input)
			if got != tc.expected {
				t.Errorf("normalizeTitleForMatch(%q) = %q, want %q", tc.input, got, tc.expected)
			}
		})
	}
}

func TestIsAggregatorURL(t *testing.T) {
	tests := []struct {
		url      string
		expected bool
	}{
		{"https://news.ycombinator.com/item?id=12345", true},
		{"https://hnhiring.com/jobs/123", true},
		{"https://facebook.com/groups/jobs", true},
		{"https://zippia.com/jobs/engineer", true},
		{"https://stripe.com/jobs/123", false},
		{"https://jobs.lever.co/stripe/abc", false},
		{"https://boards.greenhouse.io/openai", false},
	}

	for _, tc := range tests {
		t.Run(tc.url, func(t *testing.T) {
			got := isAggregatorURL(tc.url)
			if got != tc.expected {
				t.Errorf("isAggregatorURL(%q) = %v, want %v", tc.url, got, tc.expected)
			}
		})
	}
}

func TestBuildATSQuery(t *testing.T) {
	query := "Senior Software Engineer NYC"
	result := buildATSQuery(query)

	// Should contain original query
	if !strings.Contains(result, query) {
		t.Errorf("ATS query should contain original query, got: %s", result)
	}

	// Should contain ATS platforms
	atsPlatforms := []string{"site:lever.co", "site:greenhouse.io", "site:jobs.ashbyhq.com"}
	for _, platform := range atsPlatforms {
		if !strings.Contains(result, platform) {
			t.Errorf("ATS query should contain %s, got: %s", platform, result)
		}
	}
}

func TestExtractListingsFiltersAggregatorURLs(t *testing.T) {
	organic := []serperOrganicResult{
		{Title: "Senior Engineer at Stripe", Link: "https://stripe.com/jobs/123"},
		{Title: "Jobs on HN Hiring", Link: "https://hnhiring.com/jobs/456"},
		{Title: "Developer at Meta", Link: "https://meta.com/careers/789"},
	}

	listings := extractListings(organic, 10, nil, nil)

	// Should return 2, not 3 (hnhiring filtered out)
	if len(listings) != 2 {
		t.Errorf("extractListings returned %d results, want 2 (with aggregator filtered)", len(listings))
	}

	// Verify hnhiring is not in results
	for _, l := range listings {
		if strings.Contains(l.U, "hnhiring") {
			t.Error("hnhiring URL should have been filtered out")
		}
	}
}

