package models

import (
	"time"

	"gorm.io/datatypes"
)

type AgentRecordingSession struct {
	ID                int64          `gorm:"primaryKey" json:"id"`
	UserID            int64          `gorm:"not null" json:"user_id"`
	TenantID          int64          `gorm:"not null" json:"tenant_id"`
	Name              *string        `json:"name,omitempty"`
	StartedAt         time.Time      `gorm:"not null;default:now()" json:"started_at"`
	EndedAt           *time.Time     `json:"ended_at,omitempty"`
	ConfigSnapshot    datatypes.JSON `json:"config_snapshot,omitempty"`
	MetricCount       int            `gorm:"default:0" json:"metric_count"`
	TotalTokens       int            `gorm:"default:0" json:"total_tokens"`
	TotalInputTokens  int            `gorm:"default:0" json:"total_input_tokens"`
	TotalOutputTokens int            `gorm:"default:0" json:"total_output_tokens"`
	TotalCostUSD      float64        `gorm:"default:0" json:"total_cost_usd"`
	TotalDurationMs   int            `gorm:"default:0" json:"total_duration_ms"`
	CreatedAt         time.Time      `gorm:"autoCreateTime" json:"created_at"`
}

func (AgentRecordingSession) TableName() string {
	return "Agent_Recording_Session"
}
