package models

import "time"

type Account struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID  *int64    `gorm:"index" json:"tenant_id"`
	Name      string    `gorm:"not null" json:"name"`
	Type      *string   `json:"type"` // prospect, customer, partner, etc.
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	CreatedBy int64     `gorm:"not null" json:"created_by"`
	UpdatedBy int64     `gorm:"not null" json:"updated_by"`

	// Relations
	Organizations []Organization `gorm:"foreignKey:AccountID" json:"organizations,omitempty"`
	Individuals   []Individual   `gorm:"foreignKey:AccountID" json:"individuals,omitempty"`
}

func (Account) TableName() string {
	return "Account"
}
