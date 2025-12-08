package models

import (
	"time"

	"github.com/shopspring/decimal"
)

type Deal struct {
	ID                int64            `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID          *int64           `gorm:"index" json:"tenant_id"`
	ProjectID         *int64           `gorm:"index" json:"project_id"`
	Name              string           `gorm:"not null" json:"name"`
	Description       *string          `json:"description"`
	OwnerIndividualID *int64           `json:"owner_individual_id"`
	CurrentStatusID   *int64           `json:"current_status_id"`
	ValueAmount       *decimal.Decimal `gorm:"type:numeric(18,2)" json:"value_amount"`
	ValueCurrency     string           `gorm:"default:USD" json:"value_currency"`
	Probability       *int16           `json:"probability"` // 0-100
	ExpectedCloseDate *time.Time       `json:"expected_close_date"`
	CloseDate         *time.Time       `json:"close_date"`
	CreatedAt         time.Time        `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time        `gorm:"autoUpdateTime" json:"updated_at"`
	CreatedBy         int64            `gorm:"not null" json:"created_by"`
	UpdatedBy         int64            `gorm:"not null" json:"updated_by"`
}

func (Deal) TableName() string {
	return "Deal"
}
