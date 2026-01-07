package models

import (
	"time"

	"gorm.io/datatypes"
)

type AgentProposal struct {
	ID           int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID     int64          `gorm:"not null;index" json:"tenant_id"`
	ProposedByID int64          `gorm:"not null" json:"proposed_by_id"`
	EntityType   string         `gorm:"not null" json:"entity_type"`
	EntityID     *int64         `json:"entity_id"`
	Operation    string         `gorm:"not null" json:"operation"`
	Payload          datatypes.JSON `gorm:"type:jsonb;not null" json:"payload"`
	Status           string         `gorm:"not null;default:pending" json:"status"`
	ValidationErrors datatypes.JSON `gorm:"type:jsonb" json:"validation_errors"`
	ReviewedByID *int64         `json:"reviewed_by_id"`
	ReviewedAt   *time.Time     `json:"reviewed_at"`
	ExecutedAt   *time.Time     `json:"executed_at"`
	ErrorMessage *string        `json:"error_message"`
	Source             *string        `json:"source"`
	AutomationConfigID *int64         `gorm:"column:automation_config_id" json:"automation_config_id"`
	CreatedAt          time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt          time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
}

func (AgentProposal) TableName() string {
	return "Agent_Proposal"
}
