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
	IntervalMinutes       int        `json:"interval_minutes"`
	LastRunAt             *time.Time `json:"last_run_at"`
	NextRunAt             *time.Time `json:"next_run_at"`
	RunCount              int        `json:"run_count"`
	LeadsPerRun           int        `json:"leads_per_run"`
	ConcurrentSearches    int        `json:"concurrent_searches"`
	CompilationTarget     int        `json:"compilation_target"`
	TotalProposalsCreated int        `json:"total_proposals_created"`
	SystemPrompt       *string    `json:"system_prompt"`
	SearchKeywords     []string   `json:"search_keywords"`
	SearchSlots        [][]int    `json:"search_slots"`
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
	LeadsFound       int        `json:"leads_found"`
	ProposalsCreated int        `json:"proposals_created"`
	ErrorMessage     *string    `json:"error_message"`
	SearchQuery      *string    `json:"search_query"`
}
