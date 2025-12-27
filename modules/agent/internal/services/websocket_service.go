package services

import (
	"crypto/sha256"
	"encoding/hex"
	"log"
	"sync"
	"time"
)

const dedupWindow = 5 * time.Second

// ClientState tracks per-client state for deduplication
type ClientState struct {
	mu           sync.Mutex
	lastMsgHash  string
	lastMsgTime  time.Time
	isProcessing bool
}

// WebSocketService handles WebSocket business logic
type WebSocketService struct {
	clients   map[string]*ClientState
	clientsMu sync.RWMutex
}

// NewWebSocketService creates a new WebSocket service
func NewWebSocketService() *WebSocketService {
	return &WebSocketService{
		clients: make(map[string]*ClientState),
	}
}

// GetOrCreateClientState returns existing client state or creates new one
func (s *WebSocketService) GetOrCreateClientState(clientID string) *ClientState {
	s.clientsMu.RLock()
	state, exists := s.clients[clientID]
	s.clientsMu.RUnlock()

	if exists {
		return state
	}

	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()

	// Double-check after acquiring write lock
	if state, exists = s.clients[clientID]; exists {
		return state
	}

	state = &ClientState{}
	s.clients[clientID] = state
	return state
}

// RemoveClient removes client state when disconnected
func (s *WebSocketService) RemoveClient(clientID string) {
	s.clientsMu.Lock()
	delete(s.clients, clientID)
	s.clientsMu.Unlock()
}

// DedupResult indicates whether a message should be processed
type DedupResult int

const (
	DedupAllowed DedupResult = iota
	DedupBlocked
	DedupBusy
)

// CheckAndMarkProcessing checks deduplication and marks client as processing
// Returns DedupAllowed if message should be processed, DedupBlocked if duplicate,
// DedupBusy if already processing another message
func (s *WebSocketService) CheckAndMarkProcessing(clientID, uuid, message string) DedupResult {
	state := s.GetOrCreateClientState(clientID)
	msgHash := hashMessage(uuid, message)

	state.mu.Lock()
	defer state.mu.Unlock()

	if msgHash == state.lastMsgHash && time.Since(state.lastMsgTime) < dedupWindow {
		log.Printf("DUPLICATE MESSAGE BLOCKED: hash=%s (within %v window)", msgHash, dedupWindow)
		return DedupBlocked
	}

	if state.isProcessing {
		log.Printf("MESSAGE BLOCKED: already processing a message for client %s", clientID)
		return DedupBusy
	}

	state.lastMsgHash = msgHash
	state.lastMsgTime = time.Now()
	state.isProcessing = true
	return DedupAllowed
}

// MarkProcessingComplete marks client as no longer processing
func (s *WebSocketService) MarkProcessingComplete(clientID string) {
	state := s.GetOrCreateClientState(clientID)
	state.mu.Lock()
	state.isProcessing = false
	state.mu.Unlock()
}

// hashMessage creates a hash of the message for deduplication
func hashMessage(uuid, message string) string {
	data := uuid + "|" + message
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:8])
}
