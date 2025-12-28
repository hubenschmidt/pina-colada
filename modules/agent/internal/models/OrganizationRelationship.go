package models

import "time"

type OrganizationRelationship struct {
	ID                 int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	FromOrganizationID int64     `gorm:"not null;index" json:"from_organization_id"`
	ToOrganizationID   int64     `gorm:"not null;index" json:"to_organization_id"`
	RelationshipType   *string   `json:"relationship_type"`
	Notes              *string   `gorm:"type:text" json:"notes"`
	CreatedAt          time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt          time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (OrganizationRelationship) TableName() string {
	return "Organization_Relationship"
}
