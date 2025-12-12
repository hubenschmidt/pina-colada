package agent_test

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/pina-colada-co/agent-go/internal/agent"
	"github.com/pina-colada-co/agent-go/internal/config"
	"github.com/pina-colada-co/agent-go/internal/services"
)

// TestJobSearchReturnsResults tests that job search returns actual job listings
// Run with: go test -v -run TestJobSearchReturnsResults ./internal/agent/
func TestJobSearchReturnsResults(t *testing.T) {
	// Skip if no API key
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		t.Skip("GEMINI_API_KEY not set, skipping integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// Create config
	cfg := &config.Config{
		GeminiAPIKey: apiKey,
		GeminiModel:  "gemini-2.5-flash",
	}

	// Create orchestrator with nil services (job search doesn't need CRM for this test)
	// Note: In a real test, you'd mock these or use test fixtures
	var indService *services.IndividualService
	var orgService *services.OrganizationService

	orchestrator, err := agent.NewOrchestrator(ctx, cfg, indService, orgService)
	if err != nil {
		t.Fatalf("Failed to create orchestrator: %v", err)
	}

	// Test prompt
	prompt := `Search for Senior Software Engineer jobs at NYC startups.
I'm looking for roles matching these skills: TypeScript, Python, React, Node.js, AI/LLM engineering.
Find 5 jobs and return direct links to the job postings (Greenhouse, Lever, or company career pages are fine).
Exclude LinkedIn, Indeed, and Glassdoor links.`

	resp, err := orchestrator.Run(ctx, agent.RunRequest{
		SessionID: "test-job-search-1",
		UserID:    "test-user",
		Message:   prompt,
	})
	if err != nil {
		t.Fatalf("Agent run failed: %v", err)
	}

	// Validate response
	if resp.Response == "" {
		t.Fatal("Expected non-empty response")
	}

	t.Logf("Response:\n%s", resp.Response)

	// Check that response contains job-related content
	response := strings.ToLower(resp.Response)

	// Should contain URLs
	hasURL := strings.Contains(response, "http://") || strings.Contains(response, "https://")
	if !hasURL {
		t.Error("Expected response to contain URLs")
	}

	// Should NOT contain apologies about not being able to search
	badPhrases := []string{
		"i can't",
		"i cannot",
		"i'm unable",
		"not able to",
		"blocking us",
		"constraint",
	}
	for _, phrase := range badPhrases {
		if strings.Contains(response, phrase) {
			t.Errorf("Response contains refusal phrase '%s' - agent should return results, not excuses", phrase)
		}
	}

	// Should contain job-related terms
	jobTerms := []string{"engineer", "software", "senior", "role", "position", "job"}
	hasJobTerm := false
	for _, term := range jobTerms {
		if strings.Contains(response, term) {
			hasJobTerm = true
			break
		}
	}
	if !hasJobTerm {
		t.Error("Expected response to contain job-related terms")
	}

	// Should contain at least one acceptable URL pattern
	acceptablePatterns := []string{
		"greenhouse.io",
		"lever.co",
		"careers",
		"jobs.",
		"/jobs",
		"apply",
	}
	hasAcceptableURL := false
	for _, pattern := range acceptablePatterns {
		if strings.Contains(response, pattern) {
			hasAcceptableURL = true
			break
		}
	}
	if !hasAcceptableURL {
		t.Log("Warning: Response may not contain direct career page URLs")
	}
}

// TestJobSearchWithResume tests job search that references CRM data
func TestJobSearchWithResume(t *testing.T) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		t.Skip("GEMINI_API_KEY not set, skipping integration test")
	}

	// This test requires database connection - skip for now
	t.Skip("Requires database connection - implement with test fixtures")
}
