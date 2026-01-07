package models

import "time"

type AutomationRunLog struct {
	ID                   int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	AutomationConfigID   int64      `gorm:"column:automation_config_id;not null" json:"automation_config_id"`
	StartedAt            time.Time  `gorm:"column:started_at;not null" json:"started_at"`
	CompletedAt          *time.Time `gorm:"column:completed_at" json:"completed_at"`
	Status               string     `gorm:"column:status;not null;default:running" json:"status"`
	LeadsFound           int        `gorm:"column:leads_found;not null;default:0" json:"leads_found"`
	ProposalsCreated     int        `gorm:"column:proposals_created;not null;default:0" json:"proposals_created"`
	ErrorMessage         *string    `gorm:"column:error_message" json:"error_message"`
	SearchQuery          *string    `gorm:"column:search_query" json:"search_query"`
}

func (AutomationRunLog) TableName() string {
	return "Automation_Run_Log"
}
