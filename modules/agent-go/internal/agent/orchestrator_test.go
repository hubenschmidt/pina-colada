package agent_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/pina-colada-co/agent-go/internal/agent"
	"github.com/pina-colada-co/agent-go/internal/config"
	"github.com/pina-colada-co/agent-go/internal/repositories"
	"github.com/pina-colada-co/agent-go/internal/services"
	"github.com/pina-colada-co/agent-go/pkg/db"
)

// loadTestPrompt reads a prompt file from internal/test_prompts directory
func loadTestPrompt(t *testing.T, filename string) string {
	// test_prompts is sibling to agent directory (both under internal/)
	path := filepath.Join("..", "test_prompts", filename)
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to load test prompt %s: %v", filename, err)
	}
	return strings.TrimSpace(string(content))
}

// TestJobSearchReturnsResults tests that job search returns actual job listings
// Run with: go test -v -run TestJobSearchReturnsResults ./internal/agent/
func TestJobSearchReturnsResults(t *testing.T) {
	// Skip if no API key
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		t.Skip("GEMINI_API_KEY not set, skipping integration test")
	}

	serperKey := os.Getenv("SERPER_API_KEY")
	if serperKey == "" {
		t.Skip("SERPER_API_KEY not set, skipping integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// Create config
	cfg := &config.Config{
		GeminiAPIKey: apiKey,
		GeminiModel:  "gemini-2.5-flash",
		SerperAPIKey: serperKey,
	}

	// Create orchestrator with nil services (job search doesn't need CRM for this test)
	// Note: In a real test, you'd mock these or use test fixtures
	var indService *services.IndividualService
	var orgService *services.OrganizationService
	var docService *services.DocumentService

	orchestrator, err := agent.NewOrchestrator(ctx, cfg, indService, orgService, docService)
	if err != nil {
		t.Fatalf("Failed to create orchestrator: %v", err)
	}

	// Simple job search query (doesn't require CRM)
	prompt := "Search for Senior Software Engineer jobs at NYC startups. Return direct links to company career pages."

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

	// Should contain at least one acceptable URL pattern (direct company career pages)
	// Note: Greenhouse/Lever are now excluded - we expect direct company URLs
	acceptablePatterns := []string{
		"careers",
		"jobs.",
		"/jobs",
		"apply",
		"/career",
		"/opportunities",
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

// TestJobSearchWithResumeStub is a placeholder - see TestJobSearchWithResume below
func TestJobSearchWithResumeStub(t *testing.T) {
	t.Skip("Superseded by TestJobSearchWithResume")
}

// TestCRMDocumentLookup tests looking up an individual and their linked documents
// Run with: go test -v -run TestCRMDocumentLookup ./internal/agent/
func TestCRMDocumentLookup(t *testing.T) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		t.Skip("GEMINI_API_KEY not set, skipping integration test")
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("DATABASE_URL not set, skipping integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// Connect to database
	database, err := db.Connect(dbURL, false)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	sqlDB, _ := database.DB()
	defer sqlDB.Close()

	// Initialize storage and repositories
	storageRepo := repositories.GetStorageRepository()
	indRepo := repositories.NewIndividualRepository(database)
	orgRepo := repositories.NewOrganizationRepository(database)
	docRepo := repositories.NewDocumentRepository(database)

	// Initialize services
	indService := services.NewIndividualService(indRepo)
	orgService := services.NewOrganizationService(orgRepo)
	docService := services.NewDocumentService(docRepo, storageRepo)

	// Create config
	cfg := &config.Config{
		GeminiAPIKey: apiKey,
		GeminiModel:  "gemini-2.5-flash",
		SerperAPIKey: os.Getenv("SERPER_API_KEY"),
	}

	// Create orchestrator with real services
	orchestrator, err := agent.NewOrchestrator(ctx, cfg, indService, orgService, docService)
	if err != nil {
		t.Fatalf("Failed to create orchestrator: %v", err)
	}

	// Load test prompt
	prompt := loadTestPrompt(t, "crm_document_lookup_prompt.txt")

	resp, err := orchestrator.Run(ctx, agent.RunRequest{
		SessionID: "test-crm-doc-lookup",
		UserID:    "test-user",
		Message:   prompt,
	})
	if err != nil {
		t.Fatalf("Agent run failed: %v", err)
	}

	t.Logf("Response:\n%s", resp.Response)

	// Validate response
	if resp.Response == "" {
		t.Fatal("Expected non-empty response")
	}

	response := strings.ToLower(resp.Response)

	// Should find William Hubenschmidt
	if !strings.Contains(response, "william") && !strings.Contains(response, "hubenschmidt") {
		t.Error("Expected response to mention William Hubenschmidt")
	}

	// Should mention documents or indicate lookup was done
	hasDocMention := strings.Contains(response, "document") ||
		strings.Contains(response, "resume") ||
		strings.Contains(response, "file") ||
		strings.Contains(response, "linked") ||
		strings.Contains(response, "found")
	if !hasDocMention {
		t.Error("Expected response to mention documents or search results")
	}

	// Should NOT contain service errors
	badPhrases := []string{
		"not configured",
		"service error",
		"unable to connect",
	}
	for _, phrase := range badPhrases {
		if strings.Contains(response, phrase) {
			t.Errorf("Response contains error phrase '%s'", phrase)
		}
	}
}

// TestJobSearchURLsAreValid tests that job search returns valid, reachable URLs
// This test validates URLs similarly to Python agent's test_job_url_validation.py
func TestJobSearchURLsAreValid(t *testing.T) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		t.Skip("GEMINI_API_KEY not set, skipping integration test")
	}

	serperKey := os.Getenv("SERPER_API_KEY")
	if serperKey == "" {
		t.Skip("SERPER_API_KEY not set, skipping integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	defer cancel()

	// Create config
	cfg := &config.Config{
		GeminiAPIKey: apiKey,
		GeminiModel:  "gemini-2.5-flash",
		SerperAPIKey: serperKey,
	}

	// Create orchestrator
	var indService *services.IndividualService
	var orgService *services.OrganizationService
	var docService *services.DocumentService

	orchestrator, err := agent.NewOrchestrator(ctx, cfg, indService, orgService, docService)
	if err != nil {
		t.Fatalf("Failed to create orchestrator: %v", err)
	}

	// Simple job search query (doesn't require CRM)
	prompt := "Search for Senior Software Engineer jobs at NYC startups. Return direct links to company career pages."

	resp, err := orchestrator.Run(ctx, agent.RunRequest{
		SessionID: "test-url-validation",
		UserID:    "test-user",
		Message:   prompt,
	})
	if err != nil {
		t.Fatalf("Agent run failed: %v", err)
	}

	t.Logf("Response:\n%s", resp.Response)

	// Extract URLs from response
	urls := agent.ExtractURLs(resp.Response)
	t.Logf("Found %d URLs in response", len(urls))

	if len(urls) == 0 {
		t.Fatal("Expected at least one URL in response")
	}

	// Validate URL structure for all URLs
	for _, url := range urls {
		if err := agent.ValidateURLStructure(url); err != nil {
			t.Errorf("URL structure invalid: %s - %v", url, err)
		}
	}

	// Test reachability for first 5 URLs (to avoid too many requests)
	maxToCheck := 5
	if len(urls) < maxToCheck {
		maxToCheck = len(urls)
	}

	urlsToCheck := urls[:maxToCheck]
	results := agent.ValidateURLs(ctx, urlsToCheck)

	// Count valid URLs
	validCount := 0
	for _, result := range results {
		if result.StructureValid && result.Reachable {
			validCount++
			t.Logf("✅ Valid URL: %s", result.URL)
		} else {
			reason := result.StructureError
			if reason == "" {
				reason = result.ReachableError
			}
			t.Logf("❌ Invalid URL: %s - %s", result.URL, reason)
		}
	}

	// Calculate validity rate
	validityRate := agent.CalculateValidityRate(results)
	t.Logf("Validity rate: %.1f%% (%d/%d)", validityRate*100, validCount, len(results))

	// Assert at least 50% validity rate (matching Python agent's threshold)
	if validityRate < 0.5 {
		t.Errorf("Validity rate %.1f%% is below 50%% threshold", validityRate*100)
	}
}

// TestJobSearchWithResume tests the full workflow: CRM lookup → read resume → job search
// Run with: GEMINI_API_KEY=... SERPER_API_KEY=... DATABASE_URL=... go test -v -run TestJobSearchWithResume ./internal/agent/
func TestJobSearchWithResume(t *testing.T) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		t.Skip("GEMINI_API_KEY not set, skipping integration test")
	}

	serperKey := os.Getenv("SERPER_API_KEY")
	if serperKey == "" {
		t.Skip("SERPER_API_KEY not set, skipping integration test")
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("DATABASE_URL not set, skipping integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	defer cancel()

	// Connect to database
	database, err := db.Connect(dbURL, false)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	sqlDB, _ := database.DB()
	defer sqlDB.Close()

	// Create storage backend
	storageRepo := repositories.GetStorageRepository()

	// Initialize repositories
	indRepo := repositories.NewIndividualRepository(database)
	orgRepo := repositories.NewOrganizationRepository(database)
	docRepo := repositories.NewDocumentRepository(database)

	// Initialize services
	indService := services.NewIndividualService(indRepo)
	orgService := services.NewOrganizationService(orgRepo)
	docService := services.NewDocumentService(docRepo, storageRepo)

	// Create config
	cfg := &config.Config{
		GeminiAPIKey: apiKey,
		GeminiModel:  "gemini-2.5-flash",
		SerperAPIKey: serperKey,
	}

	// Create orchestrator with real services
	orchestrator, err := agent.NewOrchestrator(ctx, cfg, indService, orgService, docService)
	if err != nil {
		t.Fatalf("Failed to create orchestrator: %v", err)
	}

	// Load test prompt (full job search with resume matching)
	prompt := loadTestPrompt(t, "job_search_test_prompt.txt")
	t.Logf("Test prompt: %s", prompt)

	resp, err := orchestrator.Run(ctx, agent.RunRequest{
		SessionID: "test-job-search-with-resume",
		UserID:    "test-user",
		Message:   prompt,
	})
	if err != nil {
		t.Fatalf("Agent run failed: %v", err)
	}

	t.Logf("Response:\n%s", resp.Response)
	t.Logf("Token usage - turn: in=%d out=%d total=%d", resp.TurnTokens.Input, resp.TurnTokens.Output, resp.TurnTokens.Total)

	// Validate response
	if resp.Response == "" {
		t.Fatal("Expected non-empty response")
	}

	response := strings.ToLower(resp.Response)

	// Should find William Hubenschmidt
	if !strings.Contains(response, "william") && !strings.Contains(response, "hubenschmidt") {
		t.Error("Expected response to mention William Hubenschmidt")
	}

	// Should mention documents were found
	hasDocMention := strings.Contains(response, "document") ||
		strings.Contains(response, "resume") ||
		strings.Contains(response, "linked")
	if !hasDocMention {
		t.Error("Expected response to mention documents")
	}

	// Should contain job search results (URLs)
	urls := agent.ExtractURLs(resp.Response)
	t.Logf("Found %d URLs in response", len(urls))

	if len(urls) == 0 {
		t.Error("Expected at least one job URL in response")
	}

	// Should NOT say it was unable to read resume
	if strings.Contains(response, "unable to read") || strings.Contains(response, "was unable to") {
		t.Error("Agent reported being unable to read resume - document tools may not be working")
	}

	// Check for common error phrases
	badPhrases := []string{
		"not configured",
		"service error",
		"unable to connect",
	}
	for _, phrase := range badPhrases {
		if strings.Contains(response, phrase) {
			t.Errorf("Response contains error phrase '%s'", phrase)
		}
	}

	// Validate URL structure
	validURLCount := 0
	for _, url := range urls {
		if err := agent.ValidateURLStructure(url); err == nil {
			validURLCount++
		}
	}
	t.Logf("Valid URLs: %d/%d", validURLCount, len(urls))
}

// TestURLExtraction tests the URL extraction regex
func TestURLExtraction(t *testing.T) {
	testCases := []struct {
		text     string
		expected []string
	}{
		{
			text:     "Check out https://example.com/jobs and http://test.org",
			expected: []string{"https://example.com/jobs", "http://test.org"},
		},
		{
			text:     "No URLs here",
			expected: []string{},
		},
		{
			text:     "Job at https://boards.greenhouse.io/company/jobs/12345",
			expected: []string{"https://boards.greenhouse.io/company/jobs/12345"},
		},
	}

	for _, tc := range testCases {
		urls := agent.ExtractURLs(tc.text)
		if len(urls) != len(tc.expected) {
			t.Errorf("Expected %d URLs, got %d for text: %s", len(tc.expected), len(urls), tc.text)
		}
		for i, url := range urls {
			if i < len(tc.expected) && url != tc.expected[i] {
				t.Errorf("Expected URL %s, got %s", tc.expected[i], url)
			}
		}
	}
}

// TestURLStructureValidation tests the URL structure validation
func TestURLStructureValidation(t *testing.T) {
	validURLs := []string{
		"https://example.com",
		"http://test.org/path",
		"https://boards.greenhouse.io/company/jobs/123",
	}

	for _, url := range validURLs {
		if err := agent.ValidateURLStructure(url); err != nil {
			t.Errorf("Expected valid URL %s, got error: %v", url, err)
		}
	}

	invalidURLs := []struct {
		url    string
		reason string
	}{
		{"ftp://example.com", "invalid scheme"},
		{"https://", "missing host"},
		{"https://localhost", "no dot in domain"},
	}

	for _, tc := range invalidURLs {
		if err := agent.ValidateURLStructure(tc.url); err == nil {
			t.Errorf("Expected invalid URL %s (%s), but got no error", tc.url, tc.reason)
		}
	}
}
