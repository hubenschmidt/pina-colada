package models

import "time"

type TenantPreferences struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID  int64     `gorm:"uniqueIndex;not null" json:"tenant_id"`
	Theme              string `gorm:"not null;default:light" json:"theme"`
	MinCrawlerInterval int    `gorm:"not null;default:10" json:"min_crawler_interval"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (TenantPreferences) TableName() string {
	return "Tenant_Preferences"
}
