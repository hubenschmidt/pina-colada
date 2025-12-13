package repositories

import (
	"errors"

	"github.com/google/uuid"
	apperrors "github.com/pina-colada-co/agent-go/internal/errors"
	"github.com/pina-colada-co/agent-go/internal/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// ConversationRepository handles conversation data access
type ConversationRepository struct {
	db *gorm.DB
}

// NewConversationRepository creates a new conversation repository
func NewConversationRepository(db *gorm.DB) *ConversationRepository {
	return &ConversationRepository{db: db}
}

// FindAll returns conversations for a user
func (r *ConversationRepository) FindAll(userID int64, tenantID *int64, includeArchived bool, limit int) ([]models.Conversation, error) {
	var conversations []models.Conversation

	query := r.db.Model(&models.Conversation{}).
		Where("user_id = ?", userID).
		Order("updated_at DESC")

	if tenantID != nil {
		query = query.Where("tenant_id = ?", *tenantID)
	}

	if !includeArchived {
		query = query.Where("archived_at IS NULL")
	}

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&conversations).Error
	return conversations, err
}

// FindByID returns a conversation by ID
func (r *ConversationRepository) FindByID(id int64) (*models.Conversation, error) {
	var conv models.Conversation
	err := r.db.First(&conv, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperrors.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &conv, nil
}

// FindByThreadID returns a conversation by thread ID
func (r *ConversationRepository) FindByThreadID(threadID string, userID int64) (*models.Conversation, error) {
	var conv models.Conversation
	err := r.db.Where("thread_id = ? AND user_id = ?", threadID, userID).First(&conv).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperrors.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &conv, nil
}

// FindAllForTenant returns all conversations for a tenant
func (r *ConversationRepository) FindAllForTenant(tenantID int64, search string, includeArchived bool, limit, offset int) ([]models.Conversation, int64, error) {
	var conversations []models.Conversation
	var totalCount int64

	query := r.db.Model(&models.Conversation{}).
		Where("tenant_id = ?", tenantID)

	if search != "" {
		searchTerm := "%" + search + "%"
		query = query.Where("LOWER(title) LIKE LOWER(?)", searchTerm)
	}

	if !includeArchived {
		query = query.Where("archived_at IS NULL")
	}

	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	query = query.Order("updated_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	err := query.Find(&conversations).Error
	return conversations, totalCount, err
}

// GetMessages returns messages for a conversation
func (r *ConversationRepository) GetMessages(conversationID int64) ([]models.ConversationMessage, error) {
	var messages []models.ConversationMessage
	err := r.db.Where("conversation_id = ?", conversationID).
		Order("created_at ASC").
		Find(&messages).Error
	return messages, err
}

// GetMessagesByThreadID returns recent messages for a conversation by thread ID
func (r *ConversationRepository) GetMessagesByThreadID(threadID uuid.UUID, limit int) ([]models.ConversationMessage, error) {
	var conv models.Conversation
	err := r.db.Where("thread_id = ?", threadID).First(&conv).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return []models.ConversationMessage{}, nil
	}
	if err != nil {
		return nil, err
	}

	var messages []models.ConversationMessage
	err = r.db.Where("conversation_id = ?", conv.ID).
		Order("created_at DESC").
		Limit(limit).
		Find(&messages).Error
	if err != nil {
		return nil, err
	}

	// Reverse to get chronological order
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

// UpdateTitle updates a conversation's title
func (r *ConversationRepository) UpdateTitle(conversationID int64, title string) error {
	return r.db.Model(&models.Conversation{}).
		Where("id = ?", conversationID).
		Update("title", title).Error
}

// Archive archives a conversation
func (r *ConversationRepository) Archive(conversationID int64) error {
	return r.db.Model(&models.Conversation{}).
		Where("id = ?", conversationID).
		Update("archived_at", gorm.Expr("NOW()")).Error
}

// Unarchive unarchives a conversation
func (r *ConversationRepository) Unarchive(conversationID int64) error {
	return r.db.Model(&models.Conversation{}).
		Where("id = ?", conversationID).
		Update("archived_at", nil).Error
}

// Delete permanently deletes a conversation and its messages
func (r *ConversationRepository) Delete(conversationID int64) error {
	tx := r.db.Begin()

	// Delete messages first
	if err := tx.Where("conversation_id = ?", conversationID).
		Delete(&models.ConversationMessage{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Delete conversation
	if err := tx.Delete(&models.Conversation{}, conversationID).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// GetOrCreateConversation returns an existing conversation by thread ID or creates a new one
// Returns the conversation, whether it was newly created, and any error
func (r *ConversationRepository) GetOrCreateConversation(userID, tenantID int64, threadID uuid.UUID) (*models.Conversation, bool, error) {
	var conv models.Conversation
	err := r.db.Where("thread_id = ?", threadID).First(&conv).Error
	if err == nil {
		return &conv, false, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, false, err
	}

	// Create new conversation
	conv = models.Conversation{
		TenantID:    tenantID,
		UserID:      userID,
		ThreadID:    threadID,
		CreatedByID: userID,
		UpdatedByID: userID,
	}
	if err := r.db.Create(&conv).Error; err != nil {
		return nil, false, err
	}
	return &conv, true, nil
}

// AddMessage adds a message to a conversation
func (r *ConversationRepository) AddMessage(conversationID int64, role, content string, tokenUsage datatypes.JSON) error {
	msg := models.ConversationMessage{
		ConversationID: conversationID,
		Role:           role,
		Content:        content,
		TokenUsage:     tokenUsage,
	}
	return r.db.Create(&msg).Error
}

// TouchConversation updates the updated_at timestamp for a conversation
func (r *ConversationRepository) TouchConversation(threadID uuid.UUID) error {
	return r.db.Model(&models.Conversation{}).
		Where("thread_id = ?", threadID).
		Update("updated_at", gorm.Expr("NOW()")).Error
}

// UpdateTitleByThreadID updates the title of a conversation by thread ID
func (r *ConversationRepository) UpdateTitleByThreadID(threadID uuid.UUID, title string) error {
	return r.db.Model(&models.Conversation{}).
		Where("thread_id = ?", threadID).
		Update("title", title).Error
}
