package models

import "time"

type EntityApprovalConfig struct {
	ID               int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID         int64     `gorm:"uniqueIndex:idx_approval_tenant_entity;not null" json:"tenant_id"`
	EntityType       string    `gorm:"uniqueIndex:idx_approval_tenant_entity;not null" json:"entity_type"`
	RequiresApproval bool      `gorm:"not null;default:true" json:"requires_approval"`
	CreatedAt        time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (EntityApprovalConfig) TableName() string {
	return "Entity_Approval_Config"
}
