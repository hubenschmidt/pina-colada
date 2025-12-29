package services

import (
	"context"
	"encoding/json"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"agent/internal/models"
	"agent/internal/repositories"
)

const summaryMaxTokens int64 = 300

type DocumentSummarizer struct {
	client  anthropic.Client
	docRepo *repositories.DocumentRepository
	enabled bool
}

func NewDocumentSummarizer(apiKey string, docRepo *repositories.DocumentRepository) *DocumentSummarizer {
	if apiKey == "" {
		return &DocumentSummarizer{docRepo: docRepo, enabled: false}
	}
	client := anthropic.NewClient(option.WithAPIKey(apiKey))
	return &DocumentSummarizer{client: client, docRepo: docRepo, enabled: true}
}

// ProcessDocument generates summary and compressed content for a document
func (s *DocumentSummarizer) ProcessDocument(ctx context.Context, documentID int64, content string, filename string) error {
	if !s.enabled {
		log.Printf("DocumentSummarizer: not enabled, skipping")
		return nil
	}

	compressed := compressContent(content)
	summary, err := s.generateSummary(ctx, content, filename)
	if err != nil {
		log.Printf("DocumentSummarizer: failed to generate summary for doc %d: %v", documentID, err)
		// Still save compressed even if summary fails
		return s.docRepo.UpdateDocumentSummary(documentID, nil, &compressed)
	}

	summaryJSON, err := json.Marshal(summary)
	if err != nil {
		log.Printf("DocumentSummarizer: failed to marshal summary: %v", err)
		return s.docRepo.UpdateDocumentSummary(documentID, nil, &compressed)
	}

	return s.docRepo.UpdateDocumentSummary(documentID, summaryJSON, &compressed)
}

func (s *DocumentSummarizer) generateSummary(ctx context.Context, content string, filename string) (*models.DocumentSummary, error) {
	// Truncate content for summarization (first 8000 chars should be enough)
	truncated := content
	if len(content) > 8000 {
		truncated = content[:8000]
	}

	prompt := buildSummaryPrompt(filename, truncated)

	resp, err := s.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     "claude-haiku-4-5-20251001",
		MaxTokens: summaryMaxTokens,
		System: []anthropic.TextBlockParam{
			{Text: "You are a document summarizer. Generate concise summaries."},
		},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
		},
	})
	if err != nil {
		return nil, err
	}

	if len(resp.Content) == 0 {
		return nil, nil
	}

	summaryText := resp.Content[0].Text
	keywords := extractKeywords(summaryText)

	return &models.DocumentSummary{
		Text:        summaryText,
		Keywords:    keywords,
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
	}, nil
}

func buildSummaryPrompt(filename string, content string) string {
	return "Summarize this document in 2-3 sentences. Focus on the key information.\n\n" +
		"Filename: " + filename + "\n\n" +
		"Content:\n" + content
}

// compressContent removes extra whitespace and normalizes the content
func compressContent(content string) string {
	// Replace multiple spaces with single space
	spaceRegex := regexp.MustCompile(`[ \t]+`)
	result := spaceRegex.ReplaceAllString(content, " ")

	// Replace multiple newlines with single newline
	newlineRegex := regexp.MustCompile(`\n\s*\n+`)
	result = newlineRegex.ReplaceAllString(result, "\n")

	// Trim each line
	lines := strings.Split(result, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimSpace(line)
	}

	// Remove empty lines
	var nonEmpty []string
	for _, line := range lines {
		if line == "" {
			continue
		}
		nonEmpty = append(nonEmpty, line)
	}

	return strings.Join(nonEmpty, "\n")
}

// extractKeywords extracts potential keywords from summary text
func extractKeywords(text string) []string {
	// Simple keyword extraction - look for capitalized terms and technical words
	wordRegex := regexp.MustCompile(`\b[A-Z][a-z]+(?:\s+[A-Z][a-z]+)*\b`)
	matches := wordRegex.FindAllString(text, -1)

	seen := make(map[string]bool)
	var keywords []string
	for _, match := range matches {
		lower := strings.ToLower(match)
		if seen[lower] {
			continue
		}
		// Skip common words
		if isCommonWord(lower) {
			continue
		}
		seen[lower] = true
		keywords = append(keywords, match)
	}

	// Limit to 5 keywords
	if len(keywords) > 5 {
		return keywords[:5]
	}
	return keywords
}

func isCommonWord(word string) bool {
	common := map[string]bool{
		"the": true, "this": true, "that": true, "with": true,
		"and": true, "for": true, "are": true, "but": true,
		"not": true, "you": true, "all": true, "can": true,
		"has": true, "her": true, "was": true, "one": true,
	}
	return common[word]
}
