package models

import "time"

type Account struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID  *int64    `gorm:"index" json:"tenant_id"`
	Name      string    `gorm:"not null" json:"name"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	CreatedBy int64     `gorm:"not null" json:"created_by"`
	UpdatedBy int64     `gorm:"not null" json:"updated_by"`

	// Relations
	Organizations         []Organization        `gorm:"foreignKey:AccountID" json:"organizations,omitempty"`
	Individuals           []Individual          `gorm:"foreignKey:AccountID" json:"individuals,omitempty"`
	Industries            []Industry            `gorm:"many2many:Account_Industry;joinForeignKey:account_id;joinReferences:industry_id" json:"industries,omitempty"`
	Contacts              []Contact             `gorm:"many2many:Contact_Account;joinForeignKey:account_id;joinReferences:contact_id" json:"contacts,omitempty"`
	OutgoingRelationships []AccountRelationship `gorm:"foreignKey:FromAccountID" json:"outgoing_relationships,omitempty"`
	IncomingRelationships []AccountRelationship `gorm:"foreignKey:ToAccountID" json:"incoming_relationships,omitempty"`
}

func (Account) TableName() string {
	return "Account"
}
