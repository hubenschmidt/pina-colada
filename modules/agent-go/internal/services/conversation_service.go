package services

import (
	"time"

	"github.com/google/uuid"
	"github.com/pina-colada-co/agent-go/internal/repositories"
)

// ConversationService handles conversation business logic
type ConversationService struct {
	convRepo *repositories.ConversationRepository
}

// NewConversationService creates a new conversation service
func NewConversationService(convRepo *repositories.ConversationRepository) *ConversationService {
	return &ConversationService{convRepo: convRepo}
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

// GetConversations returns conversations for a user
func (s *ConversationService) GetConversations(userID int64, tenantID *int64, includeArchived bool, limit int) ([]ConversationResponse, error) {
	conversations, err := s.convRepo.FindAll(userID, tenantID, includeArchived, limit)
	if err != nil {
		return nil, err
	}

	results := make([]ConversationResponse, len(conversations))
	for i, c := range conversations {
		results[i] = ConversationResponse{
			ID:         c.ID,
			ThreadID:   c.ThreadID,
			Title:      c.Title,
			CreatedAt:  c.CreatedAt,
			UpdatedAt:  c.UpdatedAt,
			ArchivedAt: c.ArchivedAt,
		}
	}

	return results, nil
}
