package models

import "time"

type Tenant struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string    `gorm:"not null" json:"name"`
	Slug      string    `gorm:"uniqueIndex;not null" json:"slug"`
	Status    string    `gorm:"not null;default:active" json:"status"` // active, suspended, trial, cancelled
	Plan      string    `gorm:"not null;default:free" json:"plan"`     // free, starter, professional, enterprise
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (Tenant) TableName() string {
	return "Tenant"
}
