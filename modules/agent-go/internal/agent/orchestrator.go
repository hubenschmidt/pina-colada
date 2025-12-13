package agent

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/nlpodyssey/openai-agents-go/agents"
	"github.com/nlpodyssey/openai-agents-go/usage"
	"github.com/pina-colada-co/agent-go/internal/agent/prompts"
	"github.com/pina-colada-co/agent-go/internal/agent/state"
	"github.com/pina-colada-co/agent-go/internal/agent/tools"
	"github.com/pina-colada-co/agent-go/internal/agent/workers"
	"github.com/pina-colada-co/agent-go/internal/config"
	"github.com/pina-colada-co/agent-go/internal/services"
)

// TokenUsage tracks token consumption for a turn or session
type TokenUsage struct {
	Input  int32
	Output int32
	Total  int32
}

// Per-session cumulative token tracking
var (
	sessionTokenTotals = make(map[string]*TokenUsage)
	sessionTokenMu     sync.RWMutex
)

// Orchestrator coordinates the agent system using openai-agents-go SDK
type Orchestrator struct {
	triageAgent     *agents.Agent
	stateManager    state.StateManager
	model           string
	anthropicAPIKey string
}

// NewOrchestrator creates the agent orchestrator with triage-based routing via handoffs
func NewOrchestrator(ctx context.Context, cfg *config.Config, indService *services.IndividualService, orgService *services.OrganizationService, docService *services.DocumentService, jobService *services.JobService) (*Orchestrator, error) {
	// Use OpenAI model (SDK defaults to OpenAI provider)
	model := "gpt-4.1-2025-04-14"

	// Create job service adapter for filtering applied jobs
	var jobAdapter tools.JobServiceInterface
	if jobService != nil {
		jobAdapter = &jobServiceAdapter{jobService: jobService}
	}

	// Create all tools using the adapter
	crmTools := tools.NewCRMTools(indService, orgService)
	serperTools := tools.NewSerperTools(cfg.SerperAPIKey, jobAdapter)
	docTools := tools.NewDocumentTools(docService)
	emailTools := tools.NewEmailTools(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUsername, cfg.SMTPPassword, cfg.SMTPFromEmail)
	allTools := tools.BuildAgentTools(crmTools, serperTools, docTools, emailTools)

	// Create worker agents
	jobSearchWorker := workers.NewJobSearchWorker(model, allTools)
	crmWorker := workers.NewCRMWorker(model, allTools)
	generalWorker := workers.NewGeneralWorker(model, allTools)

	// Create triage agent with handoffs to workers
	triageAgent := agents.New("triage").
		WithInstructions(prompts.TriageInstructionsWithTools).
		WithModel(model).
		WithHandoffDescription("Routes requests to specialized workers").
		WithAgentHandoffs(jobSearchWorker, crmWorker, generalWorker)

	// Create state manager
	stateMgr := state.NewMemoryStateManager()

	log.Printf("OpenAI Agents SDK orchestrator initialized with triage routing")

	return &Orchestrator{
		triageAgent:     triageAgent,
		stateManager:    stateMgr,
		model:           model,
		anthropicAPIKey: cfg.AnthropicAPIKey,
	}, nil
}

// RunRequest represents a request to run the agent
type RunRequest struct {
	SessionID    string
	UserID       string
	Message      string
	UseEvaluator bool
}

// RunResponse represents the agent's response
type RunResponse struct {
	Response         string
	SessionID        string
	Events           []Event
	TurnTokens       TokenUsage
	CumulativeTokens TokenUsage
}

// Event represents an agent event during execution
type Event struct {
	Type    string
	Content string
	Agent   string
}

// StreamEvent represents a real-time event sent during agent execution
type StreamEvent struct {
	Type string `json:"type"` // "start", "text", "tokens", "tool_start", "tool_end", "agent_start", "agent_end", "done", "error", "eval"

	// Text streaming
	Text string `json:"text,omitempty"`

	// Token usage
	Tokens *TokenUsage `json:"tokens,omitempty"`

	// Timing
	ElapsedMs int64 `json:"elapsed_ms,omitempty"`

	// Tool/Agent info
	ToolName  string `json:"tool_name,omitempty"`
	AgentName string `json:"agent_name,omitempty"`

	// Final response (on "done")
	Response         string      `json:"response,omitempty"`
	TurnTokens       *TokenUsage `json:"turn_tokens,omitempty"`
	CumulativeTokens *TokenUsage `json:"cumulative_tokens,omitempty"`

	// Error (on "error")
	Error string `json:"error,omitempty"`

	// Evaluation result (on "eval")
	EvalResult *EvaluatorResult `json:"eval_result,omitempty"`
}

// Run executes the agent with the given message (non-streaming)
func (o *Orchestrator) Run(ctx context.Context, req RunRequest) (*RunResponse, error) {
	log.Printf("Starting agent for thread: %s", req.SessionID)
	log.Printf("   Message: %s", req.Message)

	// Ensure session exists
	if err := o.ensureSession(ctx, req.UserID, req.SessionID); err != nil {
		return nil, err
	}

	// Get context from state manager
	contextMsgs, err := o.stateManager.GetMessages(ctx, req.SessionID, 8000)
	if err != nil {
		return nil, fmt.Errorf("get messages: %w", err)
	}

	// Build input with context
	input := buildInputWithContext(req.Message, contextMsgs)

	// Create usage context for token tracking
	ctx = usage.NewContext(ctx, usage.NewUsage())

	// Run the triage agent with increased turn limit (SDK handles handoffs automatically)
	runner := agents.Runner{Config: agents.RunConfig{MaxTurns: 20}}
	result, err := runner.Run(ctx, o.triageAgent, input)
	if err != nil {
		LogError("Agent run error: %v", err)
		return nil, fmt.Errorf("agent run error: %w", err)
	}

	// Extract final output
	finalOutput := extractFinalOutput(result)

	// Get token usage from context
	turnTokens := extractTokenUsage(ctx)

	// Store messages
	if err := o.stateManager.AddMessage(ctx, req.SessionID, state.Message{Role: "user", Content: req.Message}); err != nil {
		log.Printf("Warning: failed to store user message: %v", err)
	}
	if err := o.stateManager.AddMessage(ctx, req.SessionID, state.Message{Role: "assistant", Content: finalOutput}); err != nil {
		log.Printf("Warning: failed to store assistant message: %v", err)
	}

	// Update cumulative tokens
	cumulativeTokens := updateCumulativeTokens(req.SessionID, turnTokens)

	log.Printf("Token usage - turn: in=%d out=%d total=%d, cumulative: in=%d out=%d total=%d",
		turnTokens.Input, turnTokens.Output, turnTokens.Total,
		cumulativeTokens.Input, cumulativeTokens.Output, cumulativeTokens.Total)
	log.Println("Agent execution completed")

	return &RunResponse{
		Response:         finalOutput,
		SessionID:        req.SessionID,
		TurnTokens:       turnTokens,
		CumulativeTokens: cumulativeTokens,
	}, nil
}

// streamState holds mutable state for streaming event processing
type streamState struct {
	finalOutput  string
	turnTokens   TokenUsage
	currentAgent string
	bufferedText string
	useEvaluator bool
	sendEvent    func(StreamEvent)
}

func (s *streamState) handleAgentUpdated(e agents.AgentUpdatedStreamEvent) {
	if e.NewAgent == nil {
		return
	}
	if e.NewAgent.Name == s.currentAgent {
		return
	}

	if s.currentAgent == "" {
		LogAgentStart(e.NewAgent.Name)
	}
	if s.currentAgent != "" {
		LogHandoff(s.currentAgent, e.NewAgent.Name)
	}
	s.sendEvent(StreamEvent{Type: "agent_start", AgentName: e.NewAgent.Name})
	s.currentAgent = e.NewAgent.Name
}

func (s *streamState) handleToolCalled(item agents.ToolCallItem) {
	toolName := extractToolName(item)
	LogToolStart(toolName)
	s.sendEvent(StreamEvent{Type: "tool_start", ToolName: toolName})
}

func (s *streamState) handleToolOutput(item agents.ToolCallOutputItem) {
	toolName := extractOutputToolName(item)
	outputStr := ""
	if str, ok := item.Output.(string); ok {
		outputStr = truncateString(str, 100)
	}
	LogToolEnd(toolName, outputStr)
	s.sendEvent(StreamEvent{Type: "tool_end", ToolName: toolName})
}

func (s *streamState) handleMessageOutput(item agents.MessageOutputItem) {
	text := extractTextFromMessageOutput(item)
	if text == "" {
		return
	}
	s.finalOutput = text
}

func (s *streamState) handleRunItemEvent(e agents.RunItemStreamEvent) {
	if e.Name == agents.StreamEventToolCalled {
		item, ok := e.Item.(agents.ToolCallItem)
		if !ok {
			return
		}
		s.handleToolCalled(item)
		return
	}

	if e.Name == agents.StreamEventToolOutput {
		item, ok := e.Item.(agents.ToolCallOutputItem)
		if !ok {
			return
		}
		s.handleToolOutput(item)
		return
	}

	if e.Name == agents.StreamEventMessageOutputCreated {
		item, ok := e.Item.(agents.MessageOutputItem)
		if !ok {
			return
		}
		s.handleMessageOutput(item)
	}
}

func (s *streamState) handleTextDelta(delta string) {
	if s.useEvaluator {
		s.bufferedText += delta
		return
	}
	s.sendEvent(StreamEvent{Type: "text", Text: delta})
}

func (s *streamState) handleResponseCompleted(resp agents.RawResponsesStreamEvent) {
	s.turnTokens.Input = int32(resp.Data.Response.Usage.InputTokens)
	s.turnTokens.Output = int32(resp.Data.Response.Usage.OutputTokens)
	s.turnTokens.Total = int32(resp.Data.Response.Usage.TotalTokens)
	s.sendEvent(StreamEvent{
		Type:   "tokens",
		Tokens: &TokenUsage{Input: s.turnTokens.Input, Output: s.turnTokens.Output, Total: s.turnTokens.Total},
	})
}

func (s *streamState) handleRawResponse(e agents.RawResponsesStreamEvent) {
	if e.Data.Type == "response.output_text.delta" && e.Data.Delta != "" {
		s.handleTextDelta(e.Data.Delta)
		return
	}
	if e.Data.Type == "response.completed" && e.Data.Response.Usage.TotalTokens > 0 {
		s.handleResponseCompleted(e)
	}
}

func (s *streamState) handleStreamEvent(event agents.StreamEvent) {
	if e, ok := event.(agents.AgentUpdatedStreamEvent); ok {
		s.handleAgentUpdated(e)
		return
	}
	if e, ok := event.(agents.RunItemStreamEvent); ok {
		s.handleRunItemEvent(e)
		return
	}
	if e, ok := event.(agents.RawResponsesStreamEvent); ok {
		s.handleRawResponse(e)
	}
}

// RunWithStreaming executes the agent and streams events via channel
func (o *Orchestrator) RunWithStreaming(ctx context.Context, req RunRequest, eventCh chan<- StreamEvent) {
	startTime := time.Now()
	defer close(eventCh)

	log.Printf("Starting streaming agent for thread: %s", req.SessionID)
	log.Printf("   Message: %s", req.Message)

	sendEvent := func(evt StreamEvent) {
		evt.ElapsedMs = time.Since(startTime).Milliseconds()
		select {
		case eventCh <- evt:
		case <-ctx.Done():
		}
	}

	// Ensure session exists
	if err := o.ensureSession(ctx, req.UserID, req.SessionID); err != nil {
		sendEvent(StreamEvent{Type: "error", Error: err.Error()})
		return
	}

	// Get context from state manager
	contextMsgs, err := o.stateManager.GetMessages(ctx, req.SessionID, 8000)
	if err != nil {
		sendEvent(StreamEvent{Type: "error", Error: fmt.Sprintf("get messages: %v", err)})
		return
	}

	// Build input with context
	input := buildInputWithContext(req.Message, contextMsgs)

	// Create usage context for token tracking
	ctx = usage.NewContext(ctx, usage.NewUsage())

	// Send start event so frontend can begin timer
	sendEvent(StreamEvent{Type: "start"})

	// Run streamed with increased turn limit
	runner := agents.Runner{Config: agents.RunConfig{MaxTurns: 20}}
	eventsChan, errChan, err := runner.RunStreamedChan(ctx, o.triageAgent, input)
	if err != nil {
		sendEvent(StreamEvent{Type: "error", Error: err.Error()})
		return
	}

	ss := &streamState{
		useEvaluator: req.UseEvaluator,
		sendEvent:    sendEvent,
	}

	// Process events
	for event := range eventsChan {
		ss.handleStreamEvent(event)
	}

	finalOutput := ss.finalOutput
	turnTokens := ss.turnTokens
	currentAgent := ss.currentAgent
	bufferedText := ss.bufferedText

	// Check for streaming error
	if streamErr := <-errChan; streamErr != nil {
		LogError("Streaming error: %v", streamErr)
		sendEvent(StreamEvent{Type: "error", Error: streamErr.Error()})
		return
	}

	// Use buffered text for final output when evaluator is enabled
	if req.UseEvaluator && bufferedText != "" {
		finalOutput = bufferedText
	}

	if currentAgent != "" {
		LogAgentEnd(currentAgent, finalOutput, &turnTokens)
	}

	// Run evaluator if enabled - evaluate first, then send response
	if req.UseEvaluator && finalOutput != "" && o.anthropicAPIKey != "" {
		finalOutput = o.runEvaluation(ctx, req, input, finalOutput, currentAgent, sendEvent)
		// Send the final response after evaluation completes
		sendEvent(StreamEvent{Type: "text", Text: finalOutput})
	}

	// Store messages
	if err := o.stateManager.AddMessage(ctx, req.SessionID, state.Message{Role: "user", Content: req.Message}); err != nil {
		log.Printf("Warning: failed to store user message: %v", err)
	}
	if err := o.stateManager.AddMessage(ctx, req.SessionID, state.Message{Role: "assistant", Content: finalOutput}); err != nil {
		log.Printf("Warning: failed to store assistant message: %v", err)
	}

	// Update cumulative tokens
	cumulativeTokens := updateCumulativeTokens(req.SessionID, turnTokens)

	log.Printf("Token usage - turn: in=%d out=%d total=%d, cumulative: in=%d out=%d total=%d",
		turnTokens.Input, turnTokens.Output, turnTokens.Total,
		cumulativeTokens.Input, cumulativeTokens.Output, cumulativeTokens.Total)
	log.Println("Agent execution completed")

	// Send done event
	sendEvent(StreamEvent{
		Type:             "done",
		Response:         finalOutput,
		TurnTokens:       &turnTokens,
		CumulativeTokens: &cumulativeTokens,
	})
}

// ensureSession creates a session if it doesn't exist
func (o *Orchestrator) ensureSession(ctx context.Context, userID, sessionID string) error {
	session, err := o.stateManager.GetSession(ctx, userID, sessionID)
	if err != nil {
		return fmt.Errorf("get session: %w", err)
	}
	if session != nil {
		return nil
	}

	_, err = o.stateManager.CreateSession(ctx, userID, sessionID)
	if err != nil {
		return fmt.Errorf("create session: %w", err)
	}
	return nil
}

// buildInputWithContext creates the input string with conversation context
func buildInputWithContext(message string, contextMsgs []state.Message) string {
	if len(contextMsgs) == 0 {
		return message
	}

	// Include context as a preamble
	var contextStr string
	for _, msg := range contextMsgs {
		contextStr += fmt.Sprintf("%s: %s\n", msg.Role, msg.Content)
	}

	return fmt.Sprintf("Previous conversation:\n%s\nCurrent message: %s", contextStr, message)
}

// extractFinalOutput gets the text output from a RunResult
func extractFinalOutput(result *agents.RunResult) string {
	if result == nil {
		return ""
	}

	// Try to get FinalOutput as string
	if str, ok := result.FinalOutput.(string); ok {
		return str
	}

	// Fall back to extracting text from NewItems
	return agents.ItemHelpers().TextMessageOutputs(result.NewItems)
}

// extractTokenUsage gets token usage from context
func extractTokenUsage(ctx context.Context) TokenUsage {
	u, ok := usage.FromContext(ctx)
	if !ok || u == nil {
		return TokenUsage{}
	}
	return TokenUsage{
		Input:  int32(u.InputTokens),
		Output: int32(u.OutputTokens),
		Total:  int32(u.TotalTokens),
	}
}

// updateCumulativeTokens updates and returns cumulative token counts
func updateCumulativeTokens(sessionID string, turnTokens TokenUsage) TokenUsage {
	sessionTokenMu.Lock()
	defer sessionTokenMu.Unlock()

	if sessionTokenTotals[sessionID] == nil {
		sessionTokenTotals[sessionID] = &TokenUsage{}
	}
	sessionTokenTotals[sessionID].Input += turnTokens.Input
	sessionTokenTotals[sessionID].Output += turnTokens.Output
	sessionTokenTotals[sessionID].Total += turnTokens.Total
	return *sessionTokenTotals[sessionID]
}

// extractTextFromMessageOutput extracts text from a MessageOutputItem
func extractTextFromMessageOutput(item agents.MessageOutputItem) string {
	for _, content := range item.RawItem.Content {
		if content.Text != "" {
			return content.Text
		}
	}
	return ""
}

// extractToolName gets tool name from ToolCallItem
func extractToolName(item agents.ToolCallItem) string {
	if fc, ok := item.RawItem.(agents.ResponseFunctionToolCall); ok {
		return fc.Name
	}
	return "unknown"
}

// extractOutputToolName gets tool name from ToolCallOutputItem
func extractOutputToolName(item agents.ToolCallOutputItem) string {
	if fc, ok := item.RawItem.(agents.ResponseInputItemFunctionCallOutputParam); ok {
		return fc.CallID
	}
	return "unknown"
}

// truncateString truncates a string to maxLen characters
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// jobServiceAdapter adapts JobService to JobServiceInterface for serper tools
type jobServiceAdapter struct {
	jobService *services.JobService
}

func (a *jobServiceAdapter) GetLeads(statusNames []string, tenantID *int64) ([]tools.JobInfo, error) {
	leads, err := a.jobService.GetLeads(statusNames, tenantID)
	if err != nil {
		return nil, err
	}

	result := make([]tools.JobInfo, len(leads))
	for i, lead := range leads {
		result[i] = tools.JobInfo{
			JobTitle: lead.JobTitle,
			Account:  lead.Account,
		}
	}
	return result, nil
}

// workerToEvaluatorType maps worker agent names to evaluator types
func workerToEvaluatorType(workerName string) EvaluatorType {
	if workerName == "job_search" {
		return CareerEvaluator
	}
	if workerName == "crm_worker" {
		return CRMEvaluator
	}
	return GeneralEvaluator
}

// runEvaluation runs the evaluator and handles retry logic
// Returns the final output (original or retried) after evaluation completes
func (o *Orchestrator) runEvaluation(ctx context.Context, req RunRequest, input, finalOutput, currentAgent string, sendEvent func(StreamEvent)) string {
	evalType := workerToEvaluatorType(currentAgent)
	evaluator := NewEvaluator(o.anthropicAPIKey, evalType)

	sendEvent(StreamEvent{Type: "eval_start"})

	evalResult, err := evaluator.Evaluate(ctx, req.Message, finalOutput, "")
	if err != nil {
		log.Printf("Evaluator error: %v", err)
		return finalOutput
	}

	// Pass if score >= 60 or user input needed
	if evalResult.Score >= 60 || evalResult.UserInputNeeded {
		log.Printf("✅ Evaluator PASSED (score: %d)", evalResult.Score)
		sendEvent(StreamEvent{Type: "eval", EvalResult: evalResult})
		return finalOutput
	}

	log.Printf("⚠️ Evaluator REJECTED response - agent will retry!")

	// Retry needed
	log.Printf("Evaluation failed (score: %d), retrying...", evalResult.Score)
	log.Printf("📝 FEEDBACK TO AGENT: %s", evalResult.Feedback)
	evaluator.IncrementRetry()

	retryInput := fmt.Sprintf("%s\n\nPrevious attempt feedback: %s\nPlease improve the response.", input, evalResult.Feedback)
	retryResult, retryErr := agents.Run(ctx, o.triageAgent, retryInput)
	if retryErr != nil {
		log.Printf("Retry error: %v", retryErr)
		sendEvent(StreamEvent{Type: "eval", EvalResult: evalResult})
		return finalOutput
	}

	retryOutput := extractFinalOutput(retryResult)
	if retryOutput == "" {
		sendEvent(StreamEvent{Type: "eval", EvalResult: evalResult})
		return finalOutput
	}

	// Re-evaluate the retry output
	evalResult, _ = evaluator.Evaluate(ctx, req.Message, retryOutput, "")
	sendEvent(StreamEvent{Type: "eval", EvalResult: evalResult})
	return retryOutput
}
