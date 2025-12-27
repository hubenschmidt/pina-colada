package models

import "time"

type AccountProject struct {
	AccountID int64     `gorm:"primaryKey" json:"account_id"`
	ProjectID int64     `gorm:"primaryKey" json:"project_id"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (AccountProject) TableName() string {
	return "Account_Project"
}
