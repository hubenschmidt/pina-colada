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
	triageAgent  *agents.Agent
	stateManager state.StateManager
	model        string
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
	allTools := tools.BuildAgentTools(crmTools, serperTools, docTools)

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
		triageAgent:  triageAgent,
		stateManager: stateMgr,
		model:        model,
	}, nil
}

// RunRequest represents a request to run the agent
type RunRequest struct {
	SessionID string
	UserID    string
	Message   string
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
	Type string `json:"type"` // "text", "tokens", "tool_start", "tool_end", "agent_start", "agent_end", "done", "error"

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

	// Run the triage agent (SDK handles handoffs automatically)
	result, err := agents.Run(ctx, o.triageAgent, input)
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

	// Run streamed
	eventsChan, errChan, err := agents.RunStreamedChan(ctx, o.triageAgent, input)
	if err != nil {
		sendEvent(StreamEvent{Type: "error", Error: err.Error()})
		return
	}

	var finalOutput string
	var turnTokens TokenUsage
	var currentAgent string

	// Process events
	for event := range eventsChan {
		switch e := event.(type) {
		case agents.AgentUpdatedStreamEvent:
			if e.NewAgent != nil && e.NewAgent.Name != currentAgent {
				if currentAgent != "" {
					LogHandoff(currentAgent, e.NewAgent.Name)
				} else {
					LogAgentStart(e.NewAgent.Name)
				}
				sendEvent(StreamEvent{Type: "agent_start", AgentName: e.NewAgent.Name})
				currentAgent = e.NewAgent.Name
			}

		case agents.RunItemStreamEvent:
			switch e.Name {
			case agents.StreamEventToolCalled:
				if item, ok := e.Item.(agents.ToolCallItem); ok {
					toolName := extractToolName(item)
					LogToolStart(toolName)
					sendEvent(StreamEvent{Type: "tool_start", ToolName: toolName})
				}
			case agents.StreamEventToolOutput:
				if item, ok := e.Item.(agents.ToolCallOutputItem); ok {
					toolName := extractOutputToolName(item)
					outputStr := ""
					if s, ok := item.Output.(string); ok {
						outputStr = truncateString(s, 100)
					}
					LogToolEnd(toolName, outputStr)
					sendEvent(StreamEvent{Type: "tool_end", ToolName: toolName})
				}
			case agents.StreamEventMessageOutputCreated:
				// Capture final output (don't stream - we use deltas for streaming)
				if item, ok := e.Item.(agents.MessageOutputItem); ok {
					text := extractTextFromMessageOutput(item)
					if text != "" {
						finalOutput = text
					}
				}
			}

		case agents.RawResponsesStreamEvent:
			// Stream text deltas as they arrive
			if e.Data.Type == "response.output_text.delta" && e.Data.Delta != "" {
				sendEvent(StreamEvent{Type: "text", Text: e.Data.Delta})
			}
			// Extract token usage from completed response
			if e.Data.Type == "response.completed" && e.Data.Response.Usage.TotalTokens > 0 {
				turnTokens.Input = int32(e.Data.Response.Usage.InputTokens)
				turnTokens.Output = int32(e.Data.Response.Usage.OutputTokens)
				turnTokens.Total = int32(e.Data.Response.Usage.TotalTokens)
				sendEvent(StreamEvent{
					Type:   "tokens",
					Tokens: &TokenUsage{Input: turnTokens.Input, Output: turnTokens.Output, Total: turnTokens.Total},
				})
			}
		}
	}

	// Check for streaming error
	if streamErr := <-errChan; streamErr != nil {
		LogError("Streaming error: %v", streamErr)
		sendEvent(StreamEvent{Type: "error", Error: streamErr.Error()})
		return
	}

	if currentAgent != "" {
		LogAgentEnd(currentAgent, finalOutput, &turnTokens)
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
