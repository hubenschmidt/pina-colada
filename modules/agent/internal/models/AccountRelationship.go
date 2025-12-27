package models

import "time"

type AccountRelationship struct {
	ID               int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	FromAccountID    int64     `gorm:"not null;index" json:"from_account_id"`
	ToAccountID      int64     `gorm:"not null;index" json:"to_account_id"`
	RelationshipType *string   `json:"relationship_type"`
	Notes            *string   `gorm:"type:text" json:"notes"`
	CreatedAt        time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	CreatedBy        int64     `gorm:"not null" json:"created_by"`
	UpdatedBy        int64     `gorm:"not null" json:"updated_by"`

	// Relations
	FromAccount *Account `gorm:"foreignKey:FromAccountID" json:"from_account,omitempty"`
	ToAccount   *Account `gorm:"foreignKey:ToAccountID" json:"to_account,omitempty"`
}

func (AccountRelationship) TableName() string {
	return "Account_Relationship"
}
