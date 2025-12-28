package models

import "time"

type Technology struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string    `gorm:"not null" json:"name"`
	Category  string    `gorm:"not null" json:"category"` // CRM, Marketing Automation, Cloud, Database
	Vendor    *string   `json:"vendor"`                   // Salesforce, HubSpot, AWS
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (Technology) TableName() string {
	return "Technology"
}
