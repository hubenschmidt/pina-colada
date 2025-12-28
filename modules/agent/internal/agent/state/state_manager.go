package state

import "context"

// Message represents a chat message in the conversation.
type Message struct {
	Role    string // "user" or "assistant"
	Content string
}

// UserFact represents a persistent fact about a user.
type UserFact struct {
	Key   string
	Value string
}

// SessionState holds the current state of a session.
type SessionState struct {
	SessionID string
	UserID    string
	Messages  []Message
}

// StateManager defines the interface for managing conversation state.
type StateManager interface {
	// GetSession retrieves an existing session or returns nil if not found.
	GetSession(ctx context.Context, userID, sessionID string) (*SessionState, error)

	// CreateSession creates a new session.
	CreateSession(ctx context.Context, userID, sessionID string) (*SessionState, error)

	// AddMessage adds a message to a session.
	AddMessage(ctx context.Context, sessionID string, msg Message) error

	// GetMessages retrieves messages up to a token limit (sliding window).
	GetMessages(ctx context.Context, sessionID string, tokenLimit int) ([]Message, error)

	// GetUserMemory retrieves persistent facts about a user.
	GetUserMemory(ctx context.Context, userID string) ([]UserFact, error)

	// AddUserFact adds a persistent fact about a user.
	AddUserFact(ctx context.Context, userID string, fact UserFact) error
}
