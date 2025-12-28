package models

import "time"

type Role struct {
	ID          int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID    *int64    `gorm:"index" json:"tenant_id"` // nullable for system roles
	Name        string    `gorm:"not null" json:"name"`
	Description *string   `json:"description"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (Role) TableName() string {
	return "Role"
}
