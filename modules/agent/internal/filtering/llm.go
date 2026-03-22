package filtering

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
)

// ReviewWithAnthropic sends results to Claude for review and returns reviewed results
func ReviewWithAnthropic(client anthropic.Client, model string, systemPrompt, userPrompt string) []ReviewedJobResult {
	params := anthropic.MessageNewParams{
		Model:     anthropic.Model(model),
		MaxTokens: 2048,
		System: []anthropic.TextBlockParam{
			{Text: systemPrompt},
		},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(userPrompt)),
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	log.Printf("[Anthropic] Calling API - model=%s", model)
	start := time.Now()
	resp, err := client.Messages.New(ctx, params)
	if err != nil {
		log.Printf("[Anthropic] API error after %v: %v", time.Since(start), err)
		return nil
	}
	log.Printf("[Anthropic] API returned in %v", time.Since(start))

	return ParseAgentResponse(resp)
}

// BuildAgentSystemPrompt builds the system prompt for agent review
func BuildAgentSystemPrompt(customPrompt *string) string {
	base := `You are evaluating search results for relevance to a candidate's profile. Review each result and determine if it's a strong match.

Return ONLY a JSON object with this structure:
{
  "reviews": [
    {"index": 0, "approved": true, "reason": "Brief explanation"},
    {"index": 1, "approved": false, "reason": "Brief explanation"}
  ]
}

Be selective - only approve results that are genuinely relevant matches.`

	if customPrompt != nil && *customPrompt != "" {
		return base + "\n\nAdditional Instructions:\n" + *customPrompt
	}
	return base
}

// BuildAgentUserPrompt builds the user prompt with document context and results
func BuildAgentUserPrompt(documentContext string, results []JobResult) string {
	var sb strings.Builder

	if documentContext != "" {
		sb.WriteString("## Candidate Profile/Resume:\n")
		sb.WriteString(documentContext)
		sb.WriteString("\n\n")
	}

	sb.WriteString("## Search Results to Evaluate:\n")
	for i, r := range results {
		sb.WriteString(fmt.Sprintf("\n[%d] %s\nCompany: %s\nURL: %s\nSnippet: %s\n", i, r.Title, r.Company, r.URL, r.Snippet))
		sb.WriteString(TruncatedJobDescription(r.FullText, 3000))
	}

	sb.WriteString("\n\nEvaluate each result. Return JSON with reviews array.")
	return sb.String()
}

// ParseAgentResponse parses the LLM response into reviewed results
func ParseAgentResponse(resp *anthropic.Message) []ReviewedJobResult {
	if resp == nil || len(resp.Content) == 0 {
		return nil
	}

	text := ExtractTextFromResponse(resp)
	text = ExtractJSONFromText(text)

	var reviewResp AgentReviewResponse
	if err := json.Unmarshal([]byte(text), &reviewResp); err != nil {
		log.Printf("Filtering: failed to parse agent response: %v", err)
		return nil
	}

	reviewed := make([]ReviewedJobResult, 0, len(reviewResp.Reviews))
	for _, r := range reviewResp.Reviews {
		reviewed = append(reviewed, ReviewedJobResult{
			Approved: r.Approved,
			Reason:   r.Reason,
		})
	}
	return reviewed
}

// ApplyReviewsToResults maps review decisions onto job results
func ApplyReviewsToResults(results []JobResult, reviews []AgentReview) []ReviewedJobResult {
	reviewed := make([]ReviewedJobResult, len(results))
	for i, r := range results {
		reviewed[i] = ReviewedJobResult{JobResult: r, Approved: true, Reason: ""}
	}

	for _, review := range reviews {
		if review.Index < 0 || review.Index >= len(reviewed) {
			continue
		}
		reviewed[review.Index].Approved = review.Approved
		reviewed[review.Index].Reason = review.Reason
	}

	return reviewed
}

// ApproveAllResults returns all results as approved (fallback when agent is unavailable)
func ApproveAllResults(results []JobResult) []ReviewedJobResult {
	reviewed := make([]ReviewedJobResult, len(results))
	for i, r := range results {
		reviewed[i] = ReviewedJobResult{JobResult: r, Approved: true, Reason: ""}
	}
	return reviewed
}

// ExtractTextFromResponse extracts text content from an Anthropic message
func ExtractTextFromResponse(resp *anthropic.Message) string {
	var text string
	for _, block := range resp.Content {
		if block.Type == "text" {
			text += block.Text
		}
	}
	return strings.TrimSpace(text)
}

// ExtractJSONFromText strips markdown code fences and extracts JSON
func ExtractJSONFromText(text string) string {
	if strings.HasPrefix(text, "```json") {
		text = strings.TrimPrefix(text, "```json")
		if idx := strings.Index(text, "```"); idx > 0 {
			text = text[:idx]
		}
	}
	if strings.HasPrefix(text, "```") {
		text = strings.TrimPrefix(text, "```")
		if idx := strings.Index(text, "```"); idx > 0 {
			text = text[:idx]
		}
	}
	startIdx := strings.Index(text, "{")
	endIdx := strings.LastIndex(text, "}")
	if startIdx >= 0 && endIdx > startIdx {
		text = text[startIdx : endIdx+1]
	}
	return strings.TrimSpace(text)
}

// LogAgentReviews logs a summary of agent review decisions
func LogAgentReviews(reviews []ReviewedJobResult) {
	approved := 0
	for _, r := range reviews {
		if r.Approved {
			approved++
		}
		if !r.Approved {
			log.Printf("Filtering: rejected: \"%s\" - %s", r.Title, r.Reason)
		}
	}
	log.Printf("Filtering: agent approved %d/%d results", approved, len(reviews))
}
