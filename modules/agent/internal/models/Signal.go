package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// Signal represents an intent signal on an Account
type Signal struct {
	ID             int64            `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	AccountID      int64            `gorm:"column:account_id;not null" json:"account_id"`
	SignalType     string           `gorm:"column:signal_type;not null" json:"signal_type"` // 'hiring', 'expansion', 'product_launch', 'partnership', 'leadership_change', 'news'
	Headline       string           `gorm:"column:headline;not null" json:"headline"`
	Description    *string          `gorm:"column:description" json:"description"`
	SignalDate     *time.Time       `gorm:"column:signal_date" json:"signal_date"`
	Source         *string          `gorm:"column:source" json:"source"`     // 'linkedin', 'news', 'crunchbase', 'agent'
	SourceURL      *string          `gorm:"column:source_url" json:"source_url"`
	Sentiment      *string          `gorm:"column:sentiment" json:"sentiment"` // 'positive', 'neutral', 'negative'
	RelevanceScore *decimal.Decimal `gorm:"column:relevance_score;type:numeric(3,2)" json:"relevance_score"`
	CreatedAt      time.Time        `gorm:"column:created_at;not null;default:now()" json:"created_at"`
	UpdatedAt      time.Time        `gorm:"column:updated_at;not null;default:now()" json:"updated_at"`

	// Relationships
	Account *Account `gorm:"foreignKey:AccountID" json:"account,omitempty"`
}

// TableName specifies the table name for Signal
func (Signal) TableName() string {
	return "Signal"
}
