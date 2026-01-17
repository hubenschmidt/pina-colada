package models

import "time"

type Job struct {
	ID                   int64      `gorm:"primaryKey" json:"id"` // FK to Lead.id
	JobTitle             string     `gorm:"not null" json:"job_title"`
	Description          *string    `json:"description"` // was "notes", renamed in migration 043
	JobURL               *string    `json:"job_url"`
	ResumeDate           *time.Time `json:"resume_date"`
	DatePosted           *time.Time `json:"date_posted"`
	DatePostedConfidence *string    `json:"date_posted_confidence"` // high, medium, low, none
	SalaryRange          *string    `json:"salary_range"`           // Legacy field
	SalaryRangeID        *int64     `json:"salary_range_id"`

	// Relations
	Lead           *Lead        `gorm:"foreignKey:ID" json:"lead,omitempty"`
	SalaryRangeRef *SalaryRange `gorm:"foreignKey:SalaryRangeID" json:"salary_range_ref,omitempty"`
}

func (Job) TableName() string {
	return "Job"
}
