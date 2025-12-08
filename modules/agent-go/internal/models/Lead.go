package models

import "time"

type Lead struct {
	ID                int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID          *int64    `gorm:"index" json:"tenant_id"`
	AccountID         *int64    `gorm:"index" json:"account_id"`
	DealID            int64     `gorm:"not null" json:"deal_id"`
	Type              string    `gorm:"not null" json:"type"` // Job, Opportunity, Partnership
	Description       *string   `json:"description"`
	Source            *string   `json:"source"` // inbound, referral, event, campaign, agent, manual
	CurrentStatusID   *int64    `json:"current_status_id"`
	OwnerIndividualID *int64    `json:"owner_individual_id"`
	CreatedAt         time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	CreatedBy         int64     `gorm:"not null" json:"created_by"`
	UpdatedBy         int64     `gorm:"not null" json:"updated_by"`

	// Relations
	CurrentStatus *Status `gorm:"foreignKey:CurrentStatusID" json:"current_status,omitempty"`
	Account       *Account    `gorm:"foreignKey:AccountID" json:"account,omitempty"`
	Projects      []Project   `gorm:"many2many:Lead_Project;joinForeignKey:lead_id;joinReferences:project_id" json:"projects,omitempty"`
}

func (Lead) TableName() string {
	return "Lead"
}
