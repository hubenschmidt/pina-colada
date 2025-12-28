package models

import "time"

type Industry struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string    `gorm:"uniqueIndex;not null" json:"name"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (Industry) TableName() string {
	return "Industry"
}

type AccountIndustry struct {
	AccountID  int64     `gorm:"primaryKey" json:"account_id"`
	IndustryID int64     `gorm:"primaryKey" json:"industry_id"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (AccountIndustry) TableName() string {
	return "Account_Industry"
}
