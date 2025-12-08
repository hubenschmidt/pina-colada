package models

import "time"

type SavedReportProject struct {
	SavedReportID int64     `gorm:"primaryKey" json:"saved_report_id"`
	ProjectID     int64     `gorm:"primaryKey" json:"project_id"`
	CreatedAt     time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (SavedReportProject) TableName() string {
	return "Saved_Report_Project"
}
