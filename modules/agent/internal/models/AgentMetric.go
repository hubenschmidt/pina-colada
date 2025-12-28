package models

import (
	"time"

	"gorm.io/datatypes"
)

type AgentMetric struct {
	ID               int64          `gorm:"primaryKey" json:"id"`
	SessionID        int64          `gorm:"not null" json:"session_id"`
	ConversationID   *int64         `json:"conversation_id,omitempty"`
	ThreadID         *string        `json:"thread_id,omitempty"`
	StartedAt        time.Time      `gorm:"not null" json:"started_at"`
	EndedAt          time.Time      `gorm:"not null" json:"ended_at"`
	DurationMs       int            `gorm:"not null" json:"duration_ms"`
	InputTokens      int            `gorm:"not null" json:"input_tokens"`
	OutputTokens     int            `gorm:"not null" json:"output_tokens"`
	TotalTokens      int            `gorm:"not null" json:"total_tokens"`
	EstimatedCostUSD *float64       `gorm:"column:estimated_cost_usd;type:decimal(10,6)" json:"estimated_cost_usd,omitempty"`
	Model            string         `gorm:"not null;size:100" json:"model"`
	Provider         string         `gorm:"not null;size:50" json:"provider"`
	NodeName         *string        `gorm:"size:100" json:"node_name,omitempty"`
	ConfigSnapshot   datatypes.JSON `json:"config_snapshot,omitempty"`
	UserMessage      *string        `json:"user_message,omitempty"`
	CreatedAt        time.Time      `gorm:"autoCreateTime" json:"created_at"`
}

func (AgentMetric) TableName() string {
	return "Agent_Metric"
}
