package models

import "time"

type LeadProject struct {
	LeadID    int64     `gorm:"primaryKey" json:"lead_id"`
	ProjectID int64     `gorm:"primaryKey" json:"project_id"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (LeadProject) TableName() string {
	return "Lead_Project"
}
