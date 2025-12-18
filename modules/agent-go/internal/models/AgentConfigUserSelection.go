package models

import "time"

type AgentConfigUserSelection struct {
	UserID                int64     `gorm:"primaryKey" json:"user_id"`
	SelectedParamPresetID *int64    `gorm:"column:selected_param_preset_id" json:"selected_param_preset_id,omitempty"`
	SelectedCostTier      string    `gorm:"column:selected_cost_tier;default:premium" json:"selected_cost_tier"`
	SelectedProvider      string    `gorm:"column:selected_provider;default:openai" json:"selected_provider"`
	UpdatedAt             time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (AgentConfigUserSelection) TableName() string {
	return "Agent_Config_User_Selection"
}
