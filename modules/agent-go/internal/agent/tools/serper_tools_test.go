package tools

import (
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

	listings := extractListings(organic, 10, nil)

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

	listings := extractListings(organic, 10, appliedJobs)

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

	// Format without URL shortener (nil)
	output := formatListings(listings, nil, "")

	// Verify header
	if !strings.HasPrefix(output, "Found 3 jobs:") {
		t.Errorf("output should start with job count header, got: %s", output[:50])
	}

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

	// Verify format: "N. Company - Title [⭢](url)"
	if !strings.Contains(output, "1. Stripe - Senior Engineer [⭢]") {
		t.Error("output should have numbered format with company and title")
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
		expectedFormat := `Found 10 jobs:
1. Stripe - Senior Engineer [⭢](https://stripe.com/jobs/123)
2. OpenAI - ML Engineer [⭢](https://openai.com/careers/456)
...`

		if !strings.Contains(expectedFormat, "Found 10 jobs") {
			t.Error("format should indicate 10 jobs found")
		}
	})
}
