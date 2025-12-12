package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/pina-colada-co/agent-go/internal/agent/workers"
	"github.com/pina-colada-co/agent-go/internal/config"
	openaimodel "github.com/pina-colada-co/agent-go/internal/model/openai"
	"github.com/pina-colada-co/agent-go/internal/services"
	"github.com/pina-colada-co/agent-go/internal/tools"
	adkagent "google.golang.org/adk/agent"
	"google.golang.org/adk/model"
	"google.golang.org/adk/model/gemini"
	"google.golang.org/adk/runner"
	"google.golang.org/adk/session"
	"google.golang.org/genai"
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

// WorkerType identifies which specialized worker to use
type WorkerType string

const (
	WorkerJobSearch WorkerType = "job_search"
	WorkerCRM       WorkerType = "crm"
	WorkerGeneral   WorkerType = "general"
)

// Orchestrator coordinates the agent system
type Orchestrator struct {
	// Triage model for routing decisions
	triageModel model.LLM

	// Worker runners
	jobSearchRunner *runner.Runner
	crmRunner       *runner.Runner
	generalRunner   *runner.Runner

	sessionSvc session.Service
	appName    string
}

// TriageDecision represents the triage agent's routing decision
type TriageDecision struct {
	Worker string `json:"worker"` // "job_search", "crm", or "general"
	Reason string `json:"reason"` // Brief explanation
}

// NewOrchestrator creates the agent orchestrator with triage-based routing
func NewOrchestrator(ctx context.Context, cfg *config.Config, indService *services.IndividualService, orgService *services.OrganizationService, docService *services.DocumentService) (*Orchestrator, error) {
	// Initialize Gemini model for triage and workers
	geminiModel, err := gemini.NewModel(ctx, cfg.GeminiModel, &genai.ClientConfig{
		APIKey: cfg.GeminiAPIKey,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini model: %w", err)
	}

	// Create CRM tools
	crmTools := tools.NewCRMTools(indService, orgService)
	crmToolList, err := crmTools.BuildTools()
	if err != nil {
		return nil, fmt.Errorf("failed to build CRM tools: %w", err)
	}

	// Create document tools
	docTools := tools.NewDocumentTools(docService)
	docToolList, err := docTools.BuildTools()
	if err != nil {
		return nil, fmt.Errorf("failed to build document tools: %w", err)
	}

	// Combine CRM and document tools
	allCRMTools := append(crmToolList, docToolList...)

	// Create OpenAI model for job search worker
	openaiModel, err := openaimodel.NewModel(cfg.OpenAIAPIKey, "gpt-4.1-2025-04-14")
	if err != nil {
		return nil, fmt.Errorf("failed to create OpenAI model: %w", err)
	}

	// Create workers
	jobSearchWorker, err := workers.NewJobSearchWorker(openaiModel, cfg.SerperAPIKey, allCRMTools)
	if err != nil {
		return nil, fmt.Errorf("failed to create job search worker: %w", err)
	}

	crmWorker, err := workers.NewCRMWorker(geminiModel, allCRMTools)
	if err != nil {
		return nil, fmt.Errorf("failed to create CRM worker: %w", err)
	}

	generalWorker, err := workers.NewGeneralWorker(geminiModel, allCRMTools)
	if err != nil {
		return nil, fmt.Errorf("failed to create general worker: %w", err)
	}

	// Shared session service for all workers
	sessionSvc := session.InMemoryService()
	appName := "pina-colada-agent"

	// Create runners for each worker
	jobSearchRunner, err := runner.New(runner.Config{
		AppName:        appName,
		Agent:          jobSearchWorker,
		SessionService: sessionSvc,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create job search runner: %w", err)
	}

	crmRunner, err := runner.New(runner.Config{
		AppName:        appName,
		Agent:          crmWorker,
		SessionService: sessionSvc,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create CRM runner: %w", err)
	}

	generalRunner, err := runner.New(runner.Config{
		AppName:        appName,
		Agent:          generalWorker,
		SessionService: sessionSvc,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create general runner: %w", err)
	}

	log.Printf("ADK orchestrator initialized with triage routing")

	return &Orchestrator{
		triageModel:     geminiModel,
		jobSearchRunner: jobSearchRunner,
		crmRunner:       crmRunner,
		generalRunner:   generalRunner,
		sessionSvc:      sessionSvc,
		appName:         appName,
	}, nil
}

// triage uses the LLM to decide which worker should handle the request
func (o *Orchestrator) triage(ctx context.Context, message string) (WorkerType, error) {
	prompt := fmt.Sprintf(`You are a routing agent. Decide which worker should handle this request.

WORKERS:
- job_search: For job searches, career opportunities, finding positions, employment. This worker ALSO has CRM access, so use it when the request involves BOTH CRM lookups AND job searching.
- crm: For CRM-ONLY tasks (contacts, individuals, organizations, accounts, documents) with NO job searching involved.
- general: For general questions, conversation, anything else.

IMPORTANT: If the request mentions job search, careers, or employment as the END GOAL, route to job_search even if CRM lookups are needed first. The job_search worker can do both.

USER REQUEST: %s

Respond with JSON only: {"worker": "job_search|crm|general", "reason": "brief explanation"}`, message)

	req := &model.LLMRequest{
		Contents: []*genai.Content{
			{Role: "user", Parts: []*genai.Part{{Text: prompt}}},
		},
	}

	var decision TriageDecision
	for resp, err := range o.triageModel.GenerateContent(ctx, req, false) {
		if err != nil {
			log.Printf("Triage error: %v, defaulting to general", err)
			return WorkerGeneral, nil
		}
		if resp.Content != nil {
			for _, part := range resp.Content.Parts {
				if part.Text != "" {
					// Try to parse JSON from response
					text := part.Text
					// Find JSON in response
					start := -1
					end := -1
					for i, c := range text {
						if c == '{' && start == -1 {
							start = i
						}
						if c == '}' {
							end = i + 1
						}
					}
					if start >= 0 && end > start {
						if err := json.Unmarshal([]byte(text[start:end]), &decision); err == nil {
							log.Printf("🔀 TRIAGE: %s → %s", decision.Reason, decision.Worker)
							switch decision.Worker {
							case "job_search":
								return WorkerJobSearch, nil
							case "crm":
								return WorkerCRM, nil
							default:
								return WorkerGeneral, nil
							}
						}
					}
				}
			}
		}
	}

	return WorkerGeneral, nil
}

// getRunner returns the runner for a worker type
func (o *Orchestrator) getRunner(workerType WorkerType) *runner.Runner {
	switch workerType {
	case WorkerJobSearch:
		return o.jobSearchRunner
	case WorkerCRM:
		return o.crmRunner
	default:
		return o.generalRunner
	}
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

// Run executes the agent with the given message
func (o *Orchestrator) Run(ctx context.Context, req RunRequest) (*RunResponse, error) {
	log.Printf("Starting agent for thread: %s", req.SessionID)
	log.Printf("   Message: %s", req.Message)

	// Get or create session
	sessionID, err := o.getOrCreateSession(ctx, req.UserID, req.SessionID)
	if err != nil {
		return nil, err
	}

	// Triage: LLM decides which worker
	workerType, err := o.triage(ctx, req.Message)
	if err != nil {
		return nil, fmt.Errorf("triage failed: %w", err)
	}
	selectedRunner := o.getRunner(workerType)
	log.Printf("Selected worker: %s", workerType)

	// Run the worker
	userMsg := &genai.Content{
		Role:  "user",
		Parts: []*genai.Part{{Text: req.Message}},
	}

	events := selectedRunner.Run(ctx, req.UserID, sessionID, userMsg, adkagent.RunConfig{})

	var collectedEvents []Event
	var finalResponse string
	var turnTokens TokenUsage
	var lastAuthor string

	for event, err := range events {
		if err != nil {
			LogError("Agent event error: %v", err)
			return nil, fmt.Errorf("agent event error: %w", err)
		}

		if event.Author != "" && event.Author != lastAuthor {
			if lastAuthor != "" {
				LogHandoff(lastAuthor, event.Author)
			} else {
				LogAgentStart(event.Author)
			}
			lastAuthor = event.Author
		}

		if event.Content != nil {
			for _, part := range event.Content.Parts {
				if part.FunctionCall != nil {
					LogToolStart(part.FunctionCall.Name)
				}
				if part.FunctionResponse != nil {
					resultPreview := ""
					if part.FunctionResponse.Response != nil {
						if textVal, ok := part.FunctionResponse.Response["text"].(string); ok {
							resultPreview = textVal
						} else if resultVal, ok := part.FunctionResponse.Response["result"].(string); ok {
							resultPreview = resultVal
						}
					}
					LogToolEnd(part.FunctionResponse.Name, resultPreview)
				}
			}
		}

		if event.UsageMetadata != nil {
			turnTokens.Input += event.UsageMetadata.PromptTokenCount
			turnTokens.Output += event.UsageMetadata.CandidatesTokenCount
			turnTokens.Total += event.UsageMetadata.TotalTokenCount
		}

		if event.Content != nil {
			for _, part := range event.Content.Parts {
				if part.Text != "" {
					collectedEvents = append(collectedEvents, Event{
						Type:    "text",
						Content: part.Text,
						Agent:   event.Author,
					})
					finalResponse = part.Text
				}
			}
		}
	}

	if lastAuthor != "" {
		LogAgentEnd(lastAuthor, finalResponse, &turnTokens)
	}

	// Update cumulative tokens
	sessionTokenMu.Lock()
	if sessionTokenTotals[sessionID] == nil {
		sessionTokenTotals[sessionID] = &TokenUsage{}
	}
	sessionTokenTotals[sessionID].Input += turnTokens.Input
	sessionTokenTotals[sessionID].Output += turnTokens.Output
	sessionTokenTotals[sessionID].Total += turnTokens.Total
	cumulativeTokens := *sessionTokenTotals[sessionID]
	sessionTokenMu.Unlock()

	log.Printf("Token usage - turn: in=%d out=%d total=%d, cumulative: in=%d out=%d total=%d",
		turnTokens.Input, turnTokens.Output, turnTokens.Total,
		cumulativeTokens.Input, cumulativeTokens.Output, cumulativeTokens.Total)
	log.Println("Agent execution completed")

	return &RunResponse{
		Response:         finalResponse,
		SessionID:        sessionID,
		Events:           collectedEvents,
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

	// Get or create session
	sessionID, err := o.getOrCreateSession(ctx, req.UserID, req.SessionID)
	if err != nil {
		sendEvent(StreamEvent{Type: "error", Error: err.Error()})
		return
	}

	// Triage: LLM decides which worker
	workerType, err := o.triage(ctx, req.Message)
	if err != nil {
		sendEvent(StreamEvent{Type: "error", Error: fmt.Sprintf("triage failed: %v", err)})
		return
	}
	selectedRunner := o.getRunner(workerType)
	log.Printf("Selected worker: %s", workerType)

	// Run the worker
	userMsg := &genai.Content{
		Role:  "user",
		Parts: []*genai.Part{{Text: req.Message}},
	}

	events := selectedRunner.Run(ctx, req.UserID, sessionID, userMsg, adkagent.RunConfig{})

	var finalResponse string
	var turnTokens TokenUsage
	var lastAuthor string

	for event, err := range events {
		if err != nil {
			LogError("Agent event error: %v", err)
			sendEvent(StreamEvent{Type: "error", Error: err.Error()})
			return
		}

		if event.Author != "" && event.Author != lastAuthor {
			if lastAuthor != "" {
				LogHandoff(lastAuthor, event.Author)
			} else {
				LogAgentStart(event.Author)
			}
			sendEvent(StreamEvent{Type: "agent_start", AgentName: event.Author})
			lastAuthor = event.Author
		}

		if event.Content != nil {
			for _, part := range event.Content.Parts {
				if part.FunctionCall != nil {
					LogToolStart(part.FunctionCall.Name)
					sendEvent(StreamEvent{Type: "tool_start", ToolName: part.FunctionCall.Name})
				}
				if part.FunctionResponse != nil {
					resultPreview := ""
					if part.FunctionResponse.Response != nil {
						if textVal, ok := part.FunctionResponse.Response["text"].(string); ok {
							resultPreview = textVal
						} else if resultVal, ok := part.FunctionResponse.Response["result"].(string); ok {
							resultPreview = resultVal
						}
					}
					LogToolEnd(part.FunctionResponse.Name, resultPreview)
					sendEvent(StreamEvent{Type: "tool_end", ToolName: part.FunctionResponse.Name})
				}
			}
		}

		if event.UsageMetadata != nil {
			turnTokens.Input += event.UsageMetadata.PromptTokenCount
			turnTokens.Output += event.UsageMetadata.CandidatesTokenCount
			turnTokens.Total += event.UsageMetadata.TotalTokenCount
			sendEvent(StreamEvent{
				Type:   "tokens",
				Tokens: &TokenUsage{Input: turnTokens.Input, Output: turnTokens.Output, Total: turnTokens.Total},
			})
		}

		if event.Content != nil {
			for _, part := range event.Content.Parts {
				if part.Text != "" {
					finalResponse = part.Text
				}
			}
		}
	}

	if lastAuthor != "" {
		LogAgentEnd(lastAuthor, finalResponse, &turnTokens)
	}

	// Update cumulative tokens
	sessionTokenMu.Lock()
	if sessionTokenTotals[sessionID] == nil {
		sessionTokenTotals[sessionID] = &TokenUsage{}
	}
	sessionTokenTotals[sessionID].Input += turnTokens.Input
	sessionTokenTotals[sessionID].Output += turnTokens.Output
	sessionTokenTotals[sessionID].Total += turnTokens.Total
	cumulativeTokens := *sessionTokenTotals[sessionID]
	sessionTokenMu.Unlock()

	log.Printf("Token usage - turn: in=%d out=%d total=%d, cumulative: in=%d out=%d total=%d",
		turnTokens.Input, turnTokens.Output, turnTokens.Total,
		cumulativeTokens.Input, cumulativeTokens.Output, cumulativeTokens.Total)
	log.Println("Agent execution completed")

	// Stream final response
	sendEvent(StreamEvent{Type: "text", Text: finalResponse})

	// Send done event
	sendEvent(StreamEvent{
		Type:             "done",
		Response:         finalResponse,
		TurnTokens:       &turnTokens,
		CumulativeTokens: &cumulativeTokens,
	})
}

// getOrCreateSession gets existing session or creates new one
func (o *Orchestrator) getOrCreateSession(ctx context.Context, userID, sessionID string) (string, error) {
	sessResp, err := o.sessionSvc.Get(ctx, &session.GetRequest{
		AppName:   o.appName,
		UserID:    userID,
		SessionID: sessionID,
	})

	if err != nil || sessResp == nil || sessResp.Session == nil {
		createResp, err := o.sessionSvc.Create(ctx, &session.CreateRequest{
			AppName:   o.appName,
			UserID:    userID,
			SessionID: sessionID,
		})
		if err != nil {
			return "", fmt.Errorf("failed to create session: %w", err)
		}
		return createResp.Session.ID(), nil
	}
	return sessResp.Session.ID(), nil
}
