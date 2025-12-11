package agent

import (
	"context"
	"fmt"

	"github.com/pina-colada-co/agent-go/internal/agent/workers"
	"github.com/pina-colada-co/agent-go/internal/config"
	"github.com/pina-colada-co/agent-go/internal/services"
	"github.com/pina-colada-co/agent-go/internal/tools"
	adkagent "google.golang.org/adk/agent"
	"google.golang.org/adk/model/gemini"
	"google.golang.org/adk/runner"
	"google.golang.org/adk/session"
	"google.golang.org/genai"
)

// Orchestrator coordinates the agent system
type Orchestrator struct {
	triageAgent adkagent.Agent
	runner      *runner.Runner
	sessionSvc  session.Service
	appName     string
}

// NewOrchestrator creates and wires up the agent orchestrator
func NewOrchestrator(ctx context.Context, cfg *config.Config, indService *services.IndividualService, orgService *services.OrganizationService) (*Orchestrator, error) {
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

	// Create workers
	crmWorker, err := workers.NewCRMWorker(model, crmToolList)
	if err != nil {
		return nil, fmt.Errorf("failed to create CRM worker: %w", err)
	}

	generalWorker, err := workers.NewGeneralWorker(model)
	if err != nil {
		return nil, fmt.Errorf("failed to create general worker: %w", err)
	}

	// Create job search worker with Google Search + CRM tools
	jobSearchWorker, err := workers.NewJobSearchWorker(model, crmToolList)
	if err != nil {
		return nil, fmt.Errorf("failed to create job search worker: %w", err)
	}

	// Create triage agent with workers as sub-agents
	workerAgents := []adkagent.Agent{jobSearchWorker, crmWorker, generalWorker}
	triageAgent, err := NewTriageAgent(model, workerAgents)
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
	Response  string
	SessionID string
	Events    []Event
}

// Event represents an agent event during execution
type Event struct {
	Type    string
	Content string
	Agent   string
}

// Run executes the agent with the given message
func (o *Orchestrator) Run(ctx context.Context, req RunRequest) (*RunResponse, error) {
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

	for event, err := range events {
		if err != nil {
			return nil, fmt.Errorf("agent event error: %w", err)
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

	return &RunResponse{
		Response:  finalResponse,
		SessionID: sessionID,
		Events:    collectedEvents,
	}, nil
}
