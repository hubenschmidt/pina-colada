package models

import "time"

type Asset struct {
	ID               int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	AssetType        string    `gorm:"not null;index" json:"asset_type"` // document, image
	TenantID         int64     `gorm:"index;not null" json:"tenant_id"`
	UserID           int64     `gorm:"index;not null" json:"user_id"`
	Filename         string    `gorm:"not null" json:"filename"`
	ContentType      string    `gorm:"not null" json:"content_type"` // MIME type
	Description      *string   `json:"description"`
	CreatedAt        time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	CreatedBy        int64     `gorm:"not null" json:"created_by"`
	UpdatedBy        int64     `gorm:"not null" json:"updated_by"`
	ParentID         *int64    `gorm:"index" json:"parent_id"`
	VersionNumber    int       `gorm:"not null;default:1" json:"version_number"`
	IsCurrentVersion bool      `gorm:"not null;default:true" json:"is_current_version"`
}

func (Asset) TableName() string {
	return "Asset"
}
