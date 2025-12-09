package repositories

import (
	"github.com/pina-colada-co/agent-go/internal/models"
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

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &conv, nil
}

// FindByThreadID returns a conversation by thread ID
func (r *ConversationRepository) FindByThreadID(threadID string, userID int64) (*models.Conversation, error) {
	var conv models.Conversation
	err := r.db.Where("thread_id = ? AND user_id = ?", threadID, userID).First(&conv).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
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
