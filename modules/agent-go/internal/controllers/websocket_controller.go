package controllers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/pina-colada-co/agent-go/internal/agent"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins in development
	},
}

// WebSocketController handles WebSocket connections for agent chat
type WebSocketController struct {
	orchestrator *agent.Orchestrator
}

// NewWebSocketController creates a new WebSocket controller
func NewWebSocketController(orchestrator *agent.Orchestrator) *WebSocketController {
	return &WebSocketController{orchestrator: orchestrator}
}

// WSMessage represents incoming WebSocket messages
type WSMessage struct {
	UUID         string `json:"uuid"`
	Message      string `json:"message"`
	Init         bool   `json:"init,omitempty"`
	UserID       int64  `json:"user_id,omitempty"`
	TenantID     int64  `json:"tenant_id,omitempty"`
	Type         string `json:"type,omitempty"`
	UseEvaluator bool   `json:"use_evaluator,omitempty"`
}

// Client represents a connected WebSocket client
type Client struct {
	conn     *websocket.Conn
	uuid     string
	userID   int64
	tenantID int64
	mu       sync.Mutex
}

// SendJSON sends a JSON message to the client
func (c *Client) SendJSON(v interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.conn.WriteJSON(v)
}

// HandleWS handles WebSocket connections at /ws
func (wc *WebSocketController) HandleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	client := &Client{conn: conn}
	log.Printf("WebSocket client connected")

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			log.Printf("WebSocket client disconnected: %s", client.uuid)
			return
		}
		wc.processWSMessage(r.Context(), client, message)
	}
}

func (wc *WebSocketController) processWSMessage(ctx context.Context, client *Client, message []byte) {
	var msg WSMessage
	if err := json.Unmarshal(message, &msg); err != nil {
		log.Printf("Invalid message format: %v", err)
		return
	}

	// Handle init message
	if msg.Init {
		client.uuid = msg.UUID
		log.Printf("WebSocket initialized with UUID: %s", client.uuid)
		return
	}

	// Handle user context messages (ignore for now)
	if msg.Type == "user_context" || msg.Type == "user_context_update" {
		return
	}

	// Store user/tenant IDs if provided
	if msg.UserID > 0 {
		client.userID = msg.UserID
	}
	if msg.TenantID > 0 {
		client.tenantID = msg.TenantID
	}

	// Process chat message
	if msg.Message != "" {
		wc.handleChatMessage(ctx, client, msg)
	}
}

// handleChatMessage processes a chat message and streams the response in real-time
func (wc *WebSocketController) handleChatMessage(ctx context.Context, client *Client, msg WSMessage) {
	if wc.orchestrator == nil {
		client.SendJSON(map[string]interface{}{
			"type":    "error",
			"message": "Agent not configured - OPENAI_API_KEY required",
		})
		return
	}

	// Signal start of response
	client.SendJSON(map[string]bool{"on_chat_model_start": true})

	// Get user ID as string
	userID := "0"
	if client.userID > 0 {
		userID = strconv.FormatInt(client.userID, 10)
	}

	// Create event channel for streaming
	eventCh := make(chan agent.StreamEvent, 100)

	// Run the agent with streaming in a goroutine
	go wc.orchestrator.RunWithStreaming(ctx, agent.RunRequest{
		SessionID:    client.uuid,
		UserID:       userID,
		TenantID:     client.tenantID,
		Message:      msg.Message,
		UseEvaluator: msg.UseEvaluator,
	}, eventCh)

	// Process streaming events
	var lastText string
	for evt := range eventCh {
		lastText = handleStreamEvent(client, evt, lastText)
	}
}

// handleStreamEvent processes a single stream event and returns updated lastText
func handleStreamEvent(client *Client, evt agent.StreamEvent, lastText string) string {
	if evt.Type == "start" {
		client.SendJSON(map[string]interface{}{
			"on_timer_start": map[string]interface{}{"elapsed_ms": evt.ElapsedMs},
		})
		return lastText
	}

	if evt.Type == "text" {
		client.SendJSON(map[string]string{"on_chat_model_stream": evt.Text})
		return evt.Text
	}

	if evt.Type == "tokens" {
		client.SendJSON(map[string]interface{}{
			"on_token_stream": map[string]interface{}{
				"input":      evt.Tokens.Input,
				"output":     evt.Tokens.Output,
				"total":      evt.Tokens.Total,
				"elapsed_ms": evt.ElapsedMs,
			},
		})
		return lastText
	}

	if evt.Type == "tool_start" {
		client.SendJSON(map[string]interface{}{
			"on_tool_start": map[string]interface{}{
				"tool_name":  evt.ToolName,
				"elapsed_ms": evt.ElapsedMs,
			},
		})
		return lastText
	}

	if evt.Type == "tool_end" {
		client.SendJSON(map[string]interface{}{
			"on_tool_end": map[string]interface{}{
				"tool_name":  evt.ToolName,
				"elapsed_ms": evt.ElapsedMs,
			},
		})
		return lastText
	}

	if evt.Type == "agent_start" {
		client.SendJSON(map[string]interface{}{
			"on_agent_start": map[string]interface{}{
				"agent_name": evt.AgentName,
				"elapsed_ms": evt.ElapsedMs,
			},
		})
		return lastText
	}

	if evt.Type == "done" {
		log.Printf("WS sending done event (on_chat_model_end)")
		client.SendJSON(map[string]bool{"on_chat_model_end": true})
		sendFinalTokenUsage(client, evt)
		return lastText
	}

	if evt.Type == "eval_start" {
		log.Printf("WS sending eval_start event")
		client.SendJSON(map[string]interface{}{
			"type": "eval_start",
		})
		return lastText
	}

	if evt.Type == "eval" && evt.EvalResult != nil {
		log.Printf("WS sending eval event: score=%d, success=%v, user_input=%v",
			evt.EvalResult.Score, evt.EvalResult.SuccessCriteriaMet, evt.EvalResult.UserInputNeeded)
		client.SendJSON(map[string]interface{}{
			"type": "eval",
			"eval_result": map[string]interface{}{
				"feedback":             evt.EvalResult.Feedback,
				"success_criteria_met": evt.EvalResult.SuccessCriteriaMet,
				"user_input_needed":    evt.EvalResult.UserInputNeeded,
				"score":                evt.EvalResult.Score,
			},
		})
		return lastText
	}

	if evt.Type == "error" {
		log.Printf("Agent error: %v", evt.Error)
		if lastText == "" {
			client.SendJSON(map[string]string{
				"on_chat_model_stream": "\n\nSorry, there was an error generating the response.",
			})
		}
		client.SendJSON(map[string]bool{"on_chat_model_end": true})
		client.SendJSON(map[string]interface{}{
			"type":       "error",
			"message":    evt.Error,
			"elapsed_ms": evt.ElapsedMs,
		})
		return lastText
	}

	return lastText
}

func sendFinalTokenUsage(client *Client, evt agent.StreamEvent) {
	if evt.TurnTokens == nil || evt.CumulativeTokens == nil {
		return
	}
	client.SendJSON(map[string]interface{}{
		"on_token_usage": map[string]interface{}{
			"input":      evt.TurnTokens.Input,
			"output":     evt.TurnTokens.Output,
			"total":      evt.TurnTokens.Total,
			"elapsed_ms": evt.ElapsedMs,
		},
		"on_token_cumulative": map[string]interface{}{
			"input":  evt.CumulativeTokens.Input,
			"output": evt.CumulativeTokens.Output,
			"total":  evt.CumulativeTokens.Total,
		},
	})
	log.Printf("WS sending final token usage: turn=%+v cumulative=%+v elapsed=%dms",
		evt.TurnTokens, evt.CumulativeTokens, evt.ElapsedMs)
}
