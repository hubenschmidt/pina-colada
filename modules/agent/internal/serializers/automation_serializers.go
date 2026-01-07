package serializers

import "time"

// AutomationConfigResponse represents the automation config for API responses
type AutomationConfigResponse struct {
	ID                    int64      `json:"id,omitempty"`
	TenantID              int64      `json:"tenant_id"`
	UserID                int64      `json:"user_id"`
	Name                  string     `json:"name"`
	EntityType            string     `json:"entity_type"`
	Enabled               bool       `json:"enabled"`
	IntervalSeconds       int        `json:"interval_seconds"`
	LastRunAt             *time.Time `json:"last_run_at"`
	NextRunAt             *time.Time `json:"next_run_at"`
	RunCount              int        `json:"run_count"`
	ProspectsPerRun       int        `json:"prospects_per_run"`
	ConcurrentSearches    int        `json:"concurrent_searches"`
	CompilationTarget int  `json:"compilation_target"`
	DisableOnCompiled bool `json:"disable_on_compiled"`
	ActiveProposals   int  `json:"active_proposals"`
	CompiledAt        *time.Time `json:"compiled_at"`
	SystemPrompt       *string    `json:"system_prompt"`
	SearchSlots        [][]string `json:"search_slots"`
	ATSMode            bool       `json:"ats_mode"`
	TimeFilter         *string    `json:"time_filter"`
	TargetType         *string    `json:"target_type"`
	TargetIDs          []int64    `json:"target_ids"`
	SourceDocumentIDs  []int64    `json:"source_document_ids"`
	DigestEnabled      bool       `json:"digest_enabled"`
	DigestEmails       *string    `json:"digest_emails"`
	DigestTime         *string    `json:"digest_time"`
	DigestModel        *string    `json:"digest_model"`
	LastDigestAt       *time.Time `json:"last_digest_at"`
	UseAgent           bool       `json:"use_agent"`
	AgentModel         *string    `json:"agent_model"`
	CreatedAt          time.Time  `json:"created_at,omitempty"`
	UpdatedAt          time.Time  `json:"updated_at,omitempty"`
}

// AutomationRunLogResponse represents a run log entry for API responses
type AutomationRunLogResponse struct {
	ID               int64      `json:"id"`
	StartedAt        time.Time  `json:"started_at"`
	CompletedAt      *time.Time `json:"completed_at"`
	Status           string     `json:"status"`
	ProspectsFound   int        `json:"prospects_found"`
	ProposalsCreated int        `json:"proposals_created"`
	ErrorMessage     *string    `json:"error_message"`
	SearchQuery      *string    `json:"search_query"`
	Compiled bool `json:"compiled"`
	// ConfigActiveProposals is included in SSE events to update crawler stats
	ConfigActiveProposals *int `json:"config_active_proposals,omitempty"`
	// ConfigEnabled is included in SSE events when crawler is auto-paused
	ConfigEnabled *bool `json:"config_enabled,omitempty"`
}
