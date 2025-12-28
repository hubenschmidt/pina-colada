package models

import "time"

type RevenueRange struct {
	ID           int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Label        string    `gorm:"uniqueIndex;not null" json:"label"`
	MinValue     *int64    `json:"min_value"` // USD
	MaxValue     *int64    `json:"max_value"` // USD
	DisplayOrder int       `gorm:"not null;default:0" json:"display_order"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (RevenueRange) TableName() string {
	return "Revenue_Range"
}
