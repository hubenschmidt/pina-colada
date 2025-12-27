package models

import (
	"time"

	"gorm.io/datatypes"
)

type SavedReport struct {
	ID              int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID        int64          `gorm:"not null;index" json:"tenant_id"`
	Name            string         `gorm:"type:text;not null" json:"name"`
	Description     *string        `gorm:"type:text" json:"description"`
	QueryDefinition datatypes.JSON `gorm:"type:jsonb;not null" json:"query_definition"`
	CreatedBy       *int64         `json:"created_by"`
	CreatedAt       time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
}

func (SavedReport) TableName() string {
	return "Saved_Report"
}
