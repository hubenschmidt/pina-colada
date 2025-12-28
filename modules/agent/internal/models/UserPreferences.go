package models

import "time"

type UserPreferences struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    int64     `gorm:"uniqueIndex;not null" json:"user_id"`
	Theme     *string   `json:"theme"` // null = inherit from tenant
	Timezone  *string   `json:"timezone"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (UserPreferences) TableName() string {
	return "User_Preferences"
}
