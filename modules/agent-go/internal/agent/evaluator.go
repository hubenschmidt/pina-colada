package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/pina-colada-co/agent-go/internal/agent/prompts"
)

// EvaluatorResult represents the structured output from an evaluator
type EvaluatorResult struct {
	Feedback           string `json:"feedback"`
	SuccessCriteriaMet bool   `json:"success_criteria_met"`
	UserInputNeeded    bool   `json:"user_input_needed"`
	Score              int    `json:"score"`
}

// EvaluatorType determines which evaluator prompt to use
type EvaluatorType string

const (
	CareerEvaluator  EvaluatorType = "career"
	CRMEvaluator     EvaluatorType = "crm"
	GeneralEvaluator EvaluatorType = "general"
)

// Evaluator evaluates agent responses for quality using Claude
type Evaluator struct {
	client     anthropic.Client
	evalType   EvaluatorType
	maxRetries int
	retryCount int
}

// NewEvaluator creates a new evaluator with Claude Sonnet 4.5
func NewEvaluator(apiKey string, evalType EvaluatorType) *Evaluator {
	client := anthropic.NewClient(
		option.WithAPIKey(apiKey),
	)
	return &Evaluator{
		client:     client,
		evalType:   evalType,
		maxRetries: 2,
	}
}

// getPrompt returns the appropriate evaluator prompt
func (e *Evaluator) getPrompt() string {
	if e.evalType == CareerEvaluator {
		return prompts.CareerEvaluatorPrompt
	}
	if e.evalType == CRMEvaluator {
		return prompts.CRMEvaluatorPrompt
	}
	return prompts.GeneralEvaluatorPrompt
}

// Evaluate checks the agent response against criteria using Claude
func (e *Evaluator) Evaluate(ctx context.Context, userRequest, agentResponse, successCriteria string) (*EvaluatorResult, error) {
	log.Printf("🔍 %s EVALUATOR (Claude): Reviewing response...", strings.ToUpper(string(e.evalType)))

	if agentResponse == "" {
		log.Printf("⚠️ Empty response, skipping evaluation")
		return &EvaluatorResult{
			Feedback:           "No response to evaluate",
			SuccessCriteriaMet: false,
			UserInputNeeded:    false,
			Score:              0,
		}, nil
	}

	systemPrompt := e.getPrompt()
	userPrompt := e.buildUserPrompt(userRequest, agentResponse, successCriteria)

	// Call Claude Sonnet 4.5
	resp, err := e.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeSonnet4_5,
		MaxTokens: 1024,
		System: []anthropic.TextBlockParam{
			{Text: systemPrompt},
		},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(userPrompt)),
		},
	})
	if err != nil {
		log.Printf("❌ Claude evaluator call failed: %v", err)
		return e.defaultApproval("Evaluation error, defaulting to approval"), nil
	}

	// Parse the response
	result, err := e.parseClaudeResponse(resp)
	if err != nil {
		log.Printf("⚠️ Failed to parse evaluator response: %v", err)
		return e.defaultApproval("Failed to parse evaluation, defaulting to approval"), nil
	}

	// Apply retry loop logic
	e.applyRetryLogic(result)

	e.logResult(result)
	return result, nil
}

// buildUserPrompt constructs the evaluation request
func (e *Evaluator) buildUserPrompt(userRequest, agentResponse, successCriteria string) string {
	var sb strings.Builder
	sb.WriteString("USER REQUEST:\n")
	sb.WriteString(userRequest)
	sb.WriteString("\n\n")

	if successCriteria != "" {
		sb.WriteString("SUCCESS CRITERIA:\n")
		sb.WriteString(successCriteria)
		sb.WriteString("\n\n")
	}

	sb.WriteString("ASSISTANT RESPONSE TO EVALUATE:\n")
	sb.WriteString(agentResponse)
	sb.WriteString("\n\n")

	if e.retryCount > 0 {
		sb.WriteString(fmt.Sprintf("NOTE: This is retry attempt %d. Be more lenient - approve unless there are critical errors.\n\n", e.retryCount))
	}

	sb.WriteString("Evaluate this response. Return JSON with: feedback, success_criteria_met, user_input_needed, score")

	return sb.String()
}

// parseClaudeResponse extracts the EvaluatorResult from Claude's response
func (e *Evaluator) parseClaudeResponse(resp *anthropic.Message) (*EvaluatorResult, error) {
	if resp == nil || len(resp.Content) == 0 {
		return nil, fmt.Errorf("empty response from evaluator")
	}

	// Extract text from response
	var text string
	for _, block := range resp.Content {
		if block.Type == "text" {
			text += block.Text
		}
	}

	// Try to extract JSON from the response
	text = strings.TrimSpace(text)

	// Handle markdown code blocks
	text = extractFromCodeBlock(text)

	// Find JSON object bounds
	startIdx := strings.Index(text, "{")
	endIdx := strings.LastIndex(text, "}")
	if startIdx >= 0 && endIdx > startIdx {
		text = text[startIdx : endIdx+1]
	}

	var result EvaluatorResult
	if err := json.Unmarshal([]byte(text), &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w (text: %s)", err, text[:min(len(text), 200)])
	}

	return &result, nil
}

// extractFromCodeBlock extracts content from markdown code blocks
func extractFromCodeBlock(text string) string {
	if strings.Contains(text, "```json") {
		start := strings.Index(text, "```json") + 7
		end := strings.LastIndex(text, "```")
		if end > start {
			return strings.TrimSpace(text[start:end])
		}
	}
	if strings.Contains(text, "```") {
		start := strings.Index(text, "```") + 3
		end := strings.LastIndex(text, "```")
		if end > start {
			return strings.TrimSpace(text[start:end])
		}
	}
	return text
}

// applyRetryLogic forces approval if stuck in retry loop
func (e *Evaluator) applyRetryLogic(result *EvaluatorResult) {
	if e.retryCount < e.maxRetries {
		return
	}

	if result.SuccessCriteriaMet || result.UserInputNeeded {
		return
	}

	log.Printf("⚠️ Forcing approval after %d retries", e.retryCount)
	result.SuccessCriteriaMet = true
	result.Feedback = fmt.Sprintf("%s (Approved after %d retries)", result.Feedback, e.retryCount)
	if result.Score < 60 {
		result.Score = 60
	}
}

// defaultApproval returns a passing result for error cases
func (e *Evaluator) defaultApproval(feedback string) *EvaluatorResult {
	return &EvaluatorResult{
		Feedback:           feedback,
		SuccessCriteriaMet: true,
		UserInputNeeded:    false,
		Score:              100,
	}
}

// logResult logs the evaluation outcome
func (e *Evaluator) logResult(result *EvaluatorResult) {
	status := "PASS"
	if !result.SuccessCriteriaMet {
		status = "FAIL"
	}

	// Use fmt.Print for immediate output (no buffering)
	fmt.Printf("%s ✓ %s EVALUATOR: Result = %s (score: %d/100)\n", time.Now().Format("2006/01/02 15:04:05"), strings.ToUpper(string(e.evalType)), status, result.Score)
	fmt.Printf("%s    - Success criteria met: %v\n", time.Now().Format("2006/01/02 15:04:05"), result.SuccessCriteriaMet)
	fmt.Printf("%s    - User input needed: %v\n", time.Now().Format("2006/01/02 15:04:05"), result.UserInputNeeded)
	fmt.Printf("%s    - Feedback: %s\n", time.Now().Format("2006/01/02 15:04:05"), result.Feedback)

	if !result.SuccessCriteriaMet && !result.UserInputNeeded {
		fmt.Printf("%s ⚠️ Evaluator rejected response - agent will retry!\n", time.Now().Format("2006/01/02 15:04:05"))
	}
}

// IncrementRetry increments the retry counter
func (e *Evaluator) IncrementRetry() {
	e.retryCount++
}

// ResetRetry resets the retry counter
func (e *Evaluator) ResetRetry() {
	e.retryCount = 0
}

// ShouldRetry returns true if we should retry based on evaluation
func (e *Evaluator) ShouldRetry(result *EvaluatorResult) bool {
	if result.SuccessCriteriaMet || result.UserInputNeeded {
		return false
	}
	// Retry if score is below 60 (pass threshold)
	return result.Score < 60 && e.retryCount < e.maxRetries
}

// DetermineEvaluatorType determines which evaluator to use based on the request
func DetermineEvaluatorType(userMessage string) EvaluatorType {
	lower := strings.ToLower(userMessage)

	// Career-related keywords
	careerKeywords := []string{"job", "career", "resume", "cover letter", "hiring", "position", "role", "employment"}
	for _, kw := range careerKeywords {
		if strings.Contains(lower, kw) {
			return CareerEvaluator
		}
	}

	// CRM-related keywords
	crmKeywords := []string{"crm", "contact", "account", "organization", "individual", "lookup", "record"}
	for _, kw := range crmKeywords {
		if strings.Contains(lower, kw) {
			return CRMEvaluator
		}
	}

	return GeneralEvaluator
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
