package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/pina-colada-co/agent-go/internal/agent"
	"github.com/pina-colada-co/agent-go/internal/middleware"
)

// AgentController handles agent HTTP requests
type AgentController struct {
	orchestrator *agent.Orchestrator
}

// NewAgentController creates a new agent controller
func NewAgentController(orchestrator *agent.Orchestrator) *AgentController {
	return &AgentController{orchestrator: orchestrator}
}

// ChatRequest represents the request body for chat endpoint
type ChatRequest struct {
	Message   string `json:"message"`
	ThreadID  string `json:"thread_id,omitempty"`
	SessionID string `json:"session_id,omitempty"` // Alias for thread_id
}

// ChatResponse represents the response from chat endpoint
type ChatResponse struct {
	Response  string        `json:"response"`
	ThreadID  string        `json:"thread_id"`
	SessionID string        `json:"session_id"`
	Events    []agent.Event `json:"events,omitempty"`
}

// Chat handles POST /agent/chat
func (c *AgentController) Chat(w http.ResponseWriter, r *http.Request) {
	if c.orchestrator == nil {
		writeError(w, http.StatusServiceUnavailable, "agent not configured - GEMINI_API_KEY required")
		return
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	tenantID, _ := middleware.GetTenantID(r.Context())

	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Message == "" {
		writeError(w, http.StatusBadRequest, "message is required")
		return
	}

	// Use thread_id or session_id (both mean the same thing)
	sessionID := req.ThreadID
	if sessionID == "" {
		sessionID = req.SessionID
	}
	if sessionID == "" {
		sessionID = uuid.New().String()
	}

	// Run the agent
	result, err := c.orchestrator.Run(r.Context(), agent.RunRequest{
		SessionID: sessionID,
		UserID:    strconv.FormatInt(userID, 10),
		TenantID:  tenantID,
		Message:   req.Message,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("agent error: %v", err))
		return
	}

	writeJSON(w, http.StatusOK, ChatResponse{
		Response:  result.Response,
		ThreadID:  result.SessionID,
		SessionID: result.SessionID,
		Events:    result.Events,
	})
}

// HealthCheck handles GET /agent/health
func (c *AgentController) HealthCheck(w http.ResponseWriter, r *http.Request) {
	status := "ok"
	if c.orchestrator == nil {
		status = "not_configured"
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"status": status,
		"agent":  "adk-go",
	})
}

// InitOrchestrator is a helper to lazily initialize the orchestrator
// This is useful when you want to delay initialization until after config is loaded
type OrchestratorFactory func(ctx context.Context) (*agent.Orchestrator, error)
