package models

import "time"

type Partnership struct {
	ID              int64      `gorm:"primaryKey" json:"id"` // FK to Lead.id
	OrganizationID  int64      `gorm:"not null" json:"organization_id"`
	PartnershipType *string    `json:"partnership_type"`
	PartnershipName string     `gorm:"not null" json:"partnership_name"`
	StartDate       *time.Time `gorm:"type:date" json:"start_date"`
	EndDate         *time.Time `gorm:"type:date" json:"end_date"`
	Notes           *string    `json:"notes"`
	CreatedAt       time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

func (Partnership) TableName() string {
	return "Partnership"
}
