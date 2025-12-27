package models

import "time"

type Tag struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID  *int64    `gorm:"index" json:"tenant_id"`
	Name      string    `gorm:"size:100;not null" json:"name"` // unique per tenant
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (Tag) TableName() string {
	return "Tag"
}

type EntityTag struct {
	TagID      int64     `gorm:"primaryKey" json:"tag_id"`
	EntityType string    `gorm:"primaryKey;type:text" json:"entity_type"`
	EntityID   int64     `gorm:"primaryKey" json:"entity_id"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (EntityTag) TableName() string {
	return "Entity_Tag"
}
