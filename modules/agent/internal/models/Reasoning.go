package models

import (
	"time"

	"gorm.io/datatypes"
)

type Reasoning struct {
	ID          int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	Type        string         `gorm:"type:text;not null" json:"type"` // crm, finance, etc.
	TableName_  string         `gorm:"column:table_name;type:text;not null" json:"table_name"`
	Description *string        `gorm:"type:text" json:"description"`
	SchemaHint  datatypes.JSON `gorm:"type:jsonb" json:"schema_hint"`
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at"`
}

func (Reasoning) TableName() string {
	return "Reasoning"
}
