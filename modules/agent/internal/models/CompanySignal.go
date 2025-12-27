package models

import (
	"time"

	"github.com/shopspring/decimal"
)

type CompanySignal struct {
	ID             int64            `gorm:"primaryKey;autoIncrement" json:"id"`
	OrganizationID int64            `gorm:"not null" json:"organization_id"`
	SignalType     string           `gorm:"not null" json:"signal_type"` // hiring, expansion, product_launch, etc.
	Headline       string           `gorm:"not null" json:"headline"`
	Description    *string          `json:"description"`
	SignalDate     *time.Time       `json:"signal_date"`
	Source         *string          `json:"source"`    // linkedin, news, crunchbase, agent
	SourceURL      *string          `json:"source_url"`
	Sentiment      *string          `json:"sentiment"` // positive, neutral, negative
	RelevanceScore *decimal.Decimal `gorm:"type:numeric(3,2)" json:"relevance_score"`
	CreatedAt      time.Time        `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time        `gorm:"autoUpdateTime" json:"updated_at"`
}

func (CompanySignal) TableName() string {
	return "Company_Signal"
}
