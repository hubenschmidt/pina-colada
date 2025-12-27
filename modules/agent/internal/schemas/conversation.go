package schemas

import (
	"time"

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

// MessageResponse represents a message in API responses
type MessageResponse struct {
	ID         int64          `json:"id"`
	Role       string         `json:"role"`
	Content    string         `json:"content"`
	TokenUsage map[string]any `json:"token_usage"`
	CreatedAt  time.Time      `json:"created_at"`
}

// UserBrief represents a brief user summary in API responses
type UserBrief struct {
	ID    int64   `json:"id"`
	Email *string `json:"email"`
}

// ConversationResponse represents a conversation in API responses
type ConversationResponse struct {
	ID         int64      `json:"id"`
	ThreadID   uuid.UUID  `json:"thread_id"`
	Title      *string    `json:"title"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	ArchivedAt *time.Time `json:"archived_at"`
}

// ConversationWithUserResponse represents a conversation with user details
type ConversationWithUserResponse struct {
	ConversationResponse
	CreatedByID *int64     `json:"created_by_id"`
	UpdatedByID *int64     `json:"updated_by_id"`
	CreatedBy   *UserBrief `json:"created_by"`
	UpdatedBy   *UserBrief `json:"updated_by"`
}

// ConversationWithMessagesResponse represents a conversation with its messages
type ConversationWithMessagesResponse struct {
	Conversation ConversationResponse `json:"conversation"`
	Messages     []MessageResponse    `json:"messages"`
}
