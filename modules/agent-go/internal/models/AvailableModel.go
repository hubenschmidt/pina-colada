package models

import "time"

type AvailableModel struct {
	ID                    int64     `gorm:"primaryKey" json:"id"`
	Provider              string    `gorm:"not null" json:"provider"`
	ModelName             string    `gorm:"not null" json:"model_name"`
	DisplayName           string    `json:"display_name"`
	SortOrder             int       `gorm:"default:0" json:"sort_order"`
	IsActive              bool      `gorm:"default:true" json:"is_active"`
	DefaultTimeoutSeconds int       `gorm:"default:20" json:"default_timeout_seconds"`
	CreatedAt             time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (AvailableModel) TableName() string {
	return "Available_Model"
}
