package models

import "time"

type AgentConfigPreset struct {
	ID               int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID           *int64    `gorm:"column:user_id" json:"user_id,omitempty"` // NULL = global preset
	Name             string    `gorm:"size:100;not null" json:"name"`
	Temperature      *float64  `gorm:"column:temperature" json:"temperature,omitempty"`
	MaxTokens        *int      `gorm:"column:max_tokens" json:"max_tokens,omitempty"`
	TopP             *float64  `gorm:"column:top_p" json:"top_p,omitempty"`
	TopK             *int      `gorm:"column:top_k" json:"top_k,omitempty"`
	FrequencyPenalty *float64  `gorm:"column:frequency_penalty" json:"frequency_penalty,omitempty"`
	PresencePenalty  *float64  `gorm:"column:presence_penalty" json:"presence_penalty,omitempty"`
	CreatedAt        time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (AgentConfigPreset) TableName() string {
	return "Agent_Config_Preset"
}
