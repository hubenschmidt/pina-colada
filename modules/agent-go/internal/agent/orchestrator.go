package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/nlpodyssey/openai-agents-go/agents"
	"github.com/nlpodyssey/openai-agents-go/modelsettings"
	"github.com/nlpodyssey/openai-agents-go/usage"
	"github.com/openai/openai-go/v2/packages/param"
	"gorm.io/datatypes"

	"github.com/pina-colada-co/agent-go/internal/agent/prompts"
	"github.com/pina-colada-co/agent-go/internal/agent/state"
	"github.com/pina-colada-co/agent-go/internal/agent/tools"
	"github.com/pina-colada-co/agent-go/internal/agent/utils"
	"github.com/pina-colada-co/agent-go/internal/agent/workers"
	"github.com/pina-colada-co/agent-go/internal/config"
	apperrors "github.com/pina-colada-co/agent-go/internal/errors"
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
	triageAgent        *agents.Agent
	stateManager       state.StateManager
	model              string
	anthropicAPIKey    string
	convService        *services.ConversationService
	configCache        *utils.ConfigCache
	allTools           []agents.Tool
	metricService      *services.MetricService
	serperTools        *tools.SerperTools
	cacheTools         *tools.CacheTools
	openaiProvider     *agents.OpenAIProvider
	anthropicProvider  *agents.OpenAIProvider
}

// NewOrchestrator creates the agent orchestrator with triage-based routing via handoffs
func NewOrchestrator(ctx context.Context, cfg *config.Config, indService *services.IndividualService, orgService *services.OrganizationService, docService *services.DocumentService, jobService *services.JobService, convService *services.ConversationService, configCache *utils.ConfigCache, metricService *services.MetricService, cacheRepo tools.CacheRepositoryInterface) (*Orchestrator, error) {
	// Create providers for both OpenAI and Anthropic
	// The correct provider is selected at runtime based on the model's provider setting
	openaiProvider := agents.NewOpenAIProvider(agents.OpenAIProviderParams{
		APIKey: param.NewOpt(cfg.OpenAIAPIKey),
		// UseResponses defaults to true (responses API) - has proper token tracking
	})

	var anthropicProvider *agents.OpenAIProvider
	if cfg.AnthropicAPIKey != "" {
		anthropicProvider = agents.NewOpenAIProvider(agents.OpenAIProviderParams{
			BaseURL:      param.NewOpt("https://api.anthropic.com/v1/"),
			APIKey:       param.NewOpt(cfg.AnthropicAPIKey),
			UseResponses: param.NewOpt(false), // Anthropic only supports chat completions
		})
		log.Printf("Initialized Anthropic provider (OpenAI-compatible endpoint, chat completions mode)")
	}

	log.Printf("Initialized OpenAI provider (responses API mode)")

	// Use configured model as default
	model := cfg.OpenAIModel

	// Create job service adapter for filtering applied jobs
	var jobAdapter tools.JobServiceInterface
	if jobService != nil {
		jobAdapter = &jobServiceAdapter{jobService: jobService}
	}

	// Create all tools using the adapter
	crmTools := tools.NewCRMTools(indService, orgService)
	cacheTools := tools.NewCacheTools(cacheRepo)
	serperTools := tools.NewSerperTools(cfg.SerperAPIKey, jobAdapter, cacheTools, cfg.PublicURL)
	docTools := tools.NewDocumentTools(docService)
	emailTools := tools.NewEmailTools(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUsername, cfg.SMTPPassword, cfg.SMTPFromEmail)
	allTools := tools.BuildAgentTools(crmTools, serperTools, docTools, emailTools, cacheTools)

	// Create worker agents with default model (no custom settings)
	jobSearchWorker := workers.NewJobSearchWorker(model, nil, allTools)
	crmWorker := workers.NewCRMWorker(model, nil, allTools)
	generalWorker := workers.NewGeneralWorker(model, nil, allTools)

	// Create triage agent with handoffs to workers (default config)
	triageAgent := agents.New("triage").
		WithInstructions(prompts.TriageInstructionsWithTools).
		WithModel(model).
		WithHandoffDescription("Routes requests to specialized workers").
		WithAgentHandoffs(jobSearchWorker, crmWorker, generalWorker)

	// Create state manager
	stateMgr := state.NewMemoryStateManager()

	log.Printf("OpenAI Agents SDK orchestrator initialized with triage routing")

	return &Orchestrator{
		triageAgent:       triageAgent,
		stateManager:      stateMgr,
		model:             model,
		anthropicAPIKey:   cfg.AnthropicAPIKey,
		convService:       convService,
		configCache:       configCache,
		allTools:          allTools,
		metricService:     metricService,
		serperTools:       serperTools,
		cacheTools:        cacheTools,
		openaiProvider:    openaiProvider,
		anthropicProvider: anthropicProvider,
	}, nil
}

// buildModelSettings creates ModelSettings from services.LLMSettings
func buildModelSettings(settings services.LLMSettings) *modelsettings.ModelSettings {
	return buildModelSettingsWithOptions(settings, true)
}

// buildModelSettingsWithOptions creates ModelSettings with control over parallel tool calls
func buildModelSettingsWithOptions(settings services.LLMSettings, allowParallelToolCalls bool) *modelsettings.ModelSettings {
	ms := &modelsettings.ModelSettings{}

	// Always enable usage tracking for streaming (required for Anthropic chat completions)
	ms.IncludeUsage = param.NewOpt(true)

	if settings.Temperature != nil {
		ms.Temperature = param.NewOpt(*settings.Temperature)
	}
	if settings.MaxTokens != nil {
		ms.MaxTokens = param.NewOpt(int64(*settings.MaxTokens))
	}
	if settings.TopP != nil {
		ms.TopP = param.NewOpt(*settings.TopP)
	}
	if settings.FrequencyPenalty != nil {
		ms.FrequencyPenalty = param.NewOpt(*settings.FrequencyPenalty)
	}
	if settings.PresencePenalty != nil {
		ms.PresencePenalty = param.NewOpt(*settings.PresencePenalty)
	}

	// Disable parallel tool calls if specified (prevents model from calling many tools at once)
	if !allowParallelToolCalls {
		ms.ParallelToolCalls = param.NewOpt(false)
	}

	return ms
}

// logNodeConfig logs model and settings for a node
func logNodeConfig(userID int64, nodeName, model string, settings services.LLMSettings) {
	settingsStr := buildSettingsLogString(settings)
	log.Printf("[CONFIG] UserID=%d | %s: %s%s", userID, nodeName, model, settingsStr)
}

func buildSettingsLogString(settings services.LLMSettings) string {
	var parts []string
	if settings.Temperature != nil {
		parts = append(parts, fmt.Sprintf("temp=%.2f", *settings.Temperature))
	}
	if settings.MaxTokens != nil {
		parts = append(parts, fmt.Sprintf("max_tokens=%d", *settings.MaxTokens))
	}
	if settings.TopP != nil {
		parts = append(parts, fmt.Sprintf("top_p=%.2f", *settings.TopP))
	}
	if settings.TopK != nil {
		parts = append(parts, fmt.Sprintf("top_k=%d", *settings.TopK))
	}
	if settings.FrequencyPenalty != nil {
		parts = append(parts, fmt.Sprintf("freq_pen=%.2f", *settings.FrequencyPenalty))
	}
	if settings.PresencePenalty != nil {
		parts = append(parts, fmt.Sprintf("pres_pen=%.2f", *settings.PresencePenalty))
	}
	if len(parts) == 0 {
		return " (defaults)"
	}
	return " " + strings.Join(parts, " ")
}

// buildTriageAgentForUser creates a triage agent with user-specific model configuration
func (o *Orchestrator) buildTriageAgentForUser(userID int64, skipSettings bool) *agents.Agent {
	// If no config cache or userID is 0, use default agent
	if o.configCache == nil || userID == 0 {
		return o.triageAgent
	}

	// Get models for each node from cache
	triageModel := o.configCache.GetModel(userID, "triage_orchestrator")
	jobSearchModel := o.configCache.GetModel(userID, "job_search_worker")
	crmModel := o.configCache.GetModel(userID, "crm_worker")
	generalModel := o.configCache.GetModel(userID, "general_worker")

	// Get settings (empty if skipSettings), filtered for model compatibility
	var triageSettings, jobSearchSettings, crmSettings, generalSettings services.LLMSettings
	if !skipSettings {
		triageSettings = services.FilterSettingsForModel(o.configCache.GetSettings(userID, "triage_orchestrator"), triageModel)
		jobSearchSettings = services.FilterSettingsForModel(o.configCache.GetSettings(userID, "job_search_worker"), jobSearchModel)
		crmSettings = services.FilterSettingsForModel(o.configCache.GetSettings(userID, "crm_worker"), crmModel)
		generalSettings = services.FilterSettingsForModel(o.configCache.GetSettings(userID, "general_worker"), generalModel)
	}

	// Log model configuration for each node
	logNodeConfig(userID, "triage_orchestrator", triageModel, triageSettings)
	logNodeConfig(userID, "job_search_worker", jobSearchModel, jobSearchSettings)
	logNodeConfig(userID, "crm_worker", crmModel, crmSettings)
	logNodeConfig(userID, "general_worker", generalModel, generalSettings)

	// Create workers with user-specific models and settings
	// Parallel tool calls enabled - call limit enforced in SerperTools.JobSearchCtx
	jobSearchWorker := workers.NewJobSearchWorker(jobSearchModel, buildModelSettings(jobSearchSettings), o.allTools)
	crmWorker := workers.NewCRMWorker(crmModel, buildModelSettings(crmSettings), o.allTools)
	generalWorker := workers.NewGeneralWorker(generalModel, buildModelSettings(generalSettings), o.allTools)

	// Create triage agent with user-specific model and settings
	triageAgent := agents.New("triage").
		WithInstructions(prompts.TriageInstructionsWithTools).
		WithModel(triageModel).
		WithHandoffDescription("Routes requests to specialized workers").
		WithAgentHandoffs(jobSearchWorker, crmWorker, generalWorker)

	if ms := buildModelSettings(triageSettings); ms != nil {
		triageAgent = triageAgent.WithModelSettings(*ms)
	}

	return triageAgent
}

// RunRequest represents a request to run the agent
type RunRequest struct {
	SessionID    string
	UserID       string
	TenantID     int64
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
	URLMap           map[string]string // Maps [1], [2] references to full URLs
}

// Event represents an agent event during execution
type Event struct {
	Type    string
	Content string
	Agent   string
}

// StreamEvent represents a real-time event sent during agent execution
type StreamEvent struct {
	Type string `json:"type"` // "start", "text", "tokens", "tool_start", "tool_end", "agent_start", "agent_end", "done", "error", "warning", "eval"

	// Text streaming
	Text string `json:"text,omitempty"`

	// Warning message (non-fatal)
	Warning string `json:"warning,omitempty"`

	// Token usage
	Tokens *TokenUsage `json:"tokens,omitempty"`

	// Timing
	ElapsedMs int64 `json:"elapsed_ms,omitempty"`

	// Tool/Agent info
	ToolName  string `json:"tool_name,omitempty"`
	AgentName string `json:"agent_name,omitempty"`

	// Final response (on "done")
	Response         string            `json:"response,omitempty"`
	TurnTokens       *TokenUsage       `json:"turn_tokens,omitempty"`
	CumulativeTokens *TokenUsage       `json:"cumulative_tokens,omitempty"`
	URLMap           map[string]string `json:"url_map,omitempty"` // Maps [1], [2] references to full URLs

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

	// Seed in-memory state from database history if needed
	o.seedFromDatabaseHistory(ctx, req.SessionID)

	// Get context from state manager
	contextMsgs, err := o.stateManager.GetMessages(ctx, req.SessionID, 8000)
	if err != nil {
		return nil, fmt.Errorf("get messages: %w", err)
	}

	// Build input with context
	input := buildInputWithContext(req.Message, contextMsgs)

	// Create usage context for token tracking
	ctx = usage.NewContext(ctx, usage.NewUsage())

	// Build triage agent with user-specific models
	userID, _ := strconv.ParseInt(req.UserID, 10, 64)
	triageAgent := o.buildTriageAgentForUser(userID, false)

	// Run the triage agent with increased turn limit (SDK handles handoffs automatically)
	runner := agents.Runner{Config: agents.RunConfig{MaxTurns: 20}}
	result, err := runner.Run(ctx, triageAgent, input)
	if err != nil {
		utils.LogError("Agent run error: %v", err)
		return nil, fmt.Errorf("agent run error: %w", err)
	}

	// Extract final output and URL map
	finalOutput := extractFinalOutput(result)
	urlMap := extractURLMapFromResult(result)

	// Get token usage from context
	turnTokens := extractTokenUsage(ctx)

	// Store messages in memory
	if err := o.stateManager.AddMessage(ctx, req.SessionID, state.Message{Role: "user", Content: req.Message}); err != nil {
		log.Printf("Warning: failed to store user message: %v", err)
	}
	if err := o.stateManager.AddMessage(ctx, req.SessionID, state.Message{Role: "assistant", Content: finalOutput}); err != nil {
		log.Printf("Warning: failed to store assistant message: %v", err)
	}

	// Persist messages to database
	o.persistMessages(req, turnTokens, finalOutput)

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
		URLMap:           urlMap,
	}, nil
}

// streamState holds mutable state for streaming event processing
type streamState struct {
	finalOutput      string
	turnTokens       TokenUsage
	currentAgent     string
	bufferedText     string
	useEvaluator     bool
	sendEvent        func(StreamEvent)
	agentStartTime   time.Time
	agentTokens      TokenUsage // tokens for current agent only
	perAgentTokens   map[string]TokenUsage
	urlMap           map[string]string // Maps [1], [2] references to full URLs
}

func (s *streamState) handleAgentUpdated(e agents.AgentUpdatedStreamEvent) {
	if e.NewAgent == nil {
		return
	}
	if e.NewAgent.Name == s.currentAgent {
		return
	}
	s.logAgentTransition(e.NewAgent.Name)
	s.sendEvent(StreamEvent{Type: "agent_start", AgentName: e.NewAgent.Name})
	s.currentAgent = e.NewAgent.Name
}

func (s *streamState) logAgentTransition(newAgent string) {
	if s.currentAgent == "" {
		utils.LogAgentStart(newAgent)
		s.agentStartTime = time.Now()
		s.agentTokens = TokenUsage{}
		if s.perAgentTokens == nil {
			s.perAgentTokens = make(map[string]TokenUsage)
		}
		return
	}
	duration := time.Since(s.agentStartTime)
	// Save tokens for the completing agent
	s.perAgentTokens[s.currentAgent] = s.agentTokens
	utils.LogHandoffWithTokens(s.currentAgent, newAgent, duration, s.agentTokens.Input, s.agentTokens.Output, s.agentTokens.Total)
	// Reset for new agent
	s.agentStartTime = time.Now()
	s.agentTokens = TokenUsage{}
}

func (s *streamState) handleToolCalled(item agents.ToolCallItem) {
	toolName := extractToolName(item)
	utils.LogToolStart(toolName)
	s.sendEvent(StreamEvent{Type: "tool_start", ToolName: toolName})
}

func (s *streamState) handleToolOutput(item agents.ToolCallOutputItem) {
	toolName := extractOutputToolName(item)
	outputStr, ok := item.Output.(string)
	if !ok {
		log.Printf("handleToolOutput: output type=%T (not string)", item.Output)
		utils.LogToolEnd(toolName, "")
		s.sendEvent(StreamEvent{Type: "tool_end", ToolName: toolName})
		return
	}

	log.Printf("handleToolOutput: output type=string, len=%d, contains URLS=%v", len(outputStr), strings.Contains(outputStr, "---URLS---"))
	if strings.Contains(outputStr, "---URLS---") {
		s.captureURLMap(outputStr)
	}
	utils.LogToolEnd(toolName, truncateString(outputStr, 100))
	s.sendEvent(StreamEvent{Type: "tool_end", ToolName: toolName})
}

// captureURLMap extracts URL mappings from ---URLS--- section
func (s *streamState) captureURLMap(output string) {
	urlMap := parseURLMap(output)
	if len(urlMap) == 0 {
		log.Printf("captureURLMap: no URLs parsed from output")
		return
	}
	if s.urlMap == nil {
		s.urlMap = make(map[string]string)
	}
	for k, v := range urlMap {
		s.urlMap[k] = v
	}
	log.Printf("captureURLMap: captured %d URLs", len(s.urlMap))
}

// parseURLMap extracts [n]=url mappings from text containing ---URLS--- section
func parseURLMap(text string) map[string]string {
	const marker = "---URLS---"
	idx := strings.Index(text, marker)
	if idx == -1 {
		return nil
	}

	urlSection := text[idx+len(marker):]
	lines := strings.Split(urlSection, "\n")
	result := make(map[string]string)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Parse [n]=url format
		if !strings.HasPrefix(line, "[") {
			continue
		}
		eqIdx := strings.Index(line, "]=")
		if eqIdx == -1 {
			continue
		}
		ref := line[:eqIdx+1] // includes the ]
		url := line[eqIdx+2:]
		result[ref] = url
	}

	return result
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
		s.tryHandleToolCalled(e.Item)
		return
	}
	if e.Name == agents.StreamEventToolOutput {
		s.tryHandleToolOutput(e.Item)
		return
	}
	if e.Name == agents.StreamEventMessageOutputCreated {
		s.tryHandleMessageOutput(e.Item)
	}
}

func (s *streamState) tryHandleToolCalled(item any) {
	toolCall, ok := item.(agents.ToolCallItem)
	if !ok {
		return
	}
	s.handleToolCalled(toolCall)
}

func (s *streamState) tryHandleToolOutput(item any) {
	toolOutput, ok := item.(agents.ToolCallOutputItem)
	if !ok {
		return
	}
	s.handleToolOutput(toolOutput)
}

func (s *streamState) tryHandleMessageOutput(item any) {
	msgOutput, ok := item.(agents.MessageOutputItem)
	if !ok {
		return
	}
	s.handleMessageOutput(msgOutput)
}

func (s *streamState) handleTextDelta(delta string) {
	if s.useEvaluator {
		s.bufferedText += delta
		return
	}
	s.sendEvent(StreamEvent{Type: "text", Text: delta})
}

func (s *streamState) handleResponseCompleted(resp agents.RawResponsesStreamEvent) {
	inputTokens := int32(resp.Data.Response.Usage.InputTokens)
	outputTokens := int32(resp.Data.Response.Usage.OutputTokens)
	totalTokens := int32(resp.Data.Response.Usage.TotalTokens)

	// Accumulate tokens across all agent calls
	s.turnTokens.Input += inputTokens
	s.turnTokens.Output += outputTokens
	s.turnTokens.Total += totalTokens

	// Track per-agent tokens
	s.agentTokens.Input += inputTokens
	s.agentTokens.Output += outputTokens
	s.agentTokens.Total += totalTokens

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
	if e.Data.Type != "response.completed" {
		return
	}
	// Note: Usage may be zero for chat completions streaming - SDK limitation
	if e.Data.Response.Usage.TotalTokens == 0 {
		return
	}
	s.handleResponseCompleted(e)
}

func (s *streamState) handleStreamEvent(event agents.StreamEvent) {
	s.tryHandleAgentUpdated(event)
	s.tryHandleRunItem(event)
	s.tryHandleRawResponses(event)
}

func (s *streamState) tryHandleAgentUpdated(event agents.StreamEvent) {
	e, ok := event.(agents.AgentUpdatedStreamEvent)
	if !ok {
		return
	}
	s.handleAgentUpdated(e)
}

func (s *streamState) tryHandleRunItem(event agents.StreamEvent) {
	e, ok := event.(agents.RunItemStreamEvent)
	if !ok {
		return
	}
	s.handleRunItemEvent(e)
}

func (s *streamState) tryHandleRawResponses(event agents.StreamEvent) {
	e, ok := event.(agents.RawResponsesStreamEvent)
	if !ok {
		return
	}
	s.handleRawResponse(e)
}

// getProviderForUser returns the appropriate model provider based on user's config
func (o *Orchestrator) getProviderForUser(userID int64) agents.ModelProvider {
	if o.configCache == nil || userID == 0 {
		return o.openaiProvider
	}

	// Check the triage orchestrator's provider setting
	provider := o.configCache.GetProvider(userID, "triage_orchestrator")
	if provider == "anthropic" && o.anthropicProvider != nil {
		log.Printf("Using Anthropic provider for user %d", userID)
		return o.anthropicProvider
	}

	log.Printf("Using OpenAI provider for user %d", userID)
	return o.openaiProvider
}

// runAgentStream runs the agent and returns the stream state and any error
func (o *Orchestrator) runAgentStream(ctx context.Context, userID int64, input string, useEvaluator bool, sendEvent func(StreamEvent), skipSettings bool) (*streamState, error) {
	return o.runAgentStreamDirect(ctx, userID, input, useEvaluator, sendEvent, skipSettings)
}

// runAgentStreamDirect runs agent without model promotion
func (o *Orchestrator) runAgentStreamDirect(ctx context.Context, userID int64, input string, useEvaluator bool, sendEvent func(StreamEvent), skipSettings bool) (*streamState, error) {
	triageAgent := o.buildTriageAgentForUser(userID, skipSettings)
	provider := o.getProviderForUser(userID)

	runner := agents.Runner{Config: agents.RunConfig{MaxTurns: 20, ModelProvider: provider}}
	eventsChan, errChan, err := runner.RunStreamedChan(ctx, triageAgent, input)
	if err != nil {
		return nil, err
	}

	ss := &streamState{
		useEvaluator: useEvaluator,
		sendEvent:    sendEvent,
	}

	for event := range eventsChan {
		ss.handleStreamEvent(event)
	}

	streamErr := <-errChan
	return ss, streamErr
}

// RunWithStreaming executes the agent and streams events via channel
func (o *Orchestrator) RunWithStreaming(ctx context.Context, req RunRequest, eventCh chan<- StreamEvent) {
	startTime := time.Now()
	defer close(eventCh)

	// Reset per-turn call counters and cache state
	o.serperTools.ResetCallCount()
	if o.cacheTools != nil {
		userID, _ := strconv.ParseInt(req.UserID, 10, 64)
		o.cacheTools.SetContext(req.TenantID, userID)
		o.cacheTools.ResetRequestState()
	}

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

	// Seed in-memory state from database history if needed
	o.seedFromDatabaseHistory(ctx, req.SessionID)

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

	userID, _ := strconv.ParseInt(req.UserID, 10, 64)

	// Try running with settings first
	ss, streamErr := o.runAgentStream(ctx, userID, input, req.UseEvaluator, sendEvent, false)

	// Retry without settings if unsupported parameter error
	if streamErr != nil && strings.Contains(streamErr.Error(), "Unsupported parameter") {
		utils.LogInfo("Retrying without LLM settings due to unsupported parameter")
		sendEvent(StreamEvent{
			Type:    "warning",
			Warning: "This model doesn't support some LLM settings (temperature, top_p). Retrying without custom settings.",
		})
		ss, streamErr = o.runAgentStream(ctx, userID, input, req.UseEvaluator, sendEvent, true)
	}

	if streamErr != nil {
		utils.LogError("Streaming error: %v", streamErr)
		sendEvent(StreamEvent{Type: "error", Error: streamErr.Error()})
		return
	}

	finalOutput := ss.finalOutput
	turnTokens := ss.turnTokens
	currentAgent := ss.currentAgent
	bufferedText := ss.bufferedText
	agentStartTime := ss.agentStartTime
	agentTokens := ss.agentTokens
	perAgentTokens := ss.perAgentTokens
	urlMap := ss.urlMap

	// Use buffered text for final output when evaluator is enabled
	if req.UseEvaluator && bufferedText != "" {
		finalOutput = bufferedText
	}

	if currentAgent != "" {
		duration := time.Since(agentStartTime)
		utils.LogAgentEndWithDuration(currentAgent, finalOutput, agentTokens.Input, agentTokens.Output, agentTokens.Total, duration)
		// Add current agent's tokens to perAgentTokens for metrics recording
		if perAgentTokens == nil {
			perAgentTokens = make(map[string]TokenUsage)
		}
		perAgentTokens[currentAgent] = agentTokens
	}

	// Log total token usage across all agents
	log.Printf("📊 TOTAL TOKENS: in=%d out=%d total=%d", turnTokens.Input, turnTokens.Output, turnTokens.Total)

	// Run evaluator if enabled - evaluate first, then send response
	if req.UseEvaluator && finalOutput != "" && o.anthropicAPIKey != "" {
		finalOutput = o.runEvaluation(ctx, req, input, finalOutput, currentAgent, sendEvent)
		// Send the final response after evaluation completes
		sendEvent(StreamEvent{Type: "text", Text: finalOutput})
	}

	// Store messages in memory
	if err := o.stateManager.AddMessage(ctx, req.SessionID, state.Message{Role: "user", Content: req.Message}); err != nil {
		log.Printf("Warning: failed to store user message: %v", err)
	}
	if err := o.stateManager.AddMessage(ctx, req.SessionID, state.Message{Role: "assistant", Content: finalOutput}); err != nil {
		log.Printf("Warning: failed to store assistant message: %v", err)
	}

	// Persist messages to database
	o.persistMessages(req, turnTokens, finalOutput)

	// Update cumulative tokens
	cumulativeTokens := updateCumulativeTokens(req.SessionID, turnTokens)

	log.Printf("Token usage - turn: in=%d out=%d total=%d, cumulative: in=%d out=%d total=%d",
		turnTokens.Input, turnTokens.Output, turnTokens.Total,
		cumulativeTokens.Input, cumulativeTokens.Output, cumulativeTokens.Total)
	log.Println("Agent execution completed")

	// Record metrics for each agent if active session exists
	o.recordAllAgentMetrics(userID, req, startTime, perAgentTokens)

	// Send done event
	sendEvent(StreamEvent{
		Type:             "done",
		Response:         finalOutput,
		TurnTokens:       &turnTokens,
		CumulativeTokens: &cumulativeTokens,
		URLMap:           urlMap,
	})
}

// ensureSession creates a session if it doesn't exist
func (o *Orchestrator) ensureSession(ctx context.Context, userID, sessionID string) error {
	_, err := o.stateManager.GetSession(ctx, userID, sessionID)
	if err == nil {
		return nil
	}

	if !errors.Is(err, apperrors.ErrSessionNotFound) {
		return fmt.Errorf("get session: %w", err)
	}

	_, err = o.stateManager.CreateSession(ctx, userID, sessionID)
	if err != nil {
		return fmt.Errorf("create session: %w", err)
	}
	return nil
}

// seedFromDatabaseHistory loads conversation history from database into in-memory state
func (o *Orchestrator) seedFromDatabaseHistory(ctx context.Context, sessionID string) {
	if o.convService == nil {
		return
	}

	// Check if session already has messages (avoid re-seeding)
	existingMsgs, err := o.stateManager.GetMessages(ctx, sessionID, 1)
	if err == nil && len(existingMsgs) > 0 {
		return
	}

	// Load last 2 messages from database (1 exchange) to minimize context
	dbMessages, err := o.convService.LoadMessages(sessionID, 2)
	if err != nil {
		log.Printf("Warning: failed to load DB history: %v", err)
		return
	}

	if len(dbMessages) == 0 {
		return
	}

	log.Printf("Seeding session with %d messages from database", len(dbMessages))
	for _, msg := range dbMessages {
		if err := o.stateManager.AddMessage(ctx, sessionID, state.Message{
			Role:    msg.Role,
			Content: msg.Content,
		}); err != nil {
			log.Printf("Warning: failed to seed message: %v", err)
		}
	}
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

// extractURLMapFromResult extracts URL mappings from tool outputs in the run result
func extractURLMapFromResult(result *agents.RunResult) map[string]string {
	if result == nil {
		return nil
	}

	urlMap := make(map[string]string)
	for _, item := range result.NewItems {
		toolOutput, ok := item.(agents.ToolCallOutputItem)
		if !ok {
			continue
		}
		outputStr, ok := toolOutput.Output.(string)
		if !ok {
			continue
		}
		for k, v := range parseURLMap(outputStr) {
			urlMap[k] = v
		}
	}

	if len(urlMap) == 0 {
		return nil
	}
	return urlMap
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

// persistMessages saves user and assistant messages to the database
func (o *Orchestrator) persistMessages(req RunRequest, turnTokens TokenUsage, finalOutput string) {
	if o.convService == nil {
		return
	}

	userID, err := strconv.ParseInt(req.UserID, 10, 64)
	if err != nil {
		log.Printf("Warning: invalid user ID for persistence: %v", err)
		return
	}

	// Build token usage JSON
	tokenUsageJSON := buildTokenUsageJSON(turnTokens)

	// Persist user message
	result, err := o.convService.SaveMessage(req.SessionID, userID, req.TenantID, "user", req.Message, nil)
	if err != nil {
		log.Printf("Warning: failed to persist user message: %v", err)
		return
	}

	// Persist assistant message with token usage
	if _, err := o.convService.SaveMessage(req.SessionID, userID, req.TenantID, "assistant", finalOutput, tokenUsageJSON); err != nil {
		log.Printf("Warning: failed to persist assistant message: %v", err)
	}

	// Generate title for new conversations
	if result != nil && result.IsNewConversation {
		o.generateConversationTitle(req.SessionID, userID, req.Message, finalOutput)
	}
}

// generateConversationTitle creates a short title for a conversation using Claude Haiku
func (o *Orchestrator) generateConversationTitle(sessionID string, userID int64, userMessage, assistantResponse string) {
	title := o.callHaikuForTitle(userID, userMessage, assistantResponse)
	if title == "" {
		// Fallback to simple truncation
		title = userMessage
		if len(title) > 50 {
			title = title[:47] + "..."
		}
	}

	if err := o.convService.SetTitle(sessionID, title); err != nil {
		log.Printf("Warning: failed to set conversation title: %v", err)
		return
	}
	log.Printf("Set conversation title: %s", title)
}

// callHaikuForTitle uses Claude Haiku to generate a concise title
func (o *Orchestrator) callHaikuForTitle(userID int64, userMessage, assistantResponse string) string {
	if o.anthropicAPIKey == "" {
		return ""
	}

	// Get title generator model from config cache
	titleModel := "claude-haiku-4-5-20251001" // default
	if o.configCache != nil && userID > 0 {
		titleModel = o.configCache.GetModel(userID, "title_generator")
		log.Printf("[CONFIG] UserID=%d | title_generator: %s", userID, titleModel)
	}

	// Truncate inputs for the prompt
	userSnippet := truncateString(userMessage, 200)
	assistantSnippet := truncateString(assistantResponse, 200)

	prompt := fmt.Sprintf(`Generate a 3-5 word title summarizing this conversation. Return ONLY the title, no quotes or punctuation.

User: %s
Assistant: %s

Title:`, userSnippet, assistantSnippet)

	title, err := callAnthropicAPI(o.anthropicAPIKey, titleModel, prompt)
	if err != nil {
		log.Printf("Warning: failed to generate title with Haiku: %v", err)
		return ""
	}

	// Clean up the title
	title = strings.TrimSpace(title)
	title = strings.Trim(title, `"'`)
	if len(title) > 100 {
		title = title[:97] + "..."
	}

	return title
}

// callAnthropicAPI makes a request to Claude for title generation
func callAnthropicAPI(apiKey, model, prompt string) (string, error) {
	reqBody := map[string]interface{}{
		"model":      model,
		"max_tokens": 30,
		"system":     "You generate short, descriptive titles. Return only the title text.",
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("anthropic API returned status %d", resp.StatusCode)
	}

	var result struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if len(result.Content) == 0 {
		return "", fmt.Errorf("no content in response")
	}

	return result.Content[0].Text, nil
}

// buildTokenUsageJSON creates a JSON representation of token usage
func buildTokenUsageJSON(tokens TokenUsage) datatypes.JSON {
	data := map[string]int32{
		"input":  tokens.Input,
		"output": tokens.Output,
		"total":  tokens.Total,
	}
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return nil
	}
	return datatypes.JSON(jsonBytes)
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

	// Get evaluator model and settings from config cache
	userID, _ := strconv.ParseInt(req.UserID, 10, 64)
	evaluatorModel := ""
	var evaluatorSettings services.LLMSettings
	if o.configCache != nil && userID > 0 {
		evaluatorModel = o.configCache.GetModel(userID, "evaluator")
		evaluatorSettings = services.FilterSettingsForModel(o.configCache.GetSettings(userID, "evaluator"), evaluatorModel)
		logNodeConfig(userID, "evaluator", evaluatorModel, evaluatorSettings)
	}
	evaluator := NewEvaluator(o.anthropicAPIKey, evalType, evaluatorModel, evaluatorSettings)

	sendEvent(StreamEvent{Type: "eval_start"})

	evalStartTime := time.Now()
	evalResult, err := evaluator.Evaluate(ctx, req.Message, finalOutput, "")
	if err != nil {
		log.Printf("Evaluator error: %v", err)
		return finalOutput
	}

	// Record evaluator metrics
	o.recordEvaluatorMetric(userID, req, evalStartTime, evalResult, evaluatorModel)

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

	// Build user-specific agent for retry
	triageAgent := o.buildTriageAgentForUser(userID, false)

	retryInput := fmt.Sprintf("%s\n\nPrevious attempt feedback: %s\nPlease improve the response.", input, evalResult.Feedback)
	retryResult, retryErr := agents.Run(ctx, triageAgent, retryInput)
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
	retryEvalStartTime := time.Now()
	evalResult, _ = evaluator.Evaluate(ctx, req.Message, retryOutput, "")
	o.recordEvaluatorMetric(userID, req, retryEvalStartTime, evalResult, evaluatorModel)
	sendEvent(StreamEvent{Type: "eval", EvalResult: evalResult})
	return retryOutput
}

// agentNameToConfigKey maps SDK agent names to config cache keys
func agentNameToConfigKey(agentName string) string {
	mapping := map[string]string{
		"triage":      "triage_orchestrator",
		"job_search":  "job_search_worker",
		"crm":         "crm_worker",
		"general":     "general_worker",
		"cover_letter": "cover_letter_worker",
	}
	if key, ok := mapping[agentName]; ok {
		return key
	}
	return agentName
}

// getAgentModel returns the model for a given agent from cache, or default model
func (o *Orchestrator) getAgentModel(userID int64, agentName string) string {
	if o.configCache == nil || userID == 0 {
		return o.model
	}
	configKey := agentNameToConfigKey(agentName)
	return o.configCache.GetModel(userID, configKey)
}

// recordNodeMetric records a single metric for any node (agent, worker, evaluator)
func (o *Orchestrator) recordNodeMetric(userID int64, req RunRequest, startTime time.Time, nodeName, model string, inputTokens, outputTokens int32) {
	if o.metricService == nil {
		return
	}

	activeSession, err := o.metricService.GetActiveSession(userID)
	if err != nil {
		log.Printf("Warning: failed to check active recording session: %v", err)
		return
	}
	if activeSession == nil {
		return
	}

	endTime := time.Now()
	durationMs := int(endTime.Sub(startTime).Milliseconds())
	configSnapshot := o.buildMetricConfigSnapshot(userID)

	provider := "openai"
	if strings.HasPrefix(model, "claude") {
		provider = "anthropic"
	}

	metricInput := services.TurnMetricInput{
		SessionID:      activeSession.ID,
		ThreadID:       req.SessionID,
		StartedAt:      startTime,
		EndedAt:        endTime,
		DurationMs:     durationMs,
		InputTokens:    inputTokens,
		OutputTokens:   outputTokens,
		Model:          model,
		Provider:       provider,
		NodeName:       nodeName,
		ConfigSnapshot: configSnapshot,
		UserMessage:    req.Message,
	}

	if err := o.metricService.RecordTurn(metricInput); err != nil {
		log.Printf("Warning: failed to record metric for %s: %v", nodeName, err)
		return
	}

	log.Printf("📊 Recorded metric for session %d, node %s: %d tokens", activeSession.ID, nodeName, inputTokens+outputTokens)
}

// recordAllAgentMetrics records metrics for each agent that ran during the turn
func (o *Orchestrator) recordAllAgentMetrics(userID int64, req RunRequest, startTime time.Time, perAgentTokens map[string]TokenUsage) {
	if len(perAgentTokens) == 0 {
		return
	}

	for agentName, tokens := range perAgentTokens {
		agentModel := o.getAgentModel(userID, agentName)
		o.recordNodeMetric(userID, req, startTime, agentName, agentModel, tokens.Input, tokens.Output)
	}
}

// buildMetricConfigSnapshot creates a JSON snapshot of current agent configuration
func (o *Orchestrator) buildMetricConfigSnapshot(userID int64) json.RawMessage {
	if o.configCache == nil || userID == 0 {
		return nil
	}

	snapshot := map[string]interface{}{
		"triage": map[string]interface{}{
			"model":    o.configCache.GetModel(userID, "triage_orchestrator"),
			"settings": o.configCache.GetSettings(userID, "triage_orchestrator"),
		},
		"job_search": map[string]interface{}{
			"model":    o.configCache.GetModel(userID, "job_search_worker"),
			"settings": o.configCache.GetSettings(userID, "job_search_worker"),
		},
		"crm_worker": map[string]interface{}{
			"model":    o.configCache.GetModel(userID, "crm_worker"),
			"settings": o.configCache.GetSettings(userID, "crm_worker"),
		},
		"general_worker": map[string]interface{}{
			"model":    o.configCache.GetModel(userID, "general_worker"),
			"settings": o.configCache.GetSettings(userID, "general_worker"),
		},
		"evaluator": map[string]interface{}{
			"model":    o.configCache.GetModel(userID, "evaluator"),
			"settings": o.configCache.GetSettings(userID, "evaluator"),
		},
	}

	data, err := json.Marshal(snapshot)
	if err != nil {
		return nil
	}

	return data
}

// recordEvaluatorMetric records metrics for evaluator calls
func (o *Orchestrator) recordEvaluatorMetric(userID int64, req RunRequest, startTime time.Time, evalResult *EvaluatorResult, evaluatorModel string) {
	if evalResult == nil {
		return
	}
	o.recordNodeMetric(userID, req, startTime, "evaluator", evaluatorModel, int32(evalResult.InputTokens), int32(evalResult.OutputTokens))
}
