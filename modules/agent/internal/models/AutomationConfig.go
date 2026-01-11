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
	IntervalSeconds    int            `gorm:"column:interval_seconds;not null;default:1800" json:"interval_seconds"`
	LastRunAt          *time.Time     `gorm:"column:last_run_at" json:"last_run_at"`
	NextRunAt          *time.Time     `gorm:"column:next_run_at" json:"next_run_at"`
	RunCount           int            `gorm:"column:run_count;not null;default:0" json:"run_count"`
	ProspectsPerRun    int            `gorm:"column:prospects_per_run;not null;default:10" json:"prospects_per_run"`
	CompilationTarget  int            `gorm:"column:compilation_target;not null;default:100" json:"compilation_target"`
	DisableOnCompiled  bool           `gorm:"column:disable_on_compiled;not null;default:false" json:"disable_on_compiled"`
	CompiledAt         *time.Time     `gorm:"column:compiled_at" json:"compiled_at"`
	SystemPrompt       *string        `gorm:"column:system_prompt" json:"system_prompt"`
	SuggestedPrompt    *string        `gorm:"column:suggested_prompt" json:"suggested_prompt"`
	UseSuggestedPrompt  bool           `gorm:"column:use_suggested_prompt;not null;default:false" json:"use_suggested_prompt"`
	SuggestionThreshold int     `gorm:"column:suggestion_threshold;not null;default:50" json:"suggestion_threshold"`
	SearchQuery       *string `gorm:"column:search_query" json:"search_query"`
	SuggestedQuery    *string `gorm:"column:suggested_query" json:"suggested_query"`
	UseSuggestedQuery bool    `gorm:"column:use_suggested_query;not null;default:false" json:"use_suggested_query"`
	Excluded          *string `gorm:"column:excluded" json:"excluded"`
	ATSMode            bool           `gorm:"column:ats_mode;not null;default:true" json:"ats_mode"`
	TimeFilter         *string        `gorm:"column:time_filter" json:"time_filter"`
	Location           *string        `gorm:"column:location" json:"location"`
	TargetType        *string        `gorm:"column:target_type" json:"target_type"`
	TargetIDs         datatypes.JSON `gorm:"column:target_ids;type:jsonb" json:"target_ids"`
	SourceDocumentIDs datatypes.JSON `gorm:"column:source_document_ids;type:jsonb" json:"source_document_ids"`
	DigestEnabled      bool           `gorm:"column:digest_enabled;not null;default:true" json:"digest_enabled"`
	DigestEmails       *string        `gorm:"column:digest_emails" json:"digest_emails"`
	DigestTime         *string        `gorm:"column:digest_time;default:'09:00'" json:"digest_time"`
	DigestModel        *string        `gorm:"column:digest_model" json:"digest_model"`
	LastDigestAt       *time.Time     `gorm:"column:last_digest_at" json:"last_digest_at"`
	UseAgent           bool           `gorm:"column:use_agent;not null;default:false" json:"use_agent"`
	AgentModel           *string        `gorm:"column:agent_model" json:"agent_model"`
	ConsecutiveZeroRuns  int            `gorm:"column:consecutive_zero_runs;not null;default:0" json:"consecutive_zero_runs"`
	EmptyProposalLimit   int            `gorm:"column:empty_proposal_limit;not null;default:0" json:"empty_proposal_limit"`
	CreatedAt            time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt            time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
}

func (AutomationConfig) TableName() string {
	return "Automation_Config"
}
