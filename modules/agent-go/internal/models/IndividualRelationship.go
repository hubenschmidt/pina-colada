package models

import "time"

type IndividualRelationship struct {
	ID               int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	FromIndividualID int64     `gorm:"not null;index" json:"from_individual_id"`
	ToIndividualID   int64     `gorm:"not null;index" json:"to_individual_id"`
	RelationshipType *string   `json:"relationship_type"`
	Notes            *string   `gorm:"type:text" json:"notes"`
	CreatedAt        time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (IndividualRelationship) TableName() string {
	return "Individual_Relationship"
}
