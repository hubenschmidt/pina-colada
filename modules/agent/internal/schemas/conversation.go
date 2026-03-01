package schemas

import (
	"github.com/google/uuid"
)

// ConversationCreate represents the request body for creating a conversation
type ConversationCreate struct {
	ThreadID uuid.UUID `json:"thread_id" validate:"required"`
	Title    *string   `json:"title"`
}

// ConversationUpdate represents the request body for updating a conversation
type ConversationUpdate struct {
	Title string `json:"title" validate:"required"`
}

// MessageCreate represents the request body for creating a message
type MessageCreate struct {
	Role       string         `json:"role" validate:"required"`
	Content    string         `json:"content" validate:"required"`
	TokenUsage map[string]any `json:"token_usage"`
}

