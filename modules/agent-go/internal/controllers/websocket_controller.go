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
	"github.com/pina-colada-co/agent-go/internal/services"
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
	wsService    *services.WebSocketService
}

// NewWebSocketController creates a new WebSocket controller
func NewWebSocketController(orchestrator *agent.Orchestrator, wsService *services.WebSocketService) *WebSocketController {
	return &WebSocketController{
		orchestrator: orchestrator,
		wsService:    wsService,
	}
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

	wc.messageLoop(r.Context(), client)
}

func (wc *WebSocketController) messageLoop(ctx context.Context, client *Client) {
	for {
		_, message, err := client.conn.ReadMessage()
		if err != nil {
			wc.handleReadError(err, client)
			return
		}
		wc.processWSMessage(ctx, client, message)
	}
}

func (wc *WebSocketController) handleReadError(err error, client *Client) {
	if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
		log.Printf("WebSocket error: %v", err)
	}
	log.Printf("WebSocket client disconnected: %s", client.uuid)
	wc.wsService.RemoveClient(client.uuid)
}

func (wc *WebSocketController) processWSMessage(ctx context.Context, client *Client, message []byte) {
	var msg WSMessage
	if err := json.Unmarshal(message, &msg); err != nil {
		log.Printf("Invalid message format: %v", err)
		return
	}

	if msg.Init {
		client.uuid = msg.UUID
		log.Printf("WebSocket initialized with UUID: %s", client.uuid)
		return
	}

	if msg.Type == "user_context" || msg.Type == "user_context_update" {
		return
	}

	wc.updateClientContext(client, msg)

	if msg.Message == "" {
		return
	}

	wc.handleChatMessage(ctx, client, msg)
}

func (wc *WebSocketController) updateClientContext(client *Client, msg WSMessage) {
	if msg.UserID > 0 {
		client.userID = msg.UserID
	}
	if msg.TenantID > 0 {
		client.tenantID = msg.TenantID
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

	result := wc.wsService.CheckAndMarkProcessing(client.uuid, msg.UUID, msg.Message)
	if result == services.DedupBlocked || result == services.DedupBusy {
		return
	}

	defer wc.wsService.MarkProcessingComplete(client.uuid)

	client.SendJSON(map[string]bool{"on_chat_model_start": true})

	userID := formatUserID(client.userID)
	eventCh := make(chan agent.StreamEvent, 100)

	go wc.orchestrator.RunWithStreaming(ctx, agent.RunRequest{
		SessionID:    client.uuid,
		UserID:       userID,
		TenantID:     client.tenantID,
		Message:      msg.Message,
		UseEvaluator: msg.UseEvaluator,
	}, eventCh)

	wc.processStreamEvents(client, eventCh)
}

func formatUserID(userID int64) string {
	if userID > 0 {
		return strconv.FormatInt(userID, 10)
	}
	return "0"
}

func (wc *WebSocketController) processStreamEvents(client *Client, eventCh <-chan agent.StreamEvent) {
	var lastText string
	for evt := range eventCh {
		lastText = wc.handleStreamEvent(client, evt, lastText)
	}
}

func (wc *WebSocketController) handleStreamEvent(client *Client, evt agent.StreamEvent, lastText string) string {
	handler := streamEventHandlers[evt.Type]
	if handler == nil {
		return lastText
	}
	return handler(client, evt, lastText)
}

type streamEventHandler func(client *Client, evt agent.StreamEvent, lastText string) string

var streamEventHandlers = map[string]streamEventHandler{
	"start":      handleStartEvent,
	"text":       handleTextEvent,
	"tokens":     handleTokensEvent,
	"tool_start": handleToolStartEvent,
	"tool_end":   handleToolEndEvent,
	"agent_start": handleAgentStartEvent,
	"done":       handleDoneEvent,
	"eval_start": handleEvalStartEvent,
	"eval":       handleEvalEvent,
	"error":      handleErrorEvent,
}

func handleStartEvent(client *Client, evt agent.StreamEvent, lastText string) string {
	client.SendJSON(map[string]interface{}{
		"on_timer_start": map[string]interface{}{"elapsed_ms": evt.ElapsedMs},
	})
	return lastText
}

func handleTextEvent(client *Client, evt agent.StreamEvent, lastText string) string {
	client.SendJSON(map[string]string{"on_chat_model_stream": evt.Text})
	return evt.Text
}

func handleTokensEvent(client *Client, evt agent.StreamEvent, lastText string) string {
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

func handleToolStartEvent(client *Client, evt agent.StreamEvent, lastText string) string {
	client.SendJSON(map[string]interface{}{
		"on_tool_start": map[string]interface{}{
			"tool_name":  evt.ToolName,
			"elapsed_ms": evt.ElapsedMs,
		},
	})
	return lastText
}

func handleToolEndEvent(client *Client, evt agent.StreamEvent, lastText string) string {
	client.SendJSON(map[string]interface{}{
		"on_tool_end": map[string]interface{}{
			"tool_name":  evt.ToolName,
			"elapsed_ms": evt.ElapsedMs,
		},
	})
	return lastText
}

func handleAgentStartEvent(client *Client, evt agent.StreamEvent, lastText string) string {
	client.SendJSON(map[string]interface{}{
		"on_agent_start": map[string]interface{}{
			"agent_name": evt.AgentName,
			"elapsed_ms": evt.ElapsedMs,
		},
	})
	return lastText
}

func handleDoneEvent(client *Client, evt agent.StreamEvent, lastText string) string {
	log.Printf("WS sending done event (on_chat_model_end)")
	client.SendJSON(map[string]bool{"on_chat_model_end": true})
	sendFinalTokenUsage(client, evt)
	return lastText
}

func handleEvalStartEvent(client *Client, evt agent.StreamEvent, lastText string) string {
	log.Printf("WS sending eval_start event")
	client.SendJSON(map[string]interface{}{"type": "eval_start"})
	return lastText
}

func handleEvalEvent(client *Client, evt agent.StreamEvent, lastText string) string {
	if evt.EvalResult == nil {
		return lastText
	}
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

func handleErrorEvent(client *Client, evt agent.StreamEvent, lastText string) string {
	log.Printf("Agent error: %v", evt.Error)
	sendErrorFallbackText(client, lastText)
	client.SendJSON(map[string]bool{"on_chat_model_end": true})
	client.SendJSON(map[string]interface{}{
		"type":       "error",
		"message":    evt.Error,
		"elapsed_ms": evt.ElapsedMs,
	})
	return lastText
}

func sendErrorFallbackText(client *Client, lastText string) {
	if lastText != "" {
		return
	}
	client.SendJSON(map[string]string{
		"on_chat_model_stream": "\n\nSorry, there was an error generating the response.",
	})
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
