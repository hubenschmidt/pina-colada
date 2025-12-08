package models

import "time"

type Individual struct {
	ID              int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	AccountID       *int64    `json:"account_id"`
	FirstName       string    `gorm:"not null" json:"first_name"`
	LastName        string    `gorm:"not null" json:"last_name"`
	Email           *string   `json:"email"`
	Phone           *string   `json:"phone"`
	LinkedInURL     *string   `json:"linkedin_url"`
	Title           *string   `json:"title"`
	Description     *string   `json:"description"`
	CreatedAt       time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	CreatedBy       int64     `gorm:"not null" json:"created_by"`
	UpdatedBy       int64     `gorm:"not null" json:"updated_by"`

	// Contact intelligence columns
	TwitterURL      *string `json:"twitter_url"`
	GithubURL       *string `json:"github_url"`
	Bio             *string `json:"bio"`
	SeniorityLevel  *string `json:"seniority_level"`  // C-Level, VP, Director, Manager, IC
	Department      *string `json:"department"`       // Engineering, Sales, Marketing
	IsDecisionMaker *bool   `json:"is_decision_maker"`
	ReportsToID     *int64  `json:"reports_to_id"`

	// Relations (for eager loading)
	Account *Account `gorm:"foreignKey:AccountID" json:"account,omitempty"`
}

func (Individual) TableName() string {
	return "Individual"
}
