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

// MessageResponse represents a message in API responses
type MessageResponse struct {
	ID        int64     `json:"id"`
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// ConversationDetailResponse represents a conversation with messages
type ConversationDetailResponse struct {
	ConversationResponse
	Messages []MessageResponse `json:"messages"`
}

// TenantConversationsResponse represents paginated conversations
type TenantConversationsResponse struct {
	Conversations []ConversationResponse `json:"conversations"`
	Total         int64                  `json:"total"`
}

// GetTenantConversations returns all conversations for a tenant
func (s *ConversationService) GetTenantConversations(tenantID int64, search string, includeArchived bool, limit, offset int) (*TenantConversationsResponse, error) {
	conversations, total, err := s.convRepo.FindAllForTenant(tenantID, search, includeArchived, limit, offset)
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

	return &TenantConversationsResponse{
		Conversations: results,
		Total:         total,
	}, nil
}

// GetConversation returns a conversation with messages by thread ID
func (s *ConversationService) GetConversation(threadID string, userID int64) (*ConversationDetailResponse, error) {
	conv, err := s.convRepo.FindByThreadID(threadID, userID)
	if err != nil {
		return nil, err
	}

	messages, err := s.convRepo.GetMessages(conv.ID)
	if err != nil {
		return nil, err
	}

	msgResponses := make([]MessageResponse, len(messages))
	for i, m := range messages {
		msgResponses[i] = MessageResponse{
			ID:        m.ID,
			Role:      m.Role,
			Content:   m.Content,
			CreatedAt: m.CreatedAt,
		}
	}

	return &ConversationDetailResponse{
		ConversationResponse: ConversationResponse{
			ID:         conv.ID,
			ThreadID:   conv.ThreadID,
			Title:      conv.Title,
			CreatedAt:  conv.CreatedAt,
			UpdatedAt:  conv.UpdatedAt,
			ArchivedAt: conv.ArchivedAt,
		},
		Messages: msgResponses,
	}, nil
}

// UpdateTitle updates a conversation's title
func (s *ConversationService) UpdateTitle(threadID string, userID int64, title string) (*ConversationResponse, error) {
	conv, err := s.convRepo.FindByThreadID(threadID, userID)
	if err != nil {
		return nil, err
	}

	if err := s.convRepo.UpdateTitle(conv.ID, title); err != nil {
		return nil, err
	}

	return &ConversationResponse{
		ID:         conv.ID,
		ThreadID:   conv.ThreadID,
		Title:      &title,
		CreatedAt:  conv.CreatedAt,
		UpdatedAt:  time.Now(),
		ArchivedAt: conv.ArchivedAt,
	}, nil
}

// ArchiveConversation archives a conversation
func (s *ConversationService) ArchiveConversation(threadID string, userID int64) error {
	conv, err := s.convRepo.FindByThreadID(threadID, userID)
	if err != nil {
		return err
	}
	if conv == nil {
		return nil
	}

	return s.convRepo.Archive(conv.ID)
}

// UnarchiveConversation unarchives a conversation
func (s *ConversationService) UnarchiveConversation(threadID string, userID int64) error {
	conv, err := s.convRepo.FindByThreadID(threadID, userID)
	if err != nil {
		return err
	}
	if conv == nil {
		return nil
	}

	return s.convRepo.Unarchive(conv.ID)
}

// DeleteConversation permanently deletes a conversation
func (s *ConversationService) DeleteConversation(threadID string, userID int64) error {
	conv, err := s.convRepo.FindByThreadID(threadID, userID)
	if err != nil {
		return err
	}
	if conv == nil {
		return nil
	}

	// Only allow deletion of archived conversations
	if conv.ArchivedAt == nil {
		return nil
	}

	return s.convRepo.Delete(conv.ID)
}
