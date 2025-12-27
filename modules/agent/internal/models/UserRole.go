package models

import "time"

type UserRole struct {
	UserID    int64     `gorm:"primaryKey" json:"user_id"`
	RoleID    int64     `gorm:"primaryKey" json:"role_id"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (UserRole) TableName() string {
	return "User_Role"
}
