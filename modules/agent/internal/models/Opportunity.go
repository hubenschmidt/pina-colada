package models

import (
	"time"

	"github.com/shopspring/decimal"
)

type Opportunity struct {
	ID                int64            `gorm:"primaryKey" json:"id"` // FK to Lead.id
	OpportunityName   string           `gorm:"not null" json:"opportunity_name"`
	EstimatedValue    *decimal.Decimal `gorm:"type:numeric(18,2)" json:"estimated_value"`
	Probability       *int16           `json:"probability"` // 0-100
	ExpectedCloseDate *time.Time       `gorm:"type:date" json:"expected_close_date"`
	Description       *string          `json:"description"`
	CreatedAt         time.Time        `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time        `gorm:"autoUpdateTime" json:"updated_at"`

	// Relations
	Lead *Lead `gorm:"foreignKey:ID;references:ID" json:"lead,omitempty"`
}

func (Opportunity) TableName() string {
	return "Opportunity"
}
