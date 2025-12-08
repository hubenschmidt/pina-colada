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
