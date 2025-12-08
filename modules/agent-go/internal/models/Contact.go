package models

import "time"

type Contact struct {
	ID           int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	IndividualID *int64    `gorm:"index" json:"individual_id"`
	FirstName    string    `gorm:"not null" json:"first_name"`
	LastName     string    `gorm:"not null" json:"last_name"`
	Title        *string   `json:"title"`
	Department   *string   `json:"department"`
	Role         *string   `json:"role"`
	Email        *string   `json:"email"`
	Phone        *string   `json:"phone"`
	IsPrimary    bool      `gorm:"default:false" json:"is_primary"`
	Notes        *string   `json:"notes"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	CreatedBy    int64     `gorm:"not null" json:"created_by"`
	UpdatedBy    int64     `gorm:"not null" json:"updated_by"`
}

func (Contact) TableName() string {
	return "Contact"
}

type ContactAccount struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	ContactID int64     `gorm:"index;not null" json:"contact_id"`
	AccountID int64     `gorm:"index;not null" json:"account_id"`
	IsPrimary bool      `gorm:"default:false" json:"is_primary"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (ContactAccount) TableName() string {
	return "Contact_Account"
}
