package schemas

// CrawlerRequest represents the request body for creating/updating a crawler
type CrawlerRequest struct {
	Name               *string  `json:"name,omitempty"`
	EntityType         *string  `json:"entity_type,omitempty"`
	Enabled            *bool    `json:"enabled,omitempty"`
	IntervalSeconds    *int     `json:"interval_seconds,omitempty"`
	CompilationTarget  *int     `json:"compilation_target,omitempty"`
	DisableOnCompiled  *bool    `json:"disable_on_compiled,omitempty"`
	SystemPrompt          *string `json:"system_prompt,omitempty"`
	UseSuggestedPrompt    *bool   `json:"use_suggested_prompt,omitempty"`
	SuggestionThreshold   *int    `json:"suggestion_threshold,omitempty"`
	MinProspectsThreshold *int    `json:"min_prospects_threshold,omitempty"`
	SearchQuery       *string `json:"search_query,omitempty"`
	UseSuggestedQuery *bool   `json:"use_suggested_query,omitempty"`
	ATSMode           *bool   `json:"ats_mode,omitempty"`
	TimeFilter         *string  `json:"time_filter,omitempty"`
	Location           *string  `json:"location,omitempty"`
	TargetType         *string  `json:"target_type,omitempty"`
	TargetIDs          []int64  `json:"target_ids,omitempty"`
	SourceDocumentIDs  []int64  `json:"source_document_ids,omitempty"`
	DigestEnabled      *bool    `json:"digest_enabled,omitempty"`
	DigestEmails       *string  `json:"digest_emails,omitempty"`
	DigestTime         *string  `json:"digest_time,omitempty"`
	DigestModel        *string  `json:"digest_model,omitempty"`
	UseAgent           *bool    `json:"use_agent,omitempty"`
	AgentModel         *string  `json:"agent_model,omitempty"`
	UseAnalytics       *bool    `json:"use_analytics,omitempty"`
	AnalyticsModel     *string  `json:"analytics_model,omitempty"`
	EmptyProposalLimit      *int    `json:"empty_proposal_limit,omitempty"`
	PromptCooldownRuns      *int    `json:"prompt_cooldown_runs,omitempty"`
	PromptCooldownProspects *int    `json:"prompt_cooldown_prospects,omitempty"`
	SuggestedQuery          *string `json:"suggested_query,omitempty"`
	SuggestedPrompt         *string `json:"suggested_prompt,omitempty"`
}

// ToggleRequest represents the request body for toggling a crawler
type ToggleRequest struct {
	Enabled bool `json:"enabled"`
}
