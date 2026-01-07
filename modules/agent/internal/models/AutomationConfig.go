package models

import (
	"time"

	"gorm.io/datatypes"
)

type AutomationConfig struct {
	ID                 int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID           int64          `gorm:"column:tenant_id;not null" json:"tenant_id"`
	UserID             int64          `gorm:"column:user_id;not null" json:"user_id"`
	Name               string         `gorm:"column:name;not null" json:"name"`
	EntityType         string         `gorm:"column:entity_type;not null;default:'job'" json:"entity_type"`
	Enabled            bool           `gorm:"column:enabled;not null;default:false" json:"enabled"`
	IntervalMinutes    int            `gorm:"column:interval_minutes;not null;default:30" json:"interval_minutes"`
	LastRunAt          *time.Time     `gorm:"column:last_run_at" json:"last_run_at"`
	NextRunAt          *time.Time     `gorm:"column:next_run_at" json:"next_run_at"`
	RunCount           int            `gorm:"column:run_count;not null;default:0" json:"run_count"`
	LeadsPerRun        int            `gorm:"column:leads_per_run;not null;default:10" json:"leads_per_run"`
	ConcurrentSearches int            `gorm:"column:concurrent_searches;not null;default:1" json:"concurrent_searches"`
	CompilationTarget  int            `gorm:"column:compilation_target;not null;default:100" json:"compilation_target"`
	SystemPrompt       *string        `gorm:"column:system_prompt" json:"system_prompt"`
	SearchKeywords     datatypes.JSON `gorm:"column:search_keywords;type:jsonb" json:"search_keywords"`
	SearchSlots        datatypes.JSON `gorm:"column:search_slots;type:jsonb" json:"search_slots"`
	ATSMode            bool           `gorm:"column:ats_mode;not null;default:true" json:"ats_mode"`
	TimeFilter         *string        `gorm:"column:time_filter" json:"time_filter"`
	TargetType        *string        `gorm:"column:target_type" json:"target_type"`
	TargetIDs         datatypes.JSON `gorm:"column:target_ids;type:jsonb" json:"target_ids"`
	SourceDocumentIDs datatypes.JSON `gorm:"column:source_document_ids;type:jsonb" json:"source_document_ids"`
	DigestEnabled      bool           `gorm:"column:digest_enabled;not null;default:true" json:"digest_enabled"`
	DigestEmails       *string        `gorm:"column:digest_emails" json:"digest_emails"`
	DigestTime         *string        `gorm:"column:digest_time;default:'09:00'" json:"digest_time"`
	DigestModel        *string        `gorm:"column:digest_model" json:"digest_model"`
	LastDigestAt       *time.Time     `gorm:"column:last_digest_at" json:"last_digest_at"`
	UseAgent           bool           `gorm:"column:use_agent;not null;default:false" json:"use_agent"`
	AgentModel         *string        `gorm:"column:agent_model" json:"agent_model"`
	CreatedAt          time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt          time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
}

func (AutomationConfig) TableName() string {
	return "Automation_Config"
}
