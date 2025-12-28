package models

import "time"

type Note struct {
	ID         int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID   int64     `gorm:"not null" json:"tenant_id"`
	EntityType string    `gorm:"size:50;not null" json:"entity_type"`
	EntityID   int64     `gorm:"not null" json:"entity_id"`
	Content    string    `gorm:"not null" json:"content"`
	CreatedBy  int64     `gorm:"not null" json:"created_by"`
	UpdatedBy  int64     `gorm:"not null" json:"updated_by"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (Note) TableName() string {
	return "Note"
}
