package models

import (
	"time"

	"gorm.io/gorm"
)

// BaseModel contains common fields for all models
type BaseModel struct {
	ID        int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// AuditModel extends BaseModel with audit columns
type AuditModel struct {
	BaseModel
	CreatedByID int64 `gorm:"not null" json:"created_by_id"`
	UpdatedByID int64 `gorm:"not null" json:"updated_by_id"`
}

// TenantModel extends AuditModel with tenant isolation
type TenantModel struct {
	AuditModel
	TenantID int64 `gorm:"index;not null" json:"tenant_id"`
}
