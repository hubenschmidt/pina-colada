package models

import "time"

type FundingStage struct {
	ID           int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Label        string    `gorm:"uniqueIndex;not null" json:"label"`
	DisplayOrder int       `gorm:"not null;default:0" json:"display_order"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (FundingStage) TableName() string {
	return "Funding_Stage"
}
