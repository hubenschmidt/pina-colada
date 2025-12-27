package models

import "time"

type EntityAsset struct {
	AssetID    int64     `gorm:"primaryKey" json:"asset_id"`
	EntityType string    `gorm:"primaryKey" json:"entity_type"` // Project, Lead, Account, etc.
	EntityID   int64     `gorm:"primaryKey" json:"entity_id"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (EntityAsset) TableName() string {
	return "Entity_Asset"
}
