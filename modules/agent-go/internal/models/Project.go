package models

import "time"

type Project struct {
	ID                int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID          *int64     `gorm:"index" json:"tenant_id"`
	Name              string     `gorm:"not null" json:"name"`
	Description       *string    `json:"description"`
	OwnerUserID       *int64     `json:"owner_user_id"`
	Status            *string    `json:"status"`
	CurrentStatusID   *int64     `gorm:"index" json:"current_status_id"`
	StartDate         *time.Time `gorm:"type:date" json:"start_date"`
	EndDate           *time.Time `gorm:"type:date" json:"end_date"`
	CreatedAt         time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
	CreatedBy         int64      `gorm:"not null" json:"created_by"`
	UpdatedBy         int64      `gorm:"not null" json:"updated_by"`
}

func (Project) TableName() string {
	return "Project"
}
