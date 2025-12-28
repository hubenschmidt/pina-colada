package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type Conversation struct {
	ID          int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID    int64      `gorm:"index;not null" json:"tenant_id"`
	UserID      int64      `gorm:"not null" json:"user_id"`
	ThreadID    uuid.UUID  `gorm:"type:uuid;uniqueIndex;not null" json:"thread_id"`
	Title       *string    `json:"title"`
	CreatedAt   time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
	ArchivedAt  *time.Time `json:"archived_at"`
	CreatedByID int64      `gorm:"not null" json:"created_by_id"`
	UpdatedByID int64      `gorm:"not null" json:"updated_by_id"`
}

func (Conversation) TableName() string {
	return "Conversation"
}

type ConversationMessage struct {
	ID             int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	ConversationID int64          `gorm:"index;not null" json:"conversation_id"`
	Role           string         `gorm:"not null" json:"role"` // "user" or "assistant"
	Content        string         `gorm:"not null" json:"content"`
	TokenUsage     datatypes.JSON `json:"token_usage,omitempty"`
	CreatedAt      time.Time      `gorm:"autoCreateTime" json:"created_at"`
}

func (ConversationMessage) TableName() string {
	return "Conversation_Message"
}
