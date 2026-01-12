package repositories

import (
	"encoding/json"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"agent/internal/models"
)

// AutomationRepository handles automation config data access
type AutomationRepository struct {
	db *gorm.DB
}

// NewAutomationRepository creates a new automation repository
func NewAutomationRepository(db *gorm.DB) *AutomationRepository {
	return &AutomationRepository{db: db}
}

// AutomationConfigDTO represents automation config for service layer
type AutomationConfigDTO struct {
	ID                 int64
	TenantID           int64
	UserID             int64
	Name               string
	EntityType         string
	Enabled            bool
	IntervalSeconds    int
	LastRunAt          *time.Time
	NextRunAt          *time.Time
	RunCount          int
	CompilationTarget int
	DisableOnCompiled   bool
	SystemPrompt        *string
	SuggestedPrompt       *string
	UseSuggestedPrompt    bool
	SuggestionThreshold   int
	MinProspectsThreshold int
	SearchQuery       *string
	SuggestedQuery    *string
	UseSuggestedQuery bool
	ATSMode           bool
	TimeFilter         *string
	Location           *string
	TargetType         *string
	TargetIDs          []int64
	SourceDocumentIDs  []int64
	DigestEnabled      bool
	DigestEmails       *string
	DigestTime         *string
	DigestModel        *string
	LastDigestAt       *time.Time
	UseAgent            bool
	AgentModel          *string
	UseAnalytics        bool
	AnalyticsModel      *string
	CompiledAt              *time.Time
	EmptyProposalLimit      int
	PromptCooldownRuns      int
	PromptCooldownProspects int
	CreatedAt               time.Time
	UpdatedAt               time.Time
}

// AutomationConfigInput for creating/updating config
type AutomationConfigInput struct {
	Name               *string
	EntityType         *string
	Enabled            *bool
	IntervalSeconds    *int
	CompilationTarget   *int
	DisableOnCompiled   *bool
	SystemPrompt          *string
	SuggestedPrompt       *string
	UseSuggestedPrompt    *bool
	SuggestionThreshold   *int
	MinProspectsThreshold *int
	SearchQuery       *string
	SuggestedQuery    *string
	UseSuggestedQuery *bool
	ATSMode           *bool
	TimeFilter         *string
	Location           *string
	TargetType         *string
	TargetIDs          []int64
	SourceDocumentIDs  []int64
	DigestEnabled      *bool
	DigestEmails       *string
	DigestTime         *string
	DigestModel        *string
	UseAgent           *bool
	AgentModel           *string
	UseAnalytics            *bool
	AnalyticsModel          *string
	EmptyProposalLimit      *int
	PromptCooldownRuns      *int
	PromptCooldownProspects *int
}

// AutomationRunLogDTO represents a run log entry
type AutomationRunLogDTO struct {
	ID                            int64
	StartedAt                     time.Time
	CompletedAt                   *time.Time
	Status                        string
	ProspectsFound                int
	ProposalsCreated              int
	ErrorMessage                  *string
	ExecutedQuery                 *string
	ExecutedSystemPrompt          *string
	ExecutedSystemPromptCharcount int
	Compiled                      bool
	QueryUpdated                  bool
	PromptUpdated                 bool
}

// GetTenantConfigs returns all automation configs for a tenant
func (r *AutomationRepository) GetTenantConfigs(tenantID int64) ([]AutomationConfigDTO, error) {
	var configs []models.AutomationConfig
	err := r.db.Where("tenant_id = ?", tenantID).
		Order("created_at ASC").
		Find(&configs).Error
	if err != nil {
		return nil, err
	}

	result := make([]AutomationConfigDTO, len(configs))
	for i := range configs {
		result[i] = *r.modelToDTO(&configs[i])
	}
	return result, nil
}

// GetConfigByID returns a specific automation config by ID
func (r *AutomationRepository) GetConfigByID(configID int64) (*AutomationConfigDTO, error) {
	var cfg models.AutomationConfig
	err := r.db.Where("id = ?", configID).First(&cfg).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return r.modelToDTO(&cfg), nil
}

// CreateConfig creates a new automation config
func (r *AutomationRepository) CreateConfig(tenantID, userID int64, input AutomationConfigInput) (*AutomationConfigDTO, error) {
	name := "Untitled Crawler"
	if input.Name != nil {
		name = *input.Name
	}
	entityType := "job"
	if input.EntityType != nil {
		entityType = *input.EntityType
	}

	cfg := models.AutomationConfig{
		TenantID:   tenantID,
		UserID:     userID,
		Name:       name,
		EntityType: entityType,
	}

	r.applyInput(&cfg, input)

	if err := r.db.Create(&cfg).Error; err != nil {
		return nil, err
	}

	return r.modelToDTO(&cfg), nil
}

// UpdateConfig updates an existing automation config
func (r *AutomationRepository) UpdateConfig(configID int64, input AutomationConfigInput) (*AutomationConfigDTO, error) {
	var cfg models.AutomationConfig
	err := r.db.Where("id = ?", configID).First(&cfg).Error
	if err != nil {
		return nil, err
	}

	r.applyInput(&cfg, input)

	if err := r.db.Save(&cfg).Error; err != nil {
		return nil, err
	}

	return r.modelToDTO(&cfg), nil
}

// DeleteConfig deletes an automation config
func (r *AutomationRepository) DeleteConfig(configID int64) error {
	return r.db.Delete(&models.AutomationConfig{}, configID).Error
}

// GetPausedCrawlers returns crawlers that are paused (enabled, compiled, not disable_on_compiled)
func (r *AutomationRepository) GetPausedCrawlers() ([]AutomationConfigDTO, error) {
	var configs []models.AutomationConfig
	err := r.db.Where("enabled = ? AND compiled_at IS NOT NULL AND disable_on_compiled = ?", true, false).Find(&configs).Error
	if err != nil {
		return nil, err
	}

	result := make([]AutomationConfigDTO, len(configs))
	for i := range configs {
		result[i] = *r.modelToDTO(&configs[i])
	}
	return result, nil
}

// GetDueAutomations returns enabled automations that are due to run.
// Uses row locking to prevent race conditions with concurrent scheduler ticks.
func (r *AutomationRepository) GetDueAutomations(now time.Time) ([]AutomationConfigDTO, error) {
	var configs []models.AutomationConfig

	err := r.db.Transaction(func(tx *gorm.DB) error {
		// Select due configs with row lock (SKIP LOCKED prevents blocking)
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
			Where("enabled = ? AND (next_run_at IS NULL OR next_run_at <= ?)", true, now).
			Find(&configs).Error; err != nil {
			return err
		}

		// Claim configs by setting next_run_at to prevent re-pickup
		for i := range configs {
			nextRun := now.Add(time.Duration(configs[i].IntervalSeconds) * time.Second)
			if err := tx.Model(&configs[i]).Update("next_run_at", nextRun).Error; err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	result := make([]AutomationConfigDTO, len(configs))
	for i := range configs {
		result[i] = *r.modelToDTO(&configs[i])
	}
	return result, nil
}

// UpdateLastRun updates the last run time and increments run count
func (r *AutomationRepository) UpdateLastRun(configID int64, lastRun, nextRun time.Time) error {
	return r.db.Model(&models.AutomationConfig{}).
		Where("id = ?", configID).
		Updates(map[string]interface{}{
			"last_run_at": lastRun,
			"next_run_at": nextRun,
			"run_count":   gorm.Expr("run_count + 1"),
			"updated_at":  time.Now(),
		}).Error
}

// SetNextRun sets the next run time without affecting last_run or run_count
func (r *AutomationRepository) SetNextRun(configID int64, nextRun time.Time) error {
	return r.db.Model(&models.AutomationConfig{}).
		Where("id = ?", configID).
		Updates(map[string]interface{}{
			"next_run_at": nextRun,
			"updated_at":  time.Now(),
		}).Error
}

// GetDigestDueConfigs returns configs that need digest emails sent
// Returns configs where digest is enabled and last_digest_at was before today
func (r *AutomationRepository) GetDigestDueConfigs(now time.Time) ([]AutomationConfigDTO, error) {
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	var configs []models.AutomationConfig
	err := r.db.Where("digest_enabled = ? AND (last_digest_at IS NULL OR last_digest_at < ?)",
		true, startOfDay).Find(&configs).Error
	if err != nil {
		return nil, err
	}

	result := make([]AutomationConfigDTO, len(configs))
	for i := range configs {
		result[i] = *r.modelToDTO(&configs[i])
	}
	return result, nil
}

// UpdateLastDigest updates the last digest timestamp
func (r *AutomationRepository) UpdateLastDigest(configID int64, digestTime time.Time) error {
	return r.db.Model(&models.AutomationConfig{}).
		Where("id = ?", configID).
		Updates(map[string]interface{}{
			"last_digest_at": digestTime,
			"updated_at":     time.Now(),
		}).Error
}

// SetCompiledAt marks when the compilation target was reached
func (r *AutomationRepository) SetCompiledAt(configID int64, compiledAt time.Time) error {
	return r.db.Model(&models.AutomationConfig{}).
		Where("id = ?", configID).
		Updates(map[string]interface{}{
			"compiled_at": compiledAt,
			"updated_at":  time.Now(),
		}).Error
}

// ClearCompiledAt clears the compiled_at timestamp (when proposals drop below target)
func (r *AutomationRepository) ClearCompiledAt(configID int64) error {
	return r.db.Model(&models.AutomationConfig{}).
		Where("id = ?", configID).
		Updates(map[string]interface{}{
			"compiled_at": nil,
			"updated_at":  time.Now(),
		}).Error
}

// DisableConfig sets enabled=false on the config
func (r *AutomationRepository) DisableConfig(configID int64) error {
	return r.db.Model(&models.AutomationConfig{}).
		Where("id = ?", configID).
		Updates(map[string]interface{}{
			"enabled":    false,
			"updated_at": time.Now(),
		}).Error
}

// CreateRunLog creates a new run log entry
func (r *AutomationRepository) CreateRunLog(configID int64, executedQuery string, executedSystemPrompt *string) (int64, error) {
	promptCharcount := 0
	if executedSystemPrompt != nil {
		promptCharcount = len(*executedSystemPrompt)
	}
	log := models.AutomationRunLog{
		AutomationConfigID:            configID,
		StartedAt:                     time.Now(),
		Status:                        "running",
		ExecutedQuery:                 &executedQuery,
		ExecutedSystemPrompt:          executedSystemPrompt,
		ExecutedSystemPromptCharcount: promptCharcount,
	}
	err := r.db.Create(&log).Error
	return log.ID, err
}

// CompleteRunLog marks a run log as completed
func (r *AutomationRepository) CompleteRunLog(logID int64, status string, prospectsFound, proposalsCreated int, errorMsg *string, compiled, queryUpdated, promptUpdated bool) error {
	now := time.Now()
	return r.db.Model(&models.AutomationRunLog{}).
		Where("id = ?", logID).
		Updates(map[string]interface{}{
			"completed_at":      now,
			"status":            status,
			"prospects_found":   prospectsFound,
			"proposals_created": proposalsCreated,
			"error_message":     errorMsg,
			"compiled":          compiled,
			"query_updated":     queryUpdated,
			"prompt_updated":    promptUpdated,
		}).Error
}

// GetRunLogs returns paginated run logs for a config
func (r *AutomationRepository) GetRunLogs(configID int64, page, limit int) ([]AutomationRunLogDTO, int64, error) {
	var totalCount int64
	if err := r.db.Model(&models.AutomationRunLog{}).
		Where("automation_config_id = ?", configID).
		Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	var logs []models.AutomationRunLog
	offset := (page - 1) * limit
	err := r.db.Where("automation_config_id = ?", configID).
		Order("started_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&logs).Error
	if err != nil {
		return nil, 0, err
	}

	result := make([]AutomationRunLogDTO, len(logs))
	for i := range logs {
		result[i] = AutomationRunLogDTO{
			ID:                            logs[i].ID,
			StartedAt:                     logs[i].StartedAt,
			CompletedAt:                   logs[i].CompletedAt,
			Status:                        logs[i].Status,
			ProspectsFound:                logs[i].ProspectsFound,
			ProposalsCreated:              logs[i].ProposalsCreated,
			ErrorMessage:                  logs[i].ErrorMessage,
			ExecutedQuery:                 logs[i].ExecutedQuery,
			ExecutedSystemPrompt:          logs[i].ExecutedSystemPrompt,
			ExecutedSystemPromptCharcount: logs[i].ExecutedSystemPromptCharcount,
			Compiled:                      logs[i].Compiled,
			QueryUpdated:                  logs[i].QueryUpdated,
			PromptUpdated:                 logs[i].PromptUpdated,
		}
	}
	return result, totalCount, nil
}

// GetActiveProposals returns count of pending proposals for a config
func (r *AutomationRepository) GetActiveProposals(configID int64) (int, error) {
	var total int64
	err := r.db.Model(&models.AgentProposal{}).
		Where("automation_config_id = ? AND status = ?", configID, "pending").
		Count(&total).Error
	return int(total), err
}

// HasRunningRun checks if a config has a run currently in progress
func (r *AutomationRepository) HasRunningRun(configID int64) bool {
	var count int64
	r.db.Model(&models.AutomationRunLog{}).
		Where("automation_config_id = ? AND status = ?", configID, "running").
		Count(&count)
	return count > 0
}

// GetLastCompletedRun returns the most recent completed run for a config
func (r *AutomationRepository) GetLastCompletedRun(configID int64) (*AutomationRunLogDTO, error) {
	var log models.AutomationRunLog
	err := r.db.Where("automation_config_id = ? AND status = ?", configID, "done").
		Order("started_at DESC").
		First(&log).Error
	if err != nil {
		return nil, err
	}
	return &AutomationRunLogDTO{
		ID:                   log.ID,
		ExecutedQuery:        log.ExecutedQuery,
		ExecutedSystemPrompt: log.ExecutedSystemPrompt,
	}, nil
}

// CleanupStaleRuns marks any "running" jobs as failed (called on startup)
func (r *AutomationRepository) CleanupStaleRuns() (int64, error) {
	result := r.db.Model(&models.AutomationRunLog{}).
		Where("status = ?", "running").
		Updates(map[string]interface{}{
			"status":        "failed",
			"completed_at":  time.Now(),
			"error_message": "Server restarted while run was in progress",
		})
	return result.RowsAffected, result.Error
}

// IncrementZeroRuns increments consecutive_zero_runs and returns the new count
func (r *AutomationRepository) IncrementZeroRuns(configID int64) (int, error) {
	err := r.db.Model(&models.AutomationConfig{}).
		Where("id = ?", configID).
		Update("consecutive_zero_runs", gorm.Expr("consecutive_zero_runs + 1")).Error
	if err != nil {
		return 0, err
	}
	var count int
	r.db.Model(&models.AutomationConfig{}).Where("id = ?", configID).
		Pluck("consecutive_zero_runs", &count)
	return count, nil
}

// ResetZeroRuns resets consecutive_zero_runs to 0
func (r *AutomationRepository) ResetZeroRuns(configID int64) error {
	return r.db.Model(&models.AutomationConfig{}).
		Where("id = ?", configID).
		Update("consecutive_zero_runs", 0).Error
}

func (r *AutomationRepository) modelToDTO(cfg *models.AutomationConfig) *AutomationConfigDTO {
	dto := &AutomationConfigDTO{
		ID:                 cfg.ID,
		TenantID:           cfg.TenantID,
		UserID:             cfg.UserID,
		Name:               cfg.Name,
		EntityType:         cfg.EntityType,
		Enabled:            cfg.Enabled,
		IntervalSeconds:    cfg.IntervalSeconds,
		LastRunAt:          cfg.LastRunAt,
		NextRunAt:          cfg.NextRunAt,
		RunCount:          cfg.RunCount,
		CompilationTarget: cfg.CompilationTarget,
		DisableOnCompiled:  cfg.DisableOnCompiled,
		SystemPrompt:       cfg.SystemPrompt,
		SuggestedPrompt:       cfg.SuggestedPrompt,
		UseSuggestedPrompt:    cfg.UseSuggestedPrompt,
		SuggestionThreshold:   cfg.SuggestionThreshold,
		MinProspectsThreshold: cfg.MinProspectsThreshold,
		SearchQuery:       cfg.SearchQuery,
		SuggestedQuery:    cfg.SuggestedQuery,
		UseSuggestedQuery: cfg.UseSuggestedQuery,
		ATSMode:           cfg.ATSMode,
		TimeFilter:         cfg.TimeFilter,
		Location:           cfg.Location,
		TargetType:         cfg.TargetType,
		DigestEnabled:      cfg.DigestEnabled,
		DigestEmails:       cfg.DigestEmails,
		DigestTime:         cfg.DigestTime,
		DigestModel:        cfg.DigestModel,
		LastDigestAt:       cfg.LastDigestAt,
		UseAgent:           cfg.UseAgent,
		AgentModel:         cfg.AgentModel,
		UseAnalytics:       cfg.UseAnalytics,
		AnalyticsModel:     cfg.AnalyticsModel,
		CompiledAt:              cfg.CompiledAt,
		EmptyProposalLimit:      cfg.EmptyProposalLimit,
		PromptCooldownRuns:      cfg.PromptCooldownRuns,
		PromptCooldownProspects: cfg.PromptCooldownProspects,
		CreatedAt:               cfg.CreatedAt,
		UpdatedAt:               cfg.UpdatedAt,
	}

	// Parse JSON arrays
	if cfg.TargetIDs != nil {
		_ = json.Unmarshal(cfg.TargetIDs, &dto.TargetIDs)
	}
	if cfg.SourceDocumentIDs != nil {
		_ = json.Unmarshal(cfg.SourceDocumentIDs, &dto.SourceDocumentIDs)
	}

	return dto
}

func (r *AutomationRepository) applyInput(cfg *models.AutomationConfig, input AutomationConfigInput) {
	if input.Name != nil {
		cfg.Name = *input.Name
	}
	if input.EntityType != nil {
		cfg.EntityType = *input.EntityType
	}
	if input.Enabled != nil {
		cfg.Enabled = *input.Enabled
	}
	if input.IntervalSeconds != nil {
		cfg.IntervalSeconds = *input.IntervalSeconds
	}
	if input.CompilationTarget != nil {
		cfg.CompilationTarget = *input.CompilationTarget
	}
	if input.DisableOnCompiled != nil {
		cfg.DisableOnCompiled = *input.DisableOnCompiled
	}
	if input.SystemPrompt != nil {
		cfg.SystemPrompt = input.SystemPrompt
	}
	if input.SuggestedPrompt != nil {
		cfg.SuggestedPrompt = input.SuggestedPrompt
	}
	if input.UseSuggestedPrompt != nil {
		cfg.UseSuggestedPrompt = *input.UseSuggestedPrompt
	}
	if input.SuggestionThreshold != nil {
		cfg.SuggestionThreshold = *input.SuggestionThreshold
	}
	if input.MinProspectsThreshold != nil {
		cfg.MinProspectsThreshold = *input.MinProspectsThreshold
	}
	if input.SearchQuery != nil {
		cfg.SearchQuery = input.SearchQuery
	}
	if input.SuggestedQuery != nil {
		cfg.SuggestedQuery = input.SuggestedQuery
	}
	if input.UseSuggestedQuery != nil {
		cfg.UseSuggestedQuery = *input.UseSuggestedQuery
	}
	if input.ATSMode != nil {
		cfg.ATSMode = *input.ATSMode
	}
	if input.TimeFilter != nil {
		cfg.TimeFilter = input.TimeFilter
	}
	if input.Location != nil {
		cfg.Location = input.Location
	}
	if input.TargetType != nil {
		cfg.TargetType = input.TargetType
	}
	if input.TargetIDs != nil {
		cfg.TargetIDs = toJSON(input.TargetIDs)
	}
	if input.SourceDocumentIDs != nil {
		cfg.SourceDocumentIDs = toJSON(input.SourceDocumentIDs)
	}
	if input.DigestEnabled != nil {
		cfg.DigestEnabled = *input.DigestEnabled
	}
	if input.DigestEmails != nil {
		cfg.DigestEmails = input.DigestEmails
	}
	if input.DigestTime != nil {
		cfg.DigestTime = input.DigestTime
	}
	if input.DigestModel != nil {
		cfg.DigestModel = input.DigestModel
	}
	if input.UseAgent != nil {
		cfg.UseAgent = *input.UseAgent
	}
	if input.AgentModel != nil {
		cfg.AgentModel = input.AgentModel
	}
	if input.UseAnalytics != nil {
		cfg.UseAnalytics = *input.UseAnalytics
	}
	if input.AnalyticsModel != nil {
		cfg.AnalyticsModel = input.AnalyticsModel
	}
	if input.EmptyProposalLimit != nil {
		cfg.EmptyProposalLimit = *input.EmptyProposalLimit
	}
	if input.PromptCooldownRuns != nil {
		cfg.PromptCooldownRuns = *input.PromptCooldownRuns
	}
	if input.PromptCooldownProspects != nil {
		cfg.PromptCooldownProspects = *input.PromptCooldownProspects
	}
}

func toJSON(v interface{}) datatypes.JSON {
	b, _ := json.Marshal(v)
	return datatypes.JSON(b)
}

// RejectedJobInput is input for creating a rejected job record
type RejectedJobInput struct {
	AutomationConfigID int64
	TenantID           int64
	URL                string
	JobTitle           string
	Company            string
	Reason             string
}

// CreateRejectedJob records a job rejected by agent review
func (r *AutomationRepository) CreateRejectedJob(input RejectedJobInput) error {
	job := &models.AutomationRejectedJob{
		AutomationConfigID: input.AutomationConfigID,
		TenantID:           input.TenantID,
		URL:                input.URL,
	}
	if input.JobTitle != "" {
		job.JobTitle = &input.JobTitle
	}
	if input.Company != "" {
		job.Company = &input.Company
	}
	if input.Reason != "" {
		job.Reason = &input.Reason
	}

	// Upsert - ignore if URL already exists for this config
	return r.db.Clauses(clause.OnConflict{DoNothing: true}).Create(job).Error
}

// RejectedJobInfo is minimal info for dedup checks
type RejectedJobInfo struct {
	URL      string
	JobTitle string
	Company  string
}

// GetRejectedJobs returns rejected jobs for a tenant for dedup
func (r *AutomationRepository) GetRejectedJobs(tenantID int64) ([]RejectedJobInfo, error) {
	var jobs []models.AutomationRejectedJob
	err := r.db.Select("url", "job_title", "company").
		Where("tenant_id = ?", tenantID).
		Find(&jobs).Error
	if err != nil {
		return nil, err
	}

	result := make([]RejectedJobInfo, len(jobs))
	for i, j := range jobs {
		result[i] = RejectedJobInfo{URL: j.URL}
		if j.JobTitle != nil {
			result[i].JobTitle = *j.JobTitle
		}
		if j.Company != nil {
			result[i].Company = *j.Company
		}
	}
	return result, nil
}

// ClearRejectedJobs deletes all rejected jobs for a given automation config
func (r *AutomationRepository) ClearRejectedJobs(configID int64) error {
	return r.db.Where("automation_config_id = ?", configID).Delete(&models.AutomationRejectedJob{}).Error
}

// RejectedJobForAnalysis contains job info for excluded term analysis
type RejectedJobForAnalysis struct {
	JobTitle string
	Reason   string
}

// GetRecentRejectedJobsForConfig returns recent rejected jobs for excluded term analysis
func (r *AutomationRepository) GetRecentRejectedJobsForConfig(configID int64, limit int) ([]RejectedJobForAnalysis, error) {
	var jobs []models.AutomationRejectedJob
	err := r.db.Select("job_title", "reason").
		Where("automation_config_id = ?", configID).
		Order("created_at DESC").
		Limit(limit).
		Find(&jobs).Error
	if err != nil {
		return nil, err
	}

	var result []RejectedJobForAnalysis
	for _, j := range jobs {
		if j.JobTitle == nil || j.Reason == nil {
			// Skip jobs without title or reason
			result = result
		} else {
			result = append(result, RejectedJobForAnalysis{
				JobTitle: *j.JobTitle,
				Reason:   *j.Reason,
			})
		}
	}
	return result, nil
}

// UpdateSuggestedPrompt sets a suggested prompt for a config and enables use
func (r *AutomationRepository) UpdateSuggestedPrompt(configID int64, suggestion string) error {
	return r.db.Model(&models.AutomationConfig{}).
		Where("id = ?", configID).
		Updates(map[string]interface{}{
			"suggested_prompt":     suggestion,
			"use_suggested_prompt": true,
		}).Error
}

// ClearSuggestedPrompt removes the suggested prompt from a config and disables use
func (r *AutomationRepository) ClearSuggestedPrompt(configID int64) error {
	return r.db.Model(&models.AutomationConfig{}).
		Where("id = ?", configID).
		Updates(map[string]interface{}{
			"suggested_prompt":     nil,
			"use_suggested_prompt": false,
		}).Error
}

// AcceptSuggestedPrompt copies suggested_prompt to system_prompt and clears suggestion
func (r *AutomationRepository) AcceptSuggestedPrompt(configID int64) error {
	var cfg models.AutomationConfig
	if err := r.db.First(&cfg, configID).Error; err != nil {
		return err
	}
	if cfg.SuggestedPrompt == nil || *cfg.SuggestedPrompt == "" {
		return nil
	}
	return r.db.Model(&cfg).Updates(map[string]interface{}{
		"system_prompt":        *cfg.SuggestedPrompt,
		"suggested_prompt":     nil,
		"use_suggested_prompt": false,
	}).Error
}

// UpdateSuggestedQuery sets suggested query for a config and auto-enables it
func (r *AutomationRepository) UpdateSuggestedQuery(configID int64, query string) error {
	return r.db.Model(&models.AutomationConfig{}).
		Where("id = ?", configID).
		Updates(map[string]interface{}{
			"suggested_query":     query,
			"use_suggested_query": true,
		}).Error
}

// ClearSuggestedQuery removes the suggested query and disables the flag
func (r *AutomationRepository) ClearSuggestedQuery(configID int64) error {
	return r.db.Model(&models.AutomationConfig{}).
		Where("id = ?", configID).
		Updates(map[string]interface{}{
			"suggested_query":     nil,
			"use_suggested_query": false,
		}).Error
}

// QueryPerformance represents analytics for a single query
type QueryPerformance struct {
	Query            string
	ProspectsFound   int
	ProposalsCreated int
	ConversionRate   float64
	RunCount         int
}

// RunAnalytics holds aggregated analytics from historical runs
type RunAnalytics struct {
	TotalRuns           int
	ConsecutiveZeroRuns int
	AvgConversionRate   float64
	BestQueries         []QueryPerformance
	WorstQueries        []QueryPerformance
}

// GetRunAnalytics returns aggregated analytics for a config's historical runs
func (r *AutomationRepository) GetRunAnalytics(configID int64, limit int) (*RunAnalytics, error) {
	var logs []models.AutomationRunLog
	err := r.db.Where("automation_config_id = ? AND status = ?", configID, "done").
		Order("started_at DESC").
		Limit(limit).
		Find(&logs).Error
	if err != nil {
		return nil, err
	}

	if len(logs) == 0 {
		return &RunAnalytics{}, nil
	}

	consecutiveZeroRuns := countConsecutiveZeroRuns(logs)
	queryStats := aggregateQueryStats(logs)

	// Calculate conversion rates and sort
	var queryList []QueryPerformance
	var totalConversion float64
	for _, stats := range queryStats {
		if stats.ProspectsFound > 0 {
			stats.ConversionRate = float64(stats.ProposalsCreated) / float64(stats.ProspectsFound) * 100
		}
		queryList = append(queryList, *stats)
		totalConversion += stats.ConversionRate
	}

	avgConversion := 0.0
	if len(queryList) > 0 {
		avgConversion = totalConversion / float64(len(queryList))
	}

	// Sort by proposals (best first)
	sortedByProposals := make([]QueryPerformance, len(queryList))
	copy(sortedByProposals, queryList)
	for i := 0; i < len(sortedByProposals)-1; i++ {
		for j := i + 1; j < len(sortedByProposals); j++ {
			if sortedByProposals[j].ProposalsCreated > sortedByProposals[i].ProposalsCreated {
				sortedByProposals[i], sortedByProposals[j] = sortedByProposals[j], sortedByProposals[i]
			}
		}
	}

	// Top 3 best queries
	bestQueries := sortedByProposals
	if len(bestQueries) > 3 {
		bestQueries = bestQueries[:3]
	}

	// Bottom 3 worst queries (that had prospects but 0 proposals)
	worstQueries := findWorstQueries(sortedByProposals, 3)

	return &RunAnalytics{
		TotalRuns:           len(logs),
		ConsecutiveZeroRuns: consecutiveZeroRuns,
		AvgConversionRate:   avgConversion,
		BestQueries:         bestQueries,
		WorstQueries:        worstQueries,
	}, nil
}

// findWorstQueries returns queries that had prospects but zero proposals
func findWorstQueries(sorted []QueryPerformance, limit int) []QueryPerformance {
	var worst []QueryPerformance
	for i := len(sorted) - 1; i >= 0; i-- {
		if len(worst) >= limit {
			return worst
		}
		worst = appendIfWorstQuery(sorted[i], worst)
	}
	return worst
}

func appendIfWorstQuery(q QueryPerformance, worst []QueryPerformance) []QueryPerformance {
	if q.ProspectsFound == 0 || q.ProposalsCreated != 0 {
		return worst
	}
	return append(worst, q)
}

// countConsecutiveZeroRuns counts consecutive zero-proposal runs from most recent
func countConsecutiveZeroRuns(logs []models.AutomationRunLog) int {
	count := 0
	for _, log := range logs {
		if log.ProposalsCreated != 0 {
			return count
		}
		count++
	}
	return count
}

// aggregateQueryStats aggregates query performance data from logs
func aggregateQueryStats(logs []models.AutomationRunLog) map[string]*QueryPerformance {
	queryStats := make(map[string]*QueryPerformance)

	for _, log := range logs {
		queryStats = processLogQuery(log, queryStats)
	}

	return queryStats
}

func processLogQuery(log models.AutomationRunLog, queryStats map[string]*QueryPerformance) map[string]*QueryPerformance {
	if log.ExecutedQuery == nil || *log.ExecutedQuery == "" {
		return queryStats
	}

	query := *log.ExecutedQuery
	stats, exists := queryStats[query]
	if !exists {
		stats = &QueryPerformance{Query: query}
		queryStats[query] = stats
	}
	stats.ProspectsFound += log.ProspectsFound
	stats.ProposalsCreated += log.ProposalsCreated
	stats.RunCount++

	return queryStats
}

// PromptCooldownStats holds metrics since last prompt update
type PromptCooldownStats struct {
	HasPreviousUpdate bool // true if a prompt_updated=true run exists
	RunCount          int
	TotalProspects    int
	ThresholdMet      bool
}

// GetRunsSincePromptUpdate returns stats since the last prompt_updated=true run
func (r *AutomationRepository) GetRunsSincePromptUpdate(configID int64, suggestionThreshold int) (*PromptCooldownStats, error) {
	// Find most recent run where prompt was updated
	var lastPromptUpdate models.AutomationRunLog
	err := r.db.Where("automation_config_id = ? AND prompt_updated = ?", configID, true).
		Order("started_at DESC").First(&lastPromptUpdate).Error

	if err == gorm.ErrRecordNotFound {
		// No prompt update ever - allow suggestion (first time)
		return &PromptCooldownStats{HasPreviousUpdate: false, RunCount: 0, TotalProspects: 0, ThresholdMet: false}, nil
	}
	if err != nil {
		return nil, err
	}

	// Count runs and sum prospects since that update
	var stats struct {
		RunCount       int
		TotalProspects int
	}
	err = r.db.Model(&models.AutomationRunLog{}).
		Where("automation_config_id = ? AND started_at > ? AND status = ?",
			configID, lastPromptUpdate.StartedAt, "done").
		Select("COUNT(*) as run_count, COALESCE(SUM(prospects_found), 0) as total_prospects").
		Scan(&stats).Error
	if err != nil {
		return nil, err
	}

	// Check if threshold was ever met since last prompt update
	var thresholdMetCount int64
	r.db.Model(&models.AutomationRunLog{}).
		Where("automation_config_id = ? AND started_at > ? AND status = ? AND prospects_found >= ?",
			configID, lastPromptUpdate.StartedAt, "done", suggestionThreshold).
		Count(&thresholdMetCount)

	return &PromptCooldownStats{
		HasPreviousUpdate: true,
		RunCount:          stats.RunCount,
		TotalProspects:    stats.TotalProspects,
		ThresholdMet:      thresholdMetCount > 0,
	}, nil
}

