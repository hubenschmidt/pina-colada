package models

import (
	"time"

	"github.com/shopspring/decimal"
)

type OrganizationTechnology struct {
	OrganizationID int64            `gorm:"primaryKey" json:"organization_id"`
	TechnologyID   int64            `gorm:"primaryKey" json:"technology_id"`
	DetectedAt     time.Time        `gorm:"autoCreateTime" json:"detected_at"`
	Source         *string          `json:"source"` // builtwith, wappalyzer, agent, manual
	Confidence     *decimal.Decimal `gorm:"type:numeric(3,2)" json:"confidence"`
	CreatedAt      time.Time        `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time        `gorm:"autoUpdateTime" json:"updated_at"`
}

func (OrganizationTechnology) TableName() string {
	return "Organization_Technology"
}
