package models

import (
	"time"

	"gorm.io/datatypes"
)

type ResearchCache struct {
	ID          int64          `gorm:"primaryKey" json:"id"`
	TenantID    int64          `gorm:"not null" json:"tenant_id"`
	CacheKey    string         `gorm:"not null;size:255" json:"cache_key"`
	CacheType   string         `gorm:"not null;size:50" json:"cache_type"`
	QueryParams datatypes.JSON `json:"query_params,omitempty"`
	ResultData  datatypes.JSON `gorm:"not null" json:"result_data"`
	ResultCount int            `gorm:"not null;default:0" json:"result_count"`
	CreatedBy   *int64         `json:"created_by,omitempty"`
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at"`
	ExpiresAt   time.Time      `gorm:"not null" json:"expires_at"`
}

func (ResearchCache) TableName() string {
	return "Research_Cache"
}
