package state

import (
	"context"
	"fmt"
	"sync"

	apperrors "github.com/pina-colada-co/agent-go/internal/errors"
)

// MemoryStateManager is an in-memory implementation of StateManager.
type MemoryStateManager struct {
	mu         sync.RWMutex
	sessions   map[string]*SessionState // sessionID -> session
	userMemory map[string][]UserFact    // userID -> facts
}

// NewMemoryStateManager creates a new in-memory state manager.
func NewMemoryStateManager() *MemoryStateManager {
	return &MemoryStateManager{
		sessions:   make(map[string]*SessionState),
		userMemory: make(map[string][]UserFact),
	}
}

func (m *MemoryStateManager) GetSession(ctx context.Context, userID, sessionID string) (*SessionState, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	session, ok := m.sessions[sessionID]
	if !ok {
		return nil, apperrors.ErrSessionNotFound
	}
	return session, nil
}

func (m *MemoryStateManager) CreateSession(ctx context.Context, userID, sessionID string) (*SessionState, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	session := &SessionState{
		SessionID: sessionID,
		UserID:    userID,
		Messages:  []Message{},
	}
	m.sessions[sessionID] = session
	return session, nil
}

func (m *MemoryStateManager) AddMessage(ctx context.Context, sessionID string, msg Message) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	session, ok := m.sessions[sessionID]
	if !ok {
		return fmt.Errorf("session %s not found", sessionID)
	}

	session.Messages = append(session.Messages, msg)
	return nil
}

func (m *MemoryStateManager) GetMessages(ctx context.Context, sessionID string, tokenLimit int) ([]Message, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	session, ok := m.sessions[sessionID]
	if !ok {
		return nil, apperrors.ErrSessionNotFound
	}

	// Estimate tokens: ~4 chars per token
	const charsPerToken = 4
	maxChars := tokenLimit * charsPerToken

	// Sliding window: take most recent messages that fit
	msgs := session.Messages
	if len(msgs) == 0 {
		return []Message{}, nil
	}

	startIdx := findMessageWindowStart(msgs, maxChars)
	result := make([]Message, len(msgs)-startIdx)
	copy(result, msgs[startIdx:])
	return result, nil
}

// findMessageWindowStart finds the earliest message index that fits within maxChars.
func findMessageWindowStart(msgs []Message, maxChars int) int {
	totalChars := 0
	startIdx := len(msgs)
	for i := len(msgs) - 1; i >= 0; i-- {
		msgChars := len(msgs[i].Content)
		if totalChars+msgChars > maxChars {
			return startIdx
		}
		totalChars += msgChars
		startIdx = i
	}
	return startIdx
}

func (m *MemoryStateManager) GetUserMemory(ctx context.Context, userID string) ([]UserFact, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	facts, ok := m.userMemory[userID]
	if !ok {
		return []UserFact{}, nil
	}

	result := make([]UserFact, len(facts))
	copy(result, facts)
	return result, nil
}

func (m *MemoryStateManager) AddUserFact(ctx context.Context, userID string, fact UserFact) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Update existing fact with same key, or append
	facts := m.userMemory[userID]
	for i, f := range facts {
		if f.Key == fact.Key {
			facts[i] = fact
			return nil
		}
	}

	m.userMemory[userID] = append(facts, fact)
	return nil
}
