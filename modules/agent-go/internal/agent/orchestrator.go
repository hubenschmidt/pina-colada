package agent

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/pina-colada-co/agent-go/internal/agent/workers"
	"github.com/pina-colada-co/agent-go/internal/config"
	"github.com/pina-colada-co/agent-go/internal/services"
	"github.com/pina-colada-co/agent-go/internal/tools"
	adkagent "google.golang.org/adk/agent"
	"google.golang.org/adk/model/gemini"
	"google.golang.org/adk/runner"
	"google.golang.org/adk/session"
	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/agenttool"
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

// Orchestrator coordinates the agent system
type Orchestrator struct {
	triageAgent adkagent.Agent
	runner      *runner.Runner
	sessionSvc  session.Service
	appName     string
}

// NewOrchestrator creates and wires up the agent orchestrator
func NewOrchestrator(ctx context.Context, cfg *config.Config, indService *services.IndividualService, orgService *services.OrganizationService, docService *services.DocumentService) (*Orchestrator, error) {
	// Initialize Gemini model
	model, err := gemini.NewModel(ctx, cfg.GeminiModel, &genai.ClientConfig{
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

	// Combine CRM and document tools for CRM worker
	allCRMTools := append(crmToolList, docToolList...)

	// Create workers - all workers get CRM tools since this is a CRM-focused app
	crmWorker, err := workers.NewCRMWorker(model, allCRMTools)
	if err != nil {
		return nil, fmt.Errorf("failed to create CRM worker: %w", err)
	}

	generalWorker, err := workers.NewGeneralWorker(model, allCRMTools)
	if err != nil {
		return nil, fmt.Errorf("failed to create general worker: %w", err)
	}

	// Create job search worker with Serper API + CRM tools for resume matching
	jobSearchWorker, err := workers.NewJobSearchWorker(model, cfg.SerperAPIKey, allCRMTools)
	if err != nil {
		return nil, fmt.Errorf("failed to create job search worker: %w", err)
	}

	// Wrap specialized agents as agenttool for clean separation of concerns
	// Each worker handles a specific domain (job search, CRM, general)
	agentTools := []tool.Tool{
		agenttool.New(jobSearchWorker, nil), // Serper-based job search
		agenttool.New(crmWorker, nil),       // CRM tools
		agenttool.New(generalWorker, nil),   // General assistance
	}

	// Create triage agent with worker agents wrapped as tools
	triageAgent, err := NewTriageAgentWithTools(model, agentTools)
	if err != nil {
		return nil, fmt.Errorf("failed to create triage agent: %w", err)
	}

	// Create in-memory session service
	sessionSvc := session.InMemoryService()

	appName := "pina-colada-agent"

	// Create runner
	r, err := runner.New(runner.Config{
		AppName:        appName,
		Agent:          triageAgent,
		SessionService: sessionSvc,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create runner: %w", err)
	}

	return &Orchestrator{
		triageAgent: triageAgent,
		runner:      r,
		sessionSvc:  sessionSvc,
		appName:     appName,
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

	// Token usage (cumulative during turn)
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
	sessResp, err := o.sessionSvc.Get(ctx, &session.GetRequest{
		AppName:   o.appName,
		UserID:    req.UserID,
		SessionID: req.SessionID,
	})

	sessionID := req.SessionID
	if err != nil || sessResp == nil || sessResp.Session == nil {
		// Create new session
		createResp, err := o.sessionSvc.Create(ctx, &session.CreateRequest{
			AppName:   o.appName,
			UserID:    req.UserID,
			SessionID: req.SessionID,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create session: %w", err)
		}
		sessionID = createResp.Session.ID()
	} else {
		sessionID = sessResp.Session.ID()
	}

	// Build user message content
	userMsg := &genai.Content{
		Role: "user",
		Parts: []*genai.Part{
			{Text: req.Message},
		},
	}

	// Run the agent - returns an iterator
	events := o.runner.Run(ctx, req.UserID, sessionID, userMsg, adkagent.RunConfig{})

	// Collect events and final response
	var collectedEvents []Event
	var finalResponse string
	var turnTokens TokenUsage
	var lastAuthor string

	for event, err := range events {
		if err != nil {
			LogError("Agent event error: %v", err)
			return nil, fmt.Errorf("agent event error: %w", err)
		}

		// Log agent transitions (handoffs)
		if event.Author != "" && event.Author != lastAuthor {
			if lastAuthor != "" {
				LogHandoff(lastAuthor, event.Author)
			} else {
				LogAgentStart(event.Author)
			}
			lastAuthor = event.Author
		}

		// Log function calls (tool invocations)
		if event.Content != nil {
			for _, part := range event.Content.Parts {
				if part.FunctionCall != nil {
					LogToolStart(part.FunctionCall.Name)
				}
				if part.FunctionResponse != nil {
					// Extract result preview from function response
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

		// Accumulate token usage from events
		if event.UsageMetadata != nil {
			turnTokens.Input += event.UsageMetadata.PromptTokenCount
			turnTokens.Output += event.UsageMetadata.CandidatesTokenCount
			turnTokens.Total += event.UsageMetadata.TotalTokenCount
		}

		// Extract text content from event
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

	// Log final agent end
	if lastAuthor != "" {
		LogAgentEnd(lastAuthor, finalResponse, &turnTokens)
	}

	// Update cumulative tokens for this session
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

// RunWithStreaming executes the agent and streams events in real-time via a channel
func (o *Orchestrator) RunWithStreaming(ctx context.Context, req RunRequest, eventCh chan<- StreamEvent) {
	startTime := time.Now()
	defer close(eventCh)

	log.Printf("Starting streaming agent for thread: %s", req.SessionID)
	log.Printf("   Message: %s", req.Message)

	// Helper to send event with elapsed time
	sendEvent := func(evt StreamEvent) {
		evt.ElapsedMs = time.Since(startTime).Milliseconds()
		select {
		case eventCh <- evt:
		case <-ctx.Done():
		}
	}

	// Get or create session
	sessResp, err := o.sessionSvc.Get(ctx, &session.GetRequest{
		AppName:   o.appName,
		UserID:    req.UserID,
		SessionID: req.SessionID,
	})

	sessionID := req.SessionID
	if err != nil || sessResp == nil || sessResp.Session == nil {
		createResp, err := o.sessionSvc.Create(ctx, &session.CreateRequest{
			AppName:   o.appName,
			UserID:    req.UserID,
			SessionID: req.SessionID,
		})
		if err != nil {
			sendEvent(StreamEvent{Type: "error", Error: fmt.Sprintf("failed to create session: %v", err)})
			return
		}
		sessionID = createResp.Session.ID()
	} else {
		sessionID = sessResp.Session.ID()
	}

	// Build user message content
	userMsg := &genai.Content{
		Role: "user",
		Parts: []*genai.Part{
			{Text: req.Message},
		},
	}

	// Run the agent - returns an iterator
	events := o.runner.Run(ctx, req.UserID, sessionID, userMsg, adkagent.RunConfig{})

	var finalResponse string
	var turnTokens TokenUsage
	var lastAuthor string

	for event, err := range events {
		if err != nil {
			LogError("Agent event error: %v", err)
			sendEvent(StreamEvent{Type: "error", Error: err.Error()})
			return
		}

		// Handle agent transitions
		if event.Author != "" && event.Author != lastAuthor {
			if lastAuthor != "" {
				LogHandoff(lastAuthor, event.Author)
			} else {
				LogAgentStart(event.Author)
			}
			sendEvent(StreamEvent{Type: "agent_start", AgentName: event.Author})
			lastAuthor = event.Author
		}

		// Handle function calls and responses
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

		// Accumulate and stream token usage
		if event.UsageMetadata != nil {
			turnTokens.Input += event.UsageMetadata.PromptTokenCount
			turnTokens.Output += event.UsageMetadata.CandidatesTokenCount
			turnTokens.Total += event.UsageMetadata.TotalTokenCount

			// Send token update
			sendEvent(StreamEvent{
				Type:   "tokens",
				Tokens: &TokenUsage{Input: turnTokens.Input, Output: turnTokens.Output, Total: turnTokens.Total},
			})
		}

		// Stream text content
		if event.Content != nil {
			for _, part := range event.Content.Parts {
				if part.Text != "" {
					finalResponse = part.Text
					sendEvent(StreamEvent{Type: "text", Text: part.Text})
				}
			}
		}
	}

	// Log final agent end
	if lastAuthor != "" {
		LogAgentEnd(lastAuthor, finalResponse, &turnTokens)
	}

	// Update cumulative tokens for this session
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

	// Send final done event
	sendEvent(StreamEvent{
		Type:             "done",
		Response:         finalResponse,
		TurnTokens:       &turnTokens,
		CumulativeTokens: &cumulativeTokens,
	})
}
