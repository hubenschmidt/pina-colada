package models

import "time"

type Status struct {
	ID          int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string    `gorm:"uniqueIndex;not null" json:"name"`
	Description *string   `json:"description"`
	Category    *string   `json:"category"` // job, lead, deal, task_status, task_priority
	IsTerminal  bool      `gorm:"default:false" json:"is_terminal"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (Status) TableName() string {
	return "Status"
}
