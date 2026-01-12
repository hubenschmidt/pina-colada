package models

import "time"

type AutomationRunLog struct {
	ID                   int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	AutomationConfigID   int64      `gorm:"column:automation_config_id;not null" json:"automation_config_id"`
	StartedAt            time.Time  `gorm:"column:started_at;not null" json:"started_at"`
	CompletedAt          *time.Time `gorm:"column:completed_at" json:"completed_at"`
	Status               string     `gorm:"column:status;not null;default:running" json:"status"`
	ProspectsFound       int        `gorm:"column:prospects_found;not null;default:0" json:"prospects_found"`
	ProposalsCreated     int        `gorm:"column:proposals_created;not null;default:0" json:"proposals_created"`
	ErrorMessage         *string    `gorm:"column:error_message" json:"error_message"`
	ExecutedQuery        *string `gorm:"column:executed_query" json:"executed_query"`
	ExecutedSystemPrompt          *string `gorm:"column:executed_system_prompt" json:"executed_system_prompt"`
	ExecutedSystemPromptCharcount int     `gorm:"column:executed_system_prompt_charcount;not null;default:0" json:"executed_system_prompt_charcount"`
	Compiled                      bool    `gorm:"column:compiled;not null;default:false" json:"compiled"`
	QueryUpdated                  bool    `gorm:"column:query_updated;not null;default:false" json:"query_updated"`
	PromptUpdated                 bool    `gorm:"column:prompt_updated;not null;default:false" json:"prompt_updated"`
}

func (AutomationRunLog) TableName() string {
	return "Automation_Run_Log"
}
