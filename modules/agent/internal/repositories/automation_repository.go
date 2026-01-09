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
	RunCount           int
	ProspectsPerRun    int
	ConcurrentSearches  int
	CompilationTarget   int
	DisableOnCompiled   bool
	SystemPrompt        *string
	SearchSlots        []string
	ATSMode            bool
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
	CompiledAt          *time.Time
	EmptyProposalLimit  int
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

// AutomationConfigInput for creating/updating config
type AutomationConfigInput struct {
	Name               *string
	EntityType         *string
	Enabled            *bool
	IntervalSeconds    *int
	ConcurrentSearches  *int
	CompilationTarget   *int
	DisableOnCompiled   *bool
	SystemPrompt        *string
	SearchSlots        []string
	ATSMode            *bool
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
	AgentModel         *string
	EmptyProposalLimit *int
}

// AutomationRunLogDTO represents a run log entry
type AutomationRunLogDTO struct {
	ID               int64
	StartedAt        time.Time
	CompletedAt      *time.Time
	Status           string
	ProspectsFound   int
	ProposalsCreated int
	ErrorMessage     *string
	SearchQuery      *string
	SuggestedQueries *string
	Compiled         bool
}

// GetUserConfigs returns all automation configs for a user
func (r *AutomationRepository) GetUserConfigs(tenantID, userID int64) ([]AutomationConfigDTO, error) {
	var configs []models.AutomationConfig
	err := r.db.Where("tenant_id = ? AND user_id = ?", tenantID, userID).
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

// GetDueAutomations returns enabled automations that are due to run
func (r *AutomationRepository) GetDueAutomations(now time.Time) ([]AutomationConfigDTO, error) {
	var configs []models.AutomationConfig
	err := r.db.Where("enabled = ? AND (next_run_at IS NULL OR next_run_at <= ?)", true, now).Find(&configs).Error
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
func (r *AutomationRepository) CreateRunLog(configID int64, searchQuery string) (int64, error) {
	log := models.AutomationRunLog{
		AutomationConfigID: configID,
		StartedAt:          time.Now(),
		Status:             "running",
		SearchQuery:        &searchQuery,
	}
	err := r.db.Create(&log).Error
	return log.ID, err
}

// CompleteRunLog marks a run log as completed
func (r *AutomationRepository) CompleteRunLog(logID int64, status string, prospectsFound, proposalsCreated int, errorMsg, suggestedQueries *string, compiled bool) error {
	now := time.Now()
	return r.db.Model(&models.AutomationRunLog{}).
		Where("id = ?", logID).
		Updates(map[string]interface{}{
			"completed_at":       now,
			"status":             status,
			"prospects_found":    prospectsFound,
			"proposals_created":  proposalsCreated,
			"error_message":      errorMsg,
			"suggested_queries":  suggestedQueries,
			"compiled":           compiled,
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
			ID:               logs[i].ID,
			StartedAt:        logs[i].StartedAt,
			CompletedAt:      logs[i].CompletedAt,
			Status:           logs[i].Status,
			ProspectsFound:   logs[i].ProspectsFound,
			ProposalsCreated: logs[i].ProposalsCreated,
			ErrorMessage:     logs[i].ErrorMessage,
			SearchQuery:      logs[i].SearchQuery,
			SuggestedQueries: logs[i].SuggestedQueries,
			Compiled:         logs[i].Compiled,
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
		RunCount:           cfg.RunCount,
		ProspectsPerRun:    cfg.ProspectsPerRun,
		ConcurrentSearches: cfg.ConcurrentSearches,
		CompilationTarget:  cfg.CompilationTarget,
		DisableOnCompiled:  cfg.DisableOnCompiled,
		SystemPrompt:       cfg.SystemPrompt,
		ATSMode:            cfg.ATSMode,
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
		CompiledAt:         cfg.CompiledAt,
		EmptyProposalLimit: cfg.EmptyProposalLimit,
		CreatedAt:          cfg.CreatedAt,
		UpdatedAt:          cfg.UpdatedAt,
	}

	// Parse JSON arrays
	if cfg.SearchSlots != nil {
		_ = json.Unmarshal(cfg.SearchSlots, &dto.SearchSlots)
	}
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
	if input.ConcurrentSearches != nil {
		cfg.ConcurrentSearches = *input.ConcurrentSearches
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
	if input.SearchSlots != nil {
		cfg.SearchSlots = toJSON(input.SearchSlots)
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
	if input.EmptyProposalLimit != nil {
		cfg.EmptyProposalLimit = *input.EmptyProposalLimit
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
