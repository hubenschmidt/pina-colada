package models

import "time"

type UsageAnalytics struct {
	ID             int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID       int64     `gorm:"not null;index" json:"tenant_id"`
	UserID         int64     `gorm:"not null;index" json:"user_id"`
	ConversationID *int64    `gorm:"index" json:"conversation_id"`
	MessageID      *int64    `gorm:"index" json:"message_id"`
	InputTokens    int       `gorm:"not null;default:0" json:"input_tokens"`
	OutputTokens   int       `gorm:"not null;default:0" json:"output_tokens"`
	TotalTokens    int       `gorm:"not null;default:0" json:"total_tokens"`
	NodeName       *string   `json:"node_name"`
	ToolName       *string   `json:"tool_name"`
	ModelName      *string   `json:"model_name"`
	CreatedAt      time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (UsageAnalytics) TableName() string {
	return "Usage_Analytics"
}
