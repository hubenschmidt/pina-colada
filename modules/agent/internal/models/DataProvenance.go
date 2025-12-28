package models

import (
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/datatypes"
)

type DataProvenance struct {
	ID         int64            `gorm:"primaryKey;autoIncrement" json:"id"`
	EntityType string           `gorm:"type:text;not null" json:"entity_type"`  // Organization, Individual
	EntityID   int64            `gorm:"not null" json:"entity_id"`
	FieldName  string           `gorm:"type:text;not null" json:"field_name"`   // revenue_range_id, seniority_level, etc.
	Source     string           `gorm:"type:text;not null" json:"source"`       // clearbit, apollo, linkedin, agent, manual
	SourceURL  *string          `gorm:"type:text" json:"source_url"`
	Confidence *decimal.Decimal `gorm:"type:numeric(3,2)" json:"confidence"`    // 0.00 to 1.00
	VerifiedAt time.Time        `gorm:"autoCreateTime" json:"verified_at"`
	VerifiedBy *int64           `json:"verified_by"`                            // NULL = AI agent
	RawValue   datatypes.JSON   `gorm:"type:jsonb" json:"raw_value"`
	CreatedAt  time.Time        `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time        `gorm:"autoUpdateTime" json:"updated_at"`
}

func (DataProvenance) TableName() string {
	return "Data_Provenance"
}
