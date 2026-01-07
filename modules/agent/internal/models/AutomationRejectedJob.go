package models

import "time"

type AutomationRejectedJob struct {
	ID                   int64     `gorm:"primaryKey" json:"id"`
	AutomationConfigID   int64     `gorm:"not null" json:"automation_config_id"`
	TenantID             int64     `gorm:"not null" json:"tenant_id"`
	URL                  string    `gorm:"not null" json:"url"`
	JobTitle             *string   `json:"job_title"`
	Company              *string   `json:"company"`
	Reason               *string   `json:"reason"`
	CreatedAt            time.Time `gorm:"not null;default:now()" json:"created_at"`
}

func (AutomationRejectedJob) TableName() string {
	return "Automation_Rejected_Job"
}
