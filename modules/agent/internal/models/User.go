package models

import "time"

type User struct {
	ID                int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID          *int64     `gorm:"index" json:"tenant_id"`
	IndividualID      *int64     `json:"individual_id"`
	Auth0Sub          *string    `gorm:"uniqueIndex" json:"auth0_sub"`
	Email             string     `gorm:"not null" json:"email"`
	FirstName         *string    `json:"first_name"`
	LastName          *string    `json:"last_name"`
	AvatarURL         *string    `json:"avatar_url"`
	Status            string     `gorm:"not null;default:active" json:"status"` // active, inactive, invited
	IsSystemUser      bool       `gorm:"default:false" json:"is_system_user"`
	LastLoginAt       *time.Time `json:"last_login_at"`
	SelectedProjectID *int64     `json:"selected_project_id"`
	CreatedAt         time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

func (User) TableName() string {
	return "User"
}
