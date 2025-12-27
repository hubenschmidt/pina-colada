package models

import "time"

type FundingRound struct {
	ID             int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	OrganizationID int64      `gorm:"not null" json:"organization_id"`
	RoundType      string     `gorm:"not null" json:"round_type"` // Pre-Seed, Seed, Series A, etc.
	Amount         *int64     `json:"amount"`                     // USD cents
	AnnouncedDate  *time.Time `json:"announced_date"`
	LeadInvestor   *string    `json:"lead_investor"`
	SourceURL      *string    `json:"source_url"`
	CreatedAt      time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

func (FundingRound) TableName() string {
	return "Funding_Round"
}
