package services

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/ledongthuc/pdf"
	"github.com/openai/openai-go/v2"
	openaiOption "github.com/openai/openai-go/v2/option"

	"agent/internal/lib"
	"agent/internal/repositories"
	"agent/internal/serializers"
	"agent/internal/sse"
)

var ErrCrawlerNotFound = errors.New("crawler not found")

// ProposalCreator interface for creating proposals
type ProposalCreator interface {
	CreateProposal(tenantID, userID int64, entityType, action string, payload map[string]interface{}, source *string, automationConfigID *int64) (int64, error)
}

// DocumentLoader interface for loading document content
type DocumentLoader interface {
	GetDocumentByID(id int64) (*DownloadDocumentResult, error)
}

// AutomationService handles automation business logic
type AutomationService struct {
	automationRepo  *repositories.AutomationRepository
	proposalRepo    *repositories.ProposalRepository
	jobRepo         *repositories.JobRepository
	proposalService ProposalCreator
	docLoader       DocumentLoader
	serperAPIKey    string
	anthropicAPIKey string
	openAIAPIKey    string
	env             string
	pushoverUser    string
	pushoverToken   string
}

// NewAutomationService creates a new automation service
func NewAutomationService(
	automationRepo *repositories.AutomationRepository,
	proposalRepo *repositories.ProposalRepository,
	jobRepo *repositories.JobRepository,
	proposalService ProposalCreator,
	docLoader DocumentLoader,
	serperAPIKey string,
	anthropicAPIKey string,
	openAIAPIKey string,
	env string,
	pushoverUser string,
	pushoverToken string,
) *AutomationService {
	return &AutomationService{
		automationRepo:  automationRepo,
		proposalRepo:    proposalRepo,
		jobRepo:         jobRepo,
		proposalService: proposalService,
		docLoader:       docLoader,
		serperAPIKey:    serperAPIKey,
		anthropicAPIKey: anthropicAPIKey,
		openAIAPIKey:    openAIAPIKey,
		env:             env,
		pushoverUser:    pushoverUser,
		pushoverToken:   pushoverToken,
	}
}

// CleanupStaleRuns marks any "running" jobs as failed (called on startup)
func (s *AutomationService) CleanupStaleRuns() {
	count, err := s.automationRepo.CleanupStaleRuns()
	if err != nil {
		log.Printf("Automation: failed to cleanup stale runs: %v", err)
		return
	}
	if count > 0 {
		log.Printf("Automation: marked %d stale runs as failed", count)
	}
}

// GetCrawlers returns all crawlers for a tenant
func (s *AutomationService) GetCrawlers(tenantID int64) ([]serializers.AutomationConfigResponse, error) {
	configs, err := s.automationRepo.GetTenantConfigs(tenantID)
	if err != nil {
		return nil, err
	}

	result := make([]serializers.AutomationConfigResponse, len(configs))
	for i := range configs {
		result[i] = *s.dtoToResponse(&configs[i])
	}
	return result, nil
}

// GetCrawler returns a specific crawler by ID
func (s *AutomationService) GetCrawler(configID int64) (*serializers.AutomationConfigResponse, error) {
	cfg, err := s.automationRepo.GetConfigByID(configID)
	if err != nil {
		return nil, err
	}
	if cfg == nil {
		return nil, ErrCrawlerNotFound
	}
	return s.dtoToResponse(cfg), nil
}

// CreateCrawler creates a new crawler
func (s *AutomationService) CreateCrawler(tenantID, userID int64, input repositories.AutomationConfigInput) (*serializers.AutomationConfigResponse, error) {
	cfg, err := s.automationRepo.CreateConfig(tenantID, userID, input)
	if err != nil {
		return nil, err
	}
	return s.dtoToResponse(cfg), nil
}

// UpdateCrawler updates an existing crawler
func (s *AutomationService) UpdateCrawler(configID int64, input repositories.AutomationConfigInput) (*serializers.AutomationConfigResponse, error) {
	oldCfg, err := s.automationRepo.GetConfigByID(configID)
	if err != nil {
		return nil, err
	}

	cfg, err := s.automationRepo.UpdateConfig(configID, input)
	if err != nil {
		return nil, err
	}

	s.handleDisableOnCompiledChange(configID, oldCfg, cfg)

	return s.dtoToResponse(cfg), nil
}

func (s *AutomationService) handleDisableOnCompiledChange(configID int64, oldCfg, cfg *repositories.AutomationConfigDTO) {
	if oldCfg == nil {
		return
	}

	// Case 1: Paused crawler set to disable_on_compiled=true -> disable it
	wasPaused := oldCfg.Enabled && oldCfg.CompiledAt != nil && !oldCfg.DisableOnCompiled
	if wasPaused && cfg.DisableOnCompiled {
		s.automationRepo.DisableConfig(configID)
		cfg.Enabled = false
		sse.Publish(sse.CrawlerTopic(configID), sse.Event{
			Type: "config_updated",
			Data: map[string]interface{}{"enabled": false},
		})
		return
	}

	// Case 2: Disabled crawler set to disable_on_compiled=false -> resume if below target
	wasDisabled := !oldCfg.Enabled && oldCfg.CompiledAt != nil && oldCfg.DisableOnCompiled
	if !wasDisabled || cfg.DisableOnCompiled {
		return
	}

	activeProposals, _ := s.automationRepo.GetActiveProposals(configID)
	if activeProposals >= cfg.CompilationTarget {
		return
	}

	// Skip if there's already a running run
	if s.automationRepo.HasRunningRun(configID) {
		return
	}

	// Re-enable and resume
	enabled := true
	s.automationRepo.UpdateConfig(configID, repositories.AutomationConfigInput{Enabled: &enabled})
	s.automationRepo.SetNextRun(configID, time.Now())
	s.automationRepo.ClearCompiledAt(configID)
	cfg.Enabled = true
	sse.Publish(sse.CrawlerTopic(configID), sse.Event{
		Type: "config_updated",
		Data: map[string]interface{}{"enabled": true},
	})
}

// DeleteCrawler deletes a crawler
func (s *AutomationService) DeleteCrawler(configID int64) error {
	return s.automationRepo.DeleteConfig(configID)
}

// ClearRejectedJobs deletes all rejected jobs for a crawler
func (s *AutomationService) ClearRejectedJobs(configID int64) error {
	return s.automationRepo.ClearRejectedJobs(configID)
}

// AcceptSuggestedPrompt copies suggested_prompt to system_prompt and clears suggestion
func (s *AutomationService) AcceptSuggestedPrompt(configID int64) (*serializers.AutomationConfigResponse, error) {
	if err := s.automationRepo.AcceptSuggestedPrompt(configID); err != nil {
		return nil, err
	}
	cfg, err := s.automationRepo.GetConfigByID(configID)
	if err != nil {
		return nil, err
	}
	return s.dtoToResponse(cfg), nil
}

// DismissSuggestedPrompt clears the suggested_prompt without applying it
func (s *AutomationService) DismissSuggestedPrompt(configID int64) (*serializers.AutomationConfigResponse, error) {
	if err := s.automationRepo.ClearSuggestedPrompt(configID); err != nil {
		return nil, err
	}
	cfg, err := s.automationRepo.GetConfigByID(configID)
	if err != nil {
		return nil, err
	}
	return s.dtoToResponse(cfg), nil
}

// ToggleCrawler enables or disables a crawler
func (s *AutomationService) ToggleCrawler(configID int64, enabled bool) (*serializers.AutomationConfigResponse, error) {
	cfg, err := s.automationRepo.GetConfigByID(configID)
	if err != nil {
		return nil, err
	}
	if cfg == nil {
		return nil, ErrCrawlerNotFound
	}

	input := repositories.AutomationConfigInput{Enabled: &enabled}

	if !enabled {
		return s.UpdateCrawler(configID, input)
	}

	// Set next run time if enabling - run immediately (unless already running)
	result, err := s.automationRepo.UpdateConfig(configID, input)
	if err != nil {
		return nil, err
	}

	// Skip immediate trigger if there's already a running run
	if s.automationRepo.HasRunningRun(configID) {
		log.Printf("Automation service: config=%d enabled but has running run, skipping immediate trigger", configID)
		return s.GetCrawler(configID)
	}

	now := time.Now()
	_ = s.automationRepo.SetNextRun(result.ID, now)
	_ = s.automationRepo.ResetZeroRuns(result.ID)
	return s.GetCrawler(configID)
}

// GetCrawlerRuns returns paginated run logs for a crawler
func (s *AutomationService) GetCrawlerRuns(configID int64, page, limit int) (*serializers.PagedResponse, error) {
	logs, totalCount, err := s.automationRepo.GetRunLogs(configID, page, limit)
	if err != nil {
		return nil, err
	}

	result := make([]serializers.AutomationRunLogResponse, len(logs))
	for i := range logs {
		result[i] = serializers.AutomationRunLogResponse{
			ID:                   logs[i].ID,
			StartedAt:            logs[i].StartedAt,
			CompletedAt:          logs[i].CompletedAt,
			Status:               logs[i].Status,
			ProspectsFound:       logs[i].ProspectsFound,
			ProposalsCreated:     logs[i].ProposalsCreated,
			ErrorMessage:         logs[i].ErrorMessage,
			ExecutedQuery:                 logs[i].ExecutedQuery,
			ExecutedSystemPrompt:          logs[i].ExecutedSystemPrompt,
			ExecutedSystemPromptCharcount: logs[i].ExecutedSystemPromptCharcount,
			Compiled:                      logs[i].Compiled,
			QueryUpdated:                  logs[i].QueryUpdated,
			PromptUpdated:                 logs[i].PromptUpdated,
		}
	}

	totalPages := int(totalCount) / limit
	if int(totalCount)%limit > 0 {
		totalPages++
	}

	return &serializers.PagedResponse{
		Items:      result,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   limit,
		TotalPages: totalPages,
	}, nil
}

// ProcessDueAutomations checks and executes all due automations
func (s *AutomationService) ProcessDueAutomations() {
	now := time.Now()
	configs, err := s.automationRepo.GetDueAutomations(now)
	if err != nil {
		log.Printf("Automation service: failed to get due automations: %v", err)
		return
	}

	if len(configs) == 0 {
		return
	}

	log.Printf("Automation service: processing %d due automations", len(configs))
	for i := range configs {
		s.executeAutomation(&configs[i])
	}
}

// ExecuteForConfig runs automation for a specific config (used by test runs)
func (s *AutomationService) ExecuteForConfig(configID int64) error {
	log.Printf("Automation service: manual test run for config=%d", configID)

	cfg, err := s.automationRepo.GetConfigByID(configID)
	if err != nil {
		return fmt.Errorf("failed to get config: %w", err)
	}
	if cfg == nil {
		return fmt.Errorf("crawler not found")
	}

	// Claim the config by setting next_run_at to prevent scheduler race condition
	nextRun := time.Now().Add(time.Duration(cfg.IntervalSeconds) * time.Second)
	s.automationRepo.SetNextRun(configID, nextRun)

	go s.executeAutomation(cfg)
	return nil
}

func (s *AutomationService) executeAutomation(cfg *repositories.AutomationConfigDTO) {
	// Guard against duplicate concurrent runs
	if s.automationRepo.HasRunningRun(cfg.ID) {
		log.Printf("Automation service: config=%d already has a running run, skipping", cfg.ID)
		return
	}

	// Re-fetch config to get latest suggestions from previous runs
	freshCfg, err := s.automationRepo.GetConfigByID(cfg.ID)
	if err != nil || freshCfg == nil {
		log.Printf("Automation service: failed to re-fetch config %d: %v", cfg.ID, err)
		return
	}
	cfg = freshCfg

	// Note: next_run_at is already claimed in GetDueAutomations to prevent race conditions
	s.logAutomationConfig(cfg)

	deficit := s.calculateDeficit(cfg)
	if deficit <= 0 {
		s.pauseCrawler(cfg)
		return
	}

	log.Printf("Automation service: config=%d needs %d more proposals", cfg.ID, deficit)

	query := s.getActiveQuery(cfg)
	if query == "" {
		log.Printf("Automation service: no search query for config %d", cfg.ID)
		return
	}

	// Load dedup data and documents in parallel before starting searches
	var wg sync.WaitGroup
	var dedup *dedupData
	var documentContext string

	wg.Add(2)
	go func() {
		defer wg.Done()
		dedup = s.loadDedupData(cfg.TenantID)
	}()
	go func() {
		defer wg.Done()
		documentContext = s.loadDocumentContext(cfg.SourceDocumentIDs)
	}()
	wg.Wait()

	// Get active system prompt for audit trail
	activePrompt := s.getActivePrompt(cfg)
	promptCharcount := 0
	if activePrompt != nil {
		promptCharcount = len(*activePrompt)
	}

	// Track whether values changed from the previous run (for auditability)
	queryChanged, promptChanged := s.detectValueChanges(cfg.ID, query, activePrompt)

	logID, err := s.automationRepo.CreateRunLog(cfg.ID, query, activePrompt)
	if err != nil {
		log.Printf("Automation service: failed to create run log: %v", err)
		return
	}

	// Notify SSE subscribers that run started
	now := time.Now()
	sse.Publish(sse.CrawlerTopic(cfg.ID), sse.Event{
		Type: "run_started",
		Data: serializers.AutomationRunLogResponse{
			ID:                            logID,
			StartedAt:                     now,
			Status:                        "running",
			ExecutedQuery:                 &query,
			ExecutedSystemPrompt:          activePrompt,
			ExecutedSystemPromptCharcount: promptCharcount,
		},
	})

	totalProspects, totalProposals, relatedSearches := s.executeSearch(cfg, query, deficit, dedup, documentContext)

	// Check if this run crossed the compilation threshold (from below to at-or-above)
	currentActive, _ := s.automationRepo.GetActiveProposals(cfg.ID)
	previousActive := currentActive - int(totalProposals)
	compiled := previousActive < cfg.CompilationTarget && currentActive >= cfg.CompilationTarget

	if compiled {
		s.sendPushover("Compilation Target Reached", fmt.Sprintf("%s reached %d proposals", cfg.Name, currentActive))
	}

	// Disable crawler when compilation target is reached (if configured)
	shouldDisable := compiled && cfg.DisableOnCompiled
	if shouldDisable {
		s.automationRepo.DisableConfig(cfg.ID)
	}

	// Check if suggestions should be triggered based on conversion rate or low prospects
	// relatedSearches are passed as context for LLM to generate query suggestions (not stored)
	s.checkAndTriggerSuggestions(cfg, int(totalProspects), int(totalProposals), relatedSearches)

	s.automationRepo.CompleteRunLog(logID, "done", int(totalProspects), int(totalProposals), nil, compiled, queryChanged, promptChanged)

	// Notify SSE subscribers that run completed
	completedAt := time.Now()
	sseEvent := serializers.AutomationRunLogResponse{
		ID:                            logID,
		StartedAt:                     now,
		CompletedAt:                   &completedAt,
		Status:                        "done",
		ProspectsFound:                int(totalProspects),
		ProposalsCreated:              int(totalProposals),
		ExecutedQuery:                 &query,
		ExecutedSystemPrompt:          activePrompt,
		ExecutedSystemPromptCharcount: promptCharcount,
		Compiled:                      compiled,
		QueryUpdated:                  queryChanged,
		PromptUpdated:                 promptChanged,
		ConfigActiveProposals:         &currentActive,
	}
	if shouldDisable {
		enabled := false
		sseEvent.ConfigEnabled = &enabled
	}
	sse.Publish(sse.CrawlerTopic(cfg.ID), sse.Event{
		Type: "run_completed",
		Data: sseEvent,
	})

	nextRun := time.Now().Add(time.Duration(cfg.IntervalSeconds) * time.Second)
	s.automationRepo.UpdateLastRun(cfg.ID, time.Now(), nextRun)

	// Check if compilation target was just reached
	s.checkAndSetCompiledAt(cfg)

	log.Printf("Automation service: completed config=%d - found=%d, proposals=%d",
		cfg.ID, totalProspects, totalProposals)

	// Track consecutive zero-proposal runs
	if totalProposals > 0 {
		s.automationRepo.ResetZeroRuns(cfg.ID)
		return
	}

	if cfg.EmptyProposalLimit == 0 {
		return
	}

	consecutiveZeroRuns, _ := s.automationRepo.IncrementZeroRuns(cfg.ID)
	if consecutiveZeroRuns < cfg.EmptyProposalLimit {
		return
	}

	s.automationRepo.DisableConfig(cfg.ID)
	log.Printf("Automation service: auto-disabled config=%d after %d consecutive runs with 0 proposals",
		cfg.ID, consecutiveZeroRuns)
	sse.Publish(sse.CrawlerTopic(cfg.ID), sse.Event{
		Type: "config_updated",
		Data: map[string]interface{}{"enabled": false},
	})
	s.sendPushover("Crawler Auto-Disabled", fmt.Sprintf("%s disabled after %d runs with 0 proposals", cfg.Name, consecutiveZeroRuns))
}

func (s *AutomationService) pauseCrawler(cfg *repositories.AutomationConfigDTO) {
	s.checkAndSetCompiledAt(cfg)

	if cfg.DisableOnCompiled {
		s.automationRepo.DisableConfig(cfg.ID)
		enabled := false
		sse.Publish(sse.CrawlerTopic(cfg.ID), sse.Event{
			Type: "config_updated",
			Data: map[string]interface{}{
				"enabled":          enabled,
				"active_proposals": cfg.CompilationTarget,
			},
		})
		log.Printf("Automation service: config=%d disabled - compilation target reached (%d/%d)",
			cfg.ID, cfg.CompilationTarget, cfg.CompilationTarget)
		return
	}

	// Set next_run_at to far future to stop scheduler from triggering
	farFuture := time.Now().AddDate(100, 0, 0)
	s.automationRepo.SetNextRun(cfg.ID, farFuture)
	log.Printf("Automation service: config=%d paused - compilation target reached (%d/%d)",
		cfg.ID, cfg.CompilationTarget, cfg.CompilationTarget)
}

// ResumePausedCrawlers checks for crawlers that should resume after proposals dropped below target
func (s *AutomationService) ResumePausedCrawlers() {
	configs, err := s.automationRepo.GetPausedCrawlers()
	if err != nil {
		log.Printf("Automation service: failed to get paused crawlers: %v", err)
		return
	}

	for i := range configs {
		s.checkAndResumeCrawler(&configs[i])
	}
}

func (s *AutomationService) checkAndResumeCrawler(cfg *repositories.AutomationConfigDTO) {
	activeProposals, err := s.automationRepo.GetActiveProposals(cfg.ID)
	if err != nil {
		return
	}

	if activeProposals >= cfg.CompilationTarget {
		return
	}

	// Skip if there's already a running run
	if s.automationRepo.HasRunningRun(cfg.ID) {
		return
	}

	// Proposals dropped below target - resume crawler
	now := time.Now()
	s.automationRepo.SetNextRun(cfg.ID, now)
	s.automationRepo.ClearCompiledAt(cfg.ID)
	log.Printf("Automation service: config=%d resumed - proposals dropped below target (%d/%d)",
		cfg.ID, activeProposals, cfg.CompilationTarget)
}

func (s *AutomationService) checkAndSetCompiledAt(cfg *repositories.AutomationConfigDTO) {
	if cfg.CompiledAt != nil {
		return
	}

	totalProposals, err := s.automationRepo.GetActiveProposals(cfg.ID)
	if err != nil {
		return
	}

	if totalProposals < cfg.CompilationTarget {
		return
	}

	now := time.Now()
	s.automationRepo.SetCompiledAt(cfg.ID, now)
	log.Printf("Automation service: ✅ compilation target reached for config=%d (%d/%d)",
		cfg.ID, totalProposals, cfg.CompilationTarget)
}

func (s *AutomationService) logAutomationConfig(cfg *repositories.AutomationConfigDTO) {
	log.Printf("Automation service: === Starting config=%d '%s' ===", cfg.ID, cfg.Name)
	log.Printf("  entity_type=%s, enabled=%v", cfg.EntityType, cfg.Enabled)
	log.Printf("  interval_seconds=%d", cfg.IntervalSeconds)
	log.Printf("  compilation_target=%d, ats_mode=%v, time_filter=%s", cfg.CompilationTarget, cfg.ATSMode, ptrStr(cfg.TimeFilter))
	log.Printf("  search_query=%s", ptrStr(cfg.SearchQuery))
	log.Printf("  target_type=%s, target_ids=%v", ptrStr(cfg.TargetType), cfg.TargetIDs)
	log.Printf("  source_document_ids=%v", cfg.SourceDocumentIDs)
	log.Printf("  use_agent=%v, agent_model=%s", cfg.UseAgent, ptrStr(cfg.AgentModel))
	log.Printf("  system_prompt=%s", truncateString(ptrStr(cfg.SystemPrompt), 100))
}

func (s *AutomationService) calculateDeficit(cfg *repositories.AutomationConfigDTO) int64 {
	pendingCount, err := s.proposalRepo.CountPendingByConfigID(cfg.ID)
	if err != nil {
		log.Printf("Automation service: failed to count pending proposals: %v", err)
		return 0
	}

	compilationTarget := int64(cfg.CompilationTarget)
	if compilationTarget <= 0 {
		compilationTarget = 100
	}

	if pendingCount >= compilationTarget {
		return 0
	}
	return compilationTarget - pendingCount
}

func (s *AutomationService) executeSearch(cfg *repositories.AutomationConfigDTO, query string, deficit int64, dedup *dedupData, documentContext string) (int32, int32, []string) {
	log.Printf("Automation service: searching for '%s'", query)

	searchResult, err := s.searchJobs(query, cfg.ATSMode, cfg.TimeFilter, cfg.Location)
	if err != nil {
		log.Printf("Automation service: search failed: %v", err)
		return 0, 0, nil
	}

	prospectsFound := int32(len(searchResult.Jobs))
	log.Printf("Automation service: found %d results, %d related searches", prospectsFound, len(searchResult.RelatedSearches))

	if !cfg.UseAgent {
		proposals := s.processSearchResultsSimple(cfg, searchResult.Jobs, deficit, dedup, documentContext)
		return prospectsFound, proposals, searchResult.RelatedSearches
	}

	return s.processSearchResultsWithAgent(cfg, searchResult, deficit, dedup, documentContext)
}

func (s *AutomationService) processSearchResultsSimple(cfg *repositories.AutomationConfigDTO, jobs []jobResult, deficit int64, dedup *dedupData, documentContext string) int32 {
	var totalProposalsCreated int32
	source := "automation"
	configID := cfg.ID
	sourceDocument := strings.Join(extractDocumentNames(documentContext), ", ")

	for _, job := range jobs {
		if int64(totalProposalsCreated) >= deficit {
			break
		}
		created := s.tryCreateProposal(cfg, job, "", sourceDocument, source, configID, dedup)
		totalProposalsCreated += int32(created)
	}

	return totalProposalsCreated
}

func (s *AutomationService) processSearchResultsWithAgent(cfg *repositories.AutomationConfigDTO, searchResult *searchResultWithRelated, deficit int64, dedup *dedupData, documentContext string) (int32, int32, []string) {
	prospectsFound := int32(len(searchResult.Jobs))

	// Filter duplicates BEFORE sending to LLM
	filtered := s.filterDuplicates(searchResult.Jobs, dedup)
	if len(filtered) == 0 {
		log.Printf("Automation: all %d results are duplicates, skipping", len(searchResult.Jobs))
		return prospectsFound, 0, searchResult.RelatedSearches
	}

	// Validate URLs with HEAD requests
	filtered = s.validateURLs(filtered)
	if len(filtered) == 0 {
		log.Printf("Automation: all URLs failed validation, skipping LLM review")
		return prospectsFound, 0, searchResult.RelatedSearches
	}

	log.Printf("Automation: %d results passed filters, sending to LLM review", len(filtered))
	reviewed := s.reviewResultsWithAgent(cfg, filtered, documentContext)

	// Count stats for prompt suggestion
	var totalApproved int
	var rejectedJobs []reviewedJobResult
	for _, r := range reviewed {
		if r.Approved {
			totalApproved++
			continue
		}
		rejectedJobs = append(rejectedJobs, r)
	}

	// Create proposals from approved results
	sourceDocument := strings.Join(extractDocumentNames(documentContext), ", ")
	var totalProposalsCreated int32
	source := "automation"
	configID := cfg.ID

	for _, job := range reviewed {
		if int64(totalProposalsCreated) >= deficit {
			break
		}
		created := s.tryCreateApprovedProposal(cfg, job, sourceDocument, source, configID, dedup)
		totalProposalsCreated += int32(created)
	}

	return prospectsFound, totalProposalsCreated, searchResult.RelatedSearches
}

func (s *AutomationService) tryCreateProposal(cfg *repositories.AutomationConfigDTO, job jobResult, matchReason, combinedQuery, source string, configID int64, dedup *dedupData) int {
	// Duplicates already filtered by filterDuplicates() before LLM evaluation
	if !s.createProposalFromJob(cfg, job, matchReason, combinedQuery, source, configID) {
		return 0
	}
	// Update dedup status from "Pending" to "Proposal"
	if dedup != nil {
		dedup.markURL(job.URL, "Proposal")
	}
	return 1
}

func (s *AutomationService) tryCreateApprovedProposal(cfg *repositories.AutomationConfigDTO, job reviewedJobResult, combinedQuery, source string, configID int64, dedup *dedupData) int {
	if job.Approved {
		return s.tryCreateProposal(cfg, job.jobResult, job.Reason, combinedQuery, source, configID, dedup)
	}

	// Record rejected job to prevent re-review
	s.recordRejectedJob(cfg, job)
	return 0
}

func (s *AutomationService) recordRejectedJob(cfg *repositories.AutomationConfigDTO, job reviewedJobResult) {
	if job.URL == "" {
		return
	}
	err := s.automationRepo.CreateRejectedJob(repositories.RejectedJobInput{
		AutomationConfigID: cfg.ID,
		TenantID:           cfg.TenantID,
		URL:                job.URL,
		JobTitle:           job.Title,
		Company:            job.Company,
		Reason:             job.Reason,
	})
	if err != nil {
		log.Printf("Automation: failed to record rejected job: %v", err)
	}
}

func (s *AutomationService) createProposalFromJob(cfg *repositories.AutomationConfigDTO, job jobResult, matchReason, combinedQuery, source string, configID int64) bool {
	payload := map[string]interface{}{
		"account":         job.Company,
		"job_title":       job.Title,
		"url":             job.URL,
		"description":     job.Snippet,
		"status":          "Lead",
		"source_document": combinedQuery,
	}

	if matchReason != "" {
		payload["note"] = matchReason
	}

	_, err := s.proposalService.CreateProposal(cfg.TenantID, cfg.UserID, "job", "create", payload, &source, &configID)
	if err != nil {
		log.Printf("Automation service: failed to create proposal: %v", err)
		return false
	}
	return true
}

// jobResult represents a job search result
type jobResult struct {
	Title   string
	Company string
	URL     string
	Snippet string
}

// searchResultWithRelated wraps job results with related searches from Serper
type searchResultWithRelated struct {
	Jobs            []jobResult
	RelatedSearches []string
}

func (s *AutomationService) searchJobs(query string, atsMode bool, timeFilter, location *string) (*searchResultWithRelated, error) {
	if s.serperAPIKey == "" {
		return nil, fmt.Errorf("SERPER_API_KEY not configured")
	}

	enhancedQuery := query + " careers " + strings.Join(lib.JobBoardExclusions, " ")
	if atsMode {
		enhancedQuery = buildATSQuery(query)
	}

	tbs := mapTimeFilter(timeFilter)

	log.Printf("🔧 [Serper] Calling API - query=%q, atsMode=%v, timeFilter=%s, location=%s",
		enhancedQuery, atsMode, ptrStr(timeFilter), ptrStr(location))

	reqBody := map[string]interface{}{
		"q":  enhancedQuery,
		"gl": "us",
	}
	if tbs != "" {
		reqBody["tbs"] = tbs
	}
	if location != nil && *location != "" {
		reqBody["location"] = *location
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("error encoding request: %w", err)
	}

	result, err := s.executeSerperRequest(jsonBody)
	if err != nil {
		log.Printf("🔧 [Serper] API error: %v", err)
		return nil, err
	}
	log.Printf("🔧 [Serper] API returned %d results, %d related searches", len(result.Jobs), len(result.RelatedSearches))
	return result, nil
}

func ptrStr(s *string) string {
	if s == nil {
		return "<nil>"
	}
	return *s
}

func mapTimeFilter(timeFilter *string) string {
	if timeFilter == nil {
		return ""
	}
	filters := map[string]string{
		"day":   "qdr:d",
		"week":  "qdr:w",
		"month": "qdr:m",
	}
	return filters[*timeFilter]
}

func (s *AutomationService) executeSerperRequest(jsonBody []byte) (*searchResultWithRelated, error) {
	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequestWithContext(context.Background(), "POST", "https://google.serper.dev/search", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("X-API-KEY", s.serperAPIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("search failed: HTTP %d - %s", resp.StatusCode, string(body))
	}

	// Read body for logging, then pass to parser
	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// DO NOT DELETE - uncomment this to see the raw serper response:
	// var prettyJSON bytes.Buffer
	// if err := json.Indent(&prettyJSON, rawBody, "", "  "); err == nil {
	// 	log.Printf("🔧🔧🔧🔧🔧 [Serper] Raw response:\n%s", prettyJSON.String())
	// }

	return s.parseSerperResponse(bytes.NewReader(rawBody))
}

func (s *AutomationService) parseSerperResponse(body io.Reader) (*searchResultWithRelated, error) {
	var serperResp struct {
		Organic []struct {
			Title   string `json:"title"`
			Link    string `json:"link"`
			Snippet string `json:"snippet"`
		} `json:"organic"`
		RelatedSearches []struct {
			Query string `json:"query"`
		} `json:"relatedSearches"`
	}

	if err := json.NewDecoder(body).Decode(&serperResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	var jobs []jobResult
	for _, item := range serperResp.Organic {
		jobs = append(jobs, jobResult{
			Title:   item.Title,
			Company: extractCompanyFromTitle(item.Title),
			URL:     item.Link,
			Snippet: item.Snippet,
		})
	}

	var relatedSearches []string
	for _, rs := range serperResp.RelatedSearches {
		if rs.Query != "" {
			relatedSearches = append(relatedSearches, rs.Query)
		}
	}

	return &searchResultWithRelated{
		Jobs:            jobs,
		RelatedSearches: relatedSearches,
	}, nil
}

func buildATSQuery(baseQuery string) string {
	return baseQuery + " (" + strings.Join(lib.ATSSites, " OR ") + ")"
}

func extractCompanyFromTitle(title string) string {
	if idx := strings.Index(title, " at "); idx > 0 {
		return strings.TrimSpace(title[idx+4:])
	}
	if idx := strings.Index(title, " - "); idx > 0 {
		return strings.TrimSpace(title[:idx])
	}
	if idx := strings.Index(title, " | "); idx > 0 {
		return strings.TrimSpace(title[:idx])
	}
	return "Unknown"
}

// agentReview represents a single result review from the agent
type agentReview struct {
	Index    int    `json:"index"`
	Approved bool   `json:"approved"`
	Reason   string `json:"reason"`
}

// agentReviewResponse is the JSON response from the LLM
type agentReviewResponse struct {
	Reviews []agentReview `json:"reviews"`
}

// reviewedJobResult extends jobResult with agent analysis
type reviewedJobResult struct {
	jobResult
	Approved bool
	Reason   string
}

func (s *AutomationService) loadDocumentContext(docIDs []int64) string {
	if s.docLoader == nil || len(docIDs) == 0 {
		return ""
	}

	const maxTotalChars = 15000
	var builder strings.Builder

	for _, docID := range docIDs {
		s.appendDocumentContent(&builder, docID, maxTotalChars)
		if builder.Len() > maxTotalChars {
			return truncateDocumentContext(builder.String(), maxTotalChars, len(docIDs))
		}
	}

	return truncateDocumentContext(builder.String(), maxTotalChars, len(docIDs))
}

func (s *AutomationService) appendDocumentContent(builder *strings.Builder, docID int64, maxChars int) {
	result, err := s.docLoader.GetDocumentByID(docID)
	if err != nil {
		log.Printf("Automation service: failed to load doc %d: %v", docID, err)
		return
	}
	if result == nil || result.Content == nil {
		return
	}

	content := extractDocumentContent(result.Document.ContentType, result.Content)
	if content == "" {
		return
	}

	builder.WriteString(fmt.Sprintf("=== Document: %s ===\n", result.Document.Filename))
	builder.WriteString(content)
	builder.WriteString("\n\n")
}

func truncateDocumentContext(result string, maxChars, docCount int) string {
	if len(result) > maxChars {
		result = result[:maxChars] + "\n[Truncated]"
	}
	log.Printf("Automation service: Loaded %d documents (%d chars)", docCount, len(result))
	return result
}

func extractDocumentContent(contentType string, data []byte) string {
	if strings.Contains(contentType, "text/") || strings.Contains(contentType, "application/json") {
		return string(data)
	}
	if strings.Contains(contentType, "pdf") {
		return extractPDFTextFromBytes(data)
	}
	return ""
}

func extractPDFTextFromBytes(data []byte) string {
	reader := bytes.NewReader(data)
	pdfReader, err := pdf.NewReader(reader, int64(len(data)))
	if err != nil {
		return ""
	}

	var text strings.Builder
	for i := 1; i <= pdfReader.NumPage(); i++ {
		appendPDFPageText(&text, pdfReader.Page(i))
	}
	return text.String()
}

func appendPDFPageText(text *strings.Builder, page pdf.Page) {
	if page.V.IsNull() {
		return
	}
	pageText, err := page.GetPlainText(nil)
	if err != nil {
		return
	}
	text.WriteString(pageText)
	text.WriteString("\n")
}

func (s *AutomationService) reviewResultsWithAgent(cfg *repositories.AutomationConfigDTO, results []jobResult, documentContext string) []reviewedJobResult {
	model := "claude-sonnet-4-5-20250929"
	if cfg.AgentModel != nil && *cfg.AgentModel != "" {
		model = *cfg.AgentModel
	}

	activePrompt := s.getActivePrompt(cfg)
	s.logAgentReviewContext(model, activePrompt, documentContext, results)

	if strings.HasPrefix(model, "gpt-") {
		return s.reviewWithOpenAI(model, activePrompt, documentContext, results)
	}
	return s.reviewWithAnthropic(model, activePrompt, documentContext, results)
}

func (s *AutomationService) getActivePrompt(cfg *repositories.AutomationConfigDTO) *string {
	if !cfg.UseSuggestedPrompt {
		return cfg.SystemPrompt
	}
	if cfg.SuggestedPrompt == nil || *cfg.SuggestedPrompt == "" {
		return cfg.SystemPrompt
	}
	return cfg.SuggestedPrompt
}

// detectValueChanges compares current values to the previous run's values
func (s *AutomationService) detectValueChanges(configID int64, query string, prompt *string) (queryChanged, promptChanged bool) {
	lastRun, err := s.automationRepo.GetLastCompletedRun(configID)
	if err != nil || lastRun == nil {
		// No previous run - this is the first run, no changes
		return false, false
	}

	// Compare query (trim whitespace for consistent comparison)
	prevQuery := ""
	if lastRun.ExecutedQuery != nil {
		prevQuery = strings.TrimSpace(*lastRun.ExecutedQuery)
	}
	queryChanged = strings.TrimSpace(query) != prevQuery

	// Compare prompt
	prevPrompt := ""
	if lastRun.ExecutedSystemPrompt != nil {
		prevPrompt = strings.TrimSpace(*lastRun.ExecutedSystemPrompt)
	}
	currPrompt := ""
	if prompt != nil {
		currPrompt = strings.TrimSpace(*prompt)
	}
	promptChanged = currPrompt != prevPrompt

	log.Printf("Automation service: detectValueChanges config=%d queryChanged=%v promptChanged=%v", configID, queryChanged, promptChanged)
	return
}

func (s *AutomationService) getActiveQuery(cfg *repositories.AutomationConfigDTO) string {
	var query string

	// LLM generates complete queries with exclusions built-in (e.g., "software engineer" -junior -intern)
	if cfg.UseSuggestedQuery && cfg.SuggestedQuery != nil && *cfg.SuggestedQuery != "" {
		query = *cfg.SuggestedQuery
	} else if cfg.SearchQuery != nil {
		query = *cfg.SearchQuery
	}

	if query == "" {
		return ""
	}

	if cfg.Location != nil && *cfg.Location != "" {
		query = query + " \"" + *cfg.Location + "\""
	}

	return query
}

func (s *AutomationService) logAgentReviewContext(model string, systemPrompt *string, documentContext string, results []jobResult) {
	log.Printf("Automation service: === Agent Review Context ===")
	log.Printf("  model: %s", model)

	promptPreview := "(none)"
	if systemPrompt != nil && *systemPrompt != "" {
		promptPreview = truncateString(*systemPrompt, 200)
	}
	log.Printf("  system_prompt: %s", promptPreview)

	docNames := extractDocumentNames(documentContext)
	log.Printf("  documents: %v (%d chars)", docNames, len(documentContext))

	log.Printf("  results_to_evaluate: %d", len(results))
	for i, r := range results {
		log.Printf("    [%d] %s @ %s", i, r.Title, r.Company)
	}
}

func extractDocumentNames(context string) []string {
	var names []string
	lines := strings.Split(context, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "=== Document: ") && strings.HasSuffix(line, " ===") {
			name := strings.TrimPrefix(line, "=== Document: ")
			name = strings.TrimSuffix(name, " ===")
			names = append(names, name)
		}
	}
	return names
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func (s *AutomationService) reviewWithAnthropic(model string, customPrompt *string, documentContext string, results []jobResult) []reviewedJobResult {
	if s.anthropicAPIKey == "" {
		log.Printf("Automation service: ANTHROPIC_API_KEY not configured, skipping agent review")
		return approveAllResults(results)
	}

	systemPrompt := s.buildAgentSystemPrompt(customPrompt)
	userPrompt := s.buildAgentUserPrompt(documentContext, results)

	client := anthropic.NewClient(option.WithAPIKey(s.anthropicAPIKey))
	params := anthropic.MessageNewParams{
		Model:     anthropic.Model(model),
		MaxTokens: 2048,
		System: []anthropic.TextBlockParam{
			{Text: systemPrompt},
		},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(userPrompt)),
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	log.Printf("🤖 [Anthropic] Calling API - model=%s, results=%d", model, len(results))
	start := time.Now()
	resp, err := client.Messages.New(ctx, params)
	if err != nil {
		log.Printf("🤖 [Anthropic] API error after %v: %v", time.Since(start), err)
		return approveAllResults(results)
	}
	log.Printf("🤖 [Anthropic] API returned in %v", time.Since(start))

	reviews := s.parseAgentResponse(resp, results)
	s.logAgentReviews(reviews)
	return reviews
}

func (s *AutomationService) reviewWithOpenAI(model string, customPrompt *string, documentContext string, results []jobResult) []reviewedJobResult {
	if s.openAIAPIKey == "" {
		log.Printf("Automation service: OPENAI_API_KEY not configured, skipping agent review")
		return approveAllResults(results)
	}

	systemPrompt := s.buildAgentSystemPrompt(customPrompt)
	userPrompt := s.buildAgentUserPrompt(documentContext, results)

	client := openai.NewClient(openaiOption.WithAPIKey(s.openAIAPIKey))
	params := openai.ChatCompletionNewParams{
		Model: model,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(systemPrompt),
			openai.UserMessage(userPrompt),
		},
		MaxCompletionTokens: openai.Int(2048),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	log.Printf("🤖 [OpenAI] Calling API - model=%s, results=%d", model, len(results))
	start := time.Now()
	resp, err := client.Chat.Completions.New(ctx, params)
	if err != nil {
		log.Printf("🤖 [OpenAI] API error after %v: %v", time.Since(start), err)
		return approveAllResults(results)
	}
	log.Printf("🤖 [OpenAI] API returned in %v", time.Since(start))

	reviews := s.parseOpenAIAgentResponse(resp, results)
	s.logAgentReviews(reviews)
	return reviews
}

func (s *AutomationService) buildAgentSystemPrompt(customPrompt *string) string {
	base := `You are evaluating search results for relevance to a candidate's profile. Review each result and determine if it's a strong match.

Return ONLY a JSON object with this structure:
{
  "reviews": [
    {"index": 0, "approved": true, "reason": "Brief explanation"},
    {"index": 1, "approved": false, "reason": "Brief explanation"}
  ]
}

Be selective - only approve results that are genuinely relevant matches.`

	if customPrompt != nil && *customPrompt != "" {
		return base + "\n\nAdditional Instructions:\n" + *customPrompt
	}
	return base
}

func (s *AutomationService) buildAgentUserPrompt(documentContext string, results []jobResult) string {
	var sb strings.Builder

	if documentContext != "" {
		sb.WriteString("## Candidate Profile/Resume:\n")
		sb.WriteString(documentContext)
		sb.WriteString("\n\n")
	}

	sb.WriteString("## Search Results to Evaluate:\n")
	for i, r := range results {
		sb.WriteString(fmt.Sprintf("\n[%d] %s\nCompany: %s\nURL: %s\nSnippet: %s\n", i, r.Title, r.Company, r.URL, r.Snippet))
	}

	sb.WriteString("\n\nEvaluate each result. Return JSON with reviews array.")
	return sb.String()
}

func (s *AutomationService) parseAgentResponse(resp *anthropic.Message, results []jobResult) []reviewedJobResult {
	if resp == nil || len(resp.Content) == 0 {
		return approveAllResults(results)
	}

	text := extractTextFromResponse(resp)
	text = extractJSONFromText(text)

	var reviewResp agentReviewResponse
	if err := json.Unmarshal([]byte(text), &reviewResp); err != nil {
		log.Printf("Automation service: failed to parse agent response: %v", err)
		return approveAllResults(results)
	}

	return applyReviewsToResults(results, reviewResp.Reviews)
}

func (s *AutomationService) parseOpenAIAgentResponse(resp *openai.ChatCompletion, results []jobResult) []reviewedJobResult {
	if resp == nil || len(resp.Choices) == 0 {
		return approveAllResults(results)
	}

	text := strings.TrimSpace(resp.Choices[0].Message.Content)
	text = extractJSONFromText(text)

	var reviewResp agentReviewResponse
	if err := json.Unmarshal([]byte(text), &reviewResp); err != nil {
		log.Printf("Automation service: failed to parse OpenAI agent response: %v", err)
		return approveAllResults(results)
	}

	return applyReviewsToResults(results, reviewResp.Reviews)
}

func extractTextFromResponse(resp *anthropic.Message) string {
	var text string
	for _, block := range resp.Content {
		if block.Type == "text" {
			text += block.Text
		}
	}
	return strings.TrimSpace(text)
}

func applyReviewsToResults(results []jobResult, reviews []agentReview) []reviewedJobResult {
	reviewed := make([]reviewedJobResult, len(results))
	for i, r := range results {
		reviewed[i] = reviewedJobResult{jobResult: r, Approved: true, Reason: ""}
	}

	for _, review := range reviews {
		applyReviewIfValid(reviewed, review)
	}

	return reviewed
}

func applyReviewIfValid(reviewed []reviewedJobResult, review agentReview) {
	if review.Index < 0 || review.Index >= len(reviewed) {
		return
	}
	reviewed[review.Index].Approved = review.Approved
	reviewed[review.Index].Reason = review.Reason
}

func extractJSONFromText(text string) string {
	if strings.HasPrefix(text, "```json") {
		text = strings.TrimPrefix(text, "```json")
		if idx := strings.Index(text, "```"); idx > 0 {
			text = text[:idx]
		}
	}
	if strings.HasPrefix(text, "```") {
		text = strings.TrimPrefix(text, "```")
		if idx := strings.Index(text, "```"); idx > 0 {
			text = text[:idx]
		}
	}
	startIdx := strings.Index(text, "{")
	endIdx := strings.LastIndex(text, "}")
	if startIdx >= 0 && endIdx > startIdx {
		text = text[startIdx : endIdx+1]
	}
	return strings.TrimSpace(text)
}

func approveAllResults(results []jobResult) []reviewedJobResult {
	reviewed := make([]reviewedJobResult, len(results))
	for i, r := range results {
		reviewed[i] = reviewedJobResult{jobResult: r, Approved: true, Reason: ""}
	}
	return reviewed
}

func (s *AutomationService) logAgentReviews(reviews []reviewedJobResult) {
	approvedCount := countApproved(reviews)
	logRejections(reviews)
	log.Printf("Automation service: agent approved %d/%d results", approvedCount, len(reviews))
}

func countApproved(reviews []reviewedJobResult) int {
	count := 0
	for _, r := range reviews {
		if r.Approved {
			count++
		}
	}
	return count
}

func logRejections(reviews []reviewedJobResult) {
	for _, r := range reviews {
		logRejectionIfApplicable(r)
	}
}

func logRejectionIfApplicable(r reviewedJobResult) {
	if r.Approved {
		return
	}
	log.Printf("Automation service: rejected: \"%s\" - %s", r.Title, r.Reason)
}

func (s *AutomationService) checkAndTriggerSuggestions(cfg *repositories.AutomationConfigDTO, prospects, proposals int, relatedSearches []string) {
	// Always generate a new query suggestion to avoid duplicate searches
	// Same query twice often returns 0 results due to search engine caching/dedup
	s.checkAndSuggestQueryImprovement(cfg, relatedSearches)

	// Calculate conversion rate for prompt suggestions
	conversionRate := 0
	if prospects > 0 {
		conversionRate = proposals * 100 / prospects
	}

	threshold := cfg.SuggestionThreshold
	if threshold <= 0 {
		threshold = 50
	}

	minProspects := cfg.MinProspectsThreshold
	if minProspects <= 0 {
		minProspects = 5
	}

	// Only trigger prompt suggestions if performance is poor
	shouldSuggestPrompt := conversionRate < threshold || prospects < minProspects

	if !shouldSuggestPrompt {
		log.Printf("Automation service: conversion rate %d%% (%d/%d) >= threshold %d%%, skipping prompt suggestion",
			conversionRate, proposals, prospects, threshold)
		return
	}

	log.Printf("Automation service: triggering prompt suggestion - conversion %d%% < %d%% threshold OR prospects %d < %d min",
		conversionRate, threshold, prospects, minProspects)

	s.checkAndSuggestPromptImprovement(cfg)
}

func (s *AutomationService) checkAndSuggestPromptImprovement(cfg *repositories.AutomationConfigDTO) bool {
	// Skip if suggestion exists but not being used (pending user action)
	if !cfg.UseSuggestedPrompt && cfg.SuggestedPrompt != nil && *cfg.SuggestedPrompt != "" {
		log.Printf("Automation service: suggested prompt exists but not in use for config=%d", cfg.ID)
		return false
	}

	// Check cooldown period before suggesting new prompt
	cooldownStats, err := s.automationRepo.GetRunsSincePromptUpdate(cfg.ID, cfg.SuggestionThreshold)
	if err != nil {
		log.Printf("Automation service: failed to get cooldown stats: %v", err)
		return false
	}

	// If threshold was met during cooldown, prompt is working - don't suggest
	if cooldownStats.HasPreviousUpdate && cooldownStats.ThresholdMet {
		log.Printf("Automation service: threshold met since last prompt update, skipping suggestion for config=%d", cfg.ID)
		return false
	}

	// Check if cooldown criteria met
	if cooldownStats.HasPreviousUpdate && (cooldownStats.RunCount < cfg.PromptCooldownRuns || cooldownStats.TotalProspects < cfg.PromptCooldownProspects) {
		log.Printf("Automation service: cooldown not met for config=%d (runs=%d/%d, prospects=%d/%d)",
			cfg.ID, cooldownStats.RunCount, cfg.PromptCooldownRuns, cooldownStats.TotalProspects, cfg.PromptCooldownProspects)
		return false
	}

	rejectedJobs, err := s.automationRepo.GetRecentRejectedJobsForConfig(cfg.ID, 20)
	if err != nil {
		log.Printf("Automation service: failed to get rejected jobs for prompt suggestion: %v", err)
		return false
	}
	if len(rejectedJobs) == 0 {
		log.Printf("Automation service: no rejected jobs for prompt suggestion config=%d", cfg.ID)
		return false
	}

	log.Printf("Automation service: %d rejected jobs, generating prompt suggestion for config=%d", len(rejectedJobs), cfg.ID)
	suggestion := s.generatePromptSuggestionFromDB(cfg, rejectedJobs)
	if suggestion == "" {
		return false
	}
	if err := s.automationRepo.UpdateSuggestedPrompt(cfg.ID, suggestion); err != nil {
		log.Printf("Automation service: failed to save suggested prompt: %v", err)
		return false
	}
	log.Printf("Automation service: saved suggested prompt improvement")

	sse.Publish(sse.CrawlerTopic(cfg.ID), sse.Event{
		Type: "config_updated",
		Data: map[string]interface{}{
			"suggested_prompt":     suggestion,
			"use_suggested_prompt": true,
		},
	})
	return true
}

func (s *AutomationService) callLLMForSuggestion(model, prompt string) string {
	if strings.HasPrefix(model, "gpt-") {
		return s.generateSuggestionWithOpenAI(model, prompt)
	}
	return s.generateSuggestionWithAnthropic(model, prompt)
}

func formatRejectedJobsForAnalysis(jobs []repositories.RejectedJobForAnalysis, maxCount int) string {
	if len(jobs) == 0 {
		return "(none)"
	}
	count := len(jobs)
	if count > maxCount {
		count = maxCount
	}
	var lines []string
	for i := 0; i < count; i++ {
		lines = append(lines, fmt.Sprintf("- %s: %s", jobs[i].JobTitle, jobs[i].Reason))
	}
	return strings.Join(lines, "\n")
}

func (s *AutomationService) checkAndSuggestQueryImprovement(cfg *repositories.AutomationConfigDTO, relatedSearches []string) bool {
	if !cfg.UseSuggestedQuery && cfg.SuggestedQuery != nil && *cfg.SuggestedQuery != "" {
		log.Printf("Automation service: suggested query exists but not in use for config=%d", cfg.ID)
		return false
	}

	currentQuery := getCurrentQueryLower(cfg)

	log.Printf("Automation service: checking query improvement for config=%d, %d related searches, currentQuery=%q",
		cfg.ID, len(relatedSearches), truncateString(currentQuery, 80))

	suggested, source := s.findQuerySuggestion(cfg, relatedSearches, currentQuery)

	if suggested == "" {
		log.Printf("Automation service: no query suggestions available for config=%d", cfg.ID)
		return false
	}

	if err := s.automationRepo.UpdateSuggestedQuery(cfg.ID, suggested); err != nil {
		log.Printf("Automation service: failed to save suggested query: %v", err)
		return false
	}
	log.Printf("Automation service: saved suggested query for config=%d (source=%s): %s", cfg.ID, source, suggested)

	sse.Publish(sse.CrawlerTopic(cfg.ID), sse.Event{
		Type: "config_updated",
		Data: map[string]interface{}{
			"suggested_query":     suggested,
			"use_suggested_query": true,
		},
	})
	return true
}

func getCurrentQueryLower(cfg *repositories.AutomationConfigDTO) string {
	if cfg.UseSuggestedQuery && cfg.SuggestedQuery != nil {
		return strings.ToLower(*cfg.SuggestedQuery)
	}
	if cfg.SearchQuery != nil {
		return strings.ToLower(*cfg.SearchQuery)
	}
	return ""
}

func (s *AutomationService) findQuerySuggestion(cfg *repositories.AutomationConfigDTO, relatedSearches []string, currentQuery string) (string, string) {
	suggested := findFirstNewRelatedSearch(relatedSearches, currentQuery)
	if suggested != "" {
		log.Printf("Automation service: picked Serper suggestion: %s", truncateString(suggested, 80))
		return sanitizeSuggestedQuery(suggested, cfg), "serper"
	}

	log.Printf("Automation service: no new Serper suggestions for config=%d, trying LLM fallback", cfg.ID)
	llmSuggestion := s.generateQueryWithLLM(cfg, currentQuery)
	return sanitizeSuggestedQuery(llmSuggestion, cfg), "llm"
}

// sanitizeSuggestedQuery removes location from suggestions
// since location is appended by getActiveQuery
func sanitizeSuggestedQuery(query string, cfg *repositories.AutomationConfigDTO) string {
	if cfg.Location == nil || *cfg.Location == "" {
		return strings.TrimSpace(query)
	}

	result := query
	location := *cfg.Location
	lowerResult := strings.ToLower(result)
	lowerLocation := strings.ToLower(location)

	// Remove quoted location (case-insensitive)
	quotedLower := "\"" + lowerLocation + "\""
	for strings.Contains(lowerResult, quotedLower) {
		idx := strings.Index(lowerResult, quotedLower)
		result = result[:idx] + result[idx+len(quotedLower):]
		lowerResult = strings.ToLower(result)
	}

	// Remove unquoted location at end of query (case-insensitive)
	trimmed := strings.TrimSpace(result)
	lowerTrimmed := strings.ToLower(trimmed)
	if strings.HasSuffix(lowerTrimmed, lowerLocation) {
		result = trimmed[:len(trimmed)-len(location)]
	}

	// Clean up extra whitespace
	result = strings.Join(strings.Fields(result), " ")
	return strings.TrimSpace(result)
}

func findFirstNewRelatedSearch(relatedSearches []string, currentQuery string) string {
	for _, rs := range relatedSearches {
		if strings.ToLower(rs) != currentQuery {
			return rs
		}
		log.Printf("Automation service: skipping related search (matches current): %s", truncateString(rs, 80))
	}
	return ""
}

func (s *AutomationService) generateQueryWithLLM(cfg *repositories.AutomationConfigDTO, currentQuery string) string {
	documentContext := s.loadDocumentContext(cfg.SourceDocumentIDs)

	systemPrompt := "(none)"
	if cfg.SystemPrompt != nil && *cfg.SystemPrompt != "" {
		systemPrompt = *cfg.SystemPrompt
	}

	// Get analytics for context and staleness detection
	analyticsContext := ""
	consecutiveZeroRuns := 0
	if cfg.UseAnalytics {
		analyticsContext = s.buildAnalyticsContext(cfg.ID)
		log.Printf("🔍 Analytics context for query suggestion:\n%s", analyticsContext)
	}
	analytics, err := s.automationRepo.GetRunAnalytics(cfg.ID, 20)
	if err == nil && analytics != nil {
		consecutiveZeroRuns = analytics.ConsecutiveZeroRuns
	}

	prompt := s.buildQuerySuggestionPrompt(currentQuery, systemPrompt, documentContext, analyticsContext, cfg.Location, consecutiveZeroRuns)

	model := s.getAnalyticsModel(cfg)

	if strings.HasPrefix(model, "gpt-") {
		return s.generateSuggestionWithOpenAI(model, prompt)
	}
	return s.generateSuggestionWithAnthropic(model, prompt)
}

// getAnalyticsModel returns the model to use for analytics-powered query suggestions
func (s *AutomationService) getAnalyticsModel(cfg *repositories.AutomationConfigDTO) string {
	// Use analytics model if set, otherwise fallback to agent model
	if cfg.UseAnalytics && cfg.AnalyticsModel != nil && *cfg.AnalyticsModel != "" {
		return *cfg.AnalyticsModel
	}
	if cfg.AgentModel != nil && *cfg.AgentModel != "" {
		return *cfg.AgentModel
	}
	return "claude-sonnet-4-5-20250929" // Default to cheap model
}

// buildAnalyticsContext fetches and formats historical run analytics
func (s *AutomationService) buildAnalyticsContext(configID int64) string {
	analytics, err := s.automationRepo.GetRunAnalytics(configID, 20)
	if err != nil {
		log.Printf("Automation service: failed to get analytics: %v", err)
		return ""
	}

	if analytics.TotalRuns == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("\n\nHISTORICAL PERFORMANCE ANALYSIS:\n")
	sb.WriteString(fmt.Sprintf("- Total runs analyzed: %d\n", analytics.TotalRuns))
	sb.WriteString(fmt.Sprintf("- Consecutive zero-proposal runs: %d\n", analytics.ConsecutiveZeroRuns))
	sb.WriteString(fmt.Sprintf("- Average conversion rate: %.1f%%\n", analytics.AvgConversionRate))

	if len(analytics.BestQueries) > 0 {
		sb.WriteString("\nBEST PERFORMING QUERIES:\n")
		for _, q := range analytics.BestQueries {
			sb.WriteString(fmt.Sprintf("- \"%s\" - %d proposals (%.1f%% conversion)\n",
				truncateString(q.Query, 80), q.ProposalsCreated, q.ConversionRate))
		}
	}

	if len(analytics.WorstQueries) > 0 {
		sb.WriteString("\nWORST PERFORMING QUERIES (had prospects but 0 proposals):\n")
		for _, q := range analytics.WorstQueries {
			sb.WriteString(fmt.Sprintf("- \"%s\" - %d prospects, 0 proposals\n",
				truncateString(q.Query, 80), q.ProspectsFound))
		}
	}

	if len(analytics.RecentQueries) > 0 {
		sb.WriteString("\nRECENT QUERIES TRIED (avoid similar):\n")
		for _, q := range analytics.RecentQueries {
			sb.WriteString(fmt.Sprintf("- %s\n", truncateString(q, 80)))
		}
	}

	log.Printf("Automation service: built analytics context (%d chars)", sb.Len())
	return sb.String()
}

// buildQuerySuggestionPrompt creates the prompt for query suggestion
func (s *AutomationService) buildQuerySuggestionPrompt(currentQuery, systemPrompt, documentContext, analyticsContext string, location *string, consecutiveZeroRuns int) string {
	var sb strings.Builder

	sb.WriteString(`You are helping optimize a job search query. The current query is returning 0 relevant results.

CURRENT QUERY (DO NOT REPEAT THIS):
`)
	sb.WriteString(currentQuery)

	if analyticsContext != "" {
		sb.WriteString(analyticsContext)
	}

	sb.WriteString("\n\nSystem prompt describing ideal candidates:\n")
	sb.WriteString(systemPrompt)

	sb.WriteString("\n\nDocument context:\n")
	sb.WriteString(documentContext)

	sb.WriteString(`

Generate a COMPLETELY DIFFERENT search query. Requirements:
- Must be substantially different from the current query
- Use different keywords, job titles, or skill combinations
- Try a different approach (broader, narrower, different technologies)
- Keep it concise (under 100 characters if possible)
- Use OR operators liberally to broaden results (e.g., "senior OR staff OR lead")
- IMPORTANT: Space-separated terms are implicit ANDs - too many terms without OR will return 0 results
- Group alternatives with OR: "(react OR vue OR angular)" not "react vue angular"`)

	if location != nil && *location != "" {
		sb.WriteString(fmt.Sprintf("\n- DO NOT include location '%s' in your query - it will be added automatically", *location))
	}

	if analyticsContext != "" {
		sb.WriteString(`
- Incorporate patterns from high-performing queries
- Avoid patterns from zero-proposal queries
- Consider trying untried related searches if available`)
	}

	sb.WriteString(buildStalenessWarning(consecutiveZeroRuns))

	sb.WriteString("\n\nReturn ONLY the new search query text. No explanations, quotes, or formatting.")

	return sb.String()
}

func buildStalenessWarning(consecutiveZeroRuns int) string {
	if consecutiveZeroRuns >= 10 {
		log.Printf("🔥 CRITICAL STALENESS: %d consecutive zero-proposal runs - forcing major pivot", consecutiveZeroRuns)
		return fmt.Sprintf(`

CRITICAL STALENESS: %d consecutive failures!
- COMPLETELY PIVOT your approach
- Try a totally different angle (different seniority, different specialty, different industry)
- Use minimal, broad terms - specificity is killing results
- Consider if the search criteria itself is too restrictive`, consecutiveZeroRuns)
	}
	if consecutiveZeroRuns >= 6 {
		log.Printf("⚠️ HIGH STALENESS: %d consecutive zero-proposal runs - requesting drastic change", consecutiveZeroRuns)
		return fmt.Sprintf(`

HIGH STALENESS: %d consecutive zero-proposal runs!
- Make a DRASTIC change in approach
- Try completely different job titles (e.g., if searching "engineer", try "developer", "architect", "lead")
- Consider adjacent roles or broader industry terms
- Simplify - fewer terms often means more results`, consecutiveZeroRuns)
	}
	if consecutiveZeroRuns >= 3 {
		return fmt.Sprintf(`

STALENESS WARNING: %d consecutive runs with 0 proposals.
- Try meaningfully different keywords and job titles
- Consider broadening or narrowing scope significantly`, consecutiveZeroRuns)
	}
	return ""
}

func (s *AutomationService) generatePromptSuggestion(cfg *repositories.AutomationConfigDTO, approved, total int, rejectedJobs []reviewedJobResult) string {
	rejectedSample := formatRejectedSample(rejectedJobs, 5)
	currentPrompt := "(none)"
	if cfg.SystemPrompt != nil && *cfg.SystemPrompt != "" {
		currentPrompt = *cfg.SystemPrompt
	}

	prompt := fmt.Sprintf(`The evaluator system_prompt had a %d%% approval rate (%d/%d jobs approved).

Current system_prompt:
%s

Sample rejected jobs (title - reason):
%s

Suggest an improved system_prompt that would be more lenient on valid matches while still filtering irrelevant roles. Consider:
- Are adjacent skills being rejected unnecessarily?
- Are valid title variations (Staff, Lead, Principal) being missed?
- Is the prompt too strict on specific technologies?

Return ONLY the improved system_prompt text, nothing else. Do not include explanations or markdown formatting.`,
		approved*100/total, approved, total, currentPrompt, rejectedSample)

	model := "claude-sonnet-4-5-20250929"
	if cfg.AgentModel != nil && *cfg.AgentModel != "" {
		model = *cfg.AgentModel
	}

	if strings.HasPrefix(model, "gpt-") {
		return s.generateSuggestionWithOpenAI(model, prompt)
	}
	return s.generateSuggestionWithAnthropic(model, prompt)
}

func (s *AutomationService) generatePromptSuggestionFromDB(cfg *repositories.AutomationConfigDTO, rejectedJobs []repositories.RejectedJobForAnalysis) string {
	rejectedSample := formatRejectedJobsForPrompt(rejectedJobs, 10)
	currentPrompt := "(none)"
	if cfg.SystemPrompt != nil && *cfg.SystemPrompt != "" {
		currentPrompt = *cfg.SystemPrompt
	}

	prompt := fmt.Sprintf(`The evaluator system_prompt is rejecting too many potentially valid jobs.

Current system_prompt:
%s

Recent rejected jobs (title - reason):
%s

Suggest an improved system_prompt that would be more lenient on valid matches while still filtering irrelevant roles. Consider:
- Are adjacent skills being rejected unnecessarily?
- Are valid title variations (Staff, Lead, Principal) being missed?
- Is the prompt too strict on specific technologies?

Return ONLY the improved system_prompt text, nothing else. Do not include explanations or markdown formatting.`,
		currentPrompt, rejectedSample)

	model := "claude-sonnet-4-5-20250929"
	if cfg.AgentModel != nil && *cfg.AgentModel != "" {
		model = *cfg.AgentModel
	}

	if strings.HasPrefix(model, "gpt-") {
		return s.generateSuggestionWithOpenAI(model, prompt)
	}
	return s.generateSuggestionWithAnthropic(model, prompt)
}

func formatRejectedJobsForPrompt(jobs []repositories.RejectedJobForAnalysis, maxCount int) string {
	if len(jobs) == 0 {
		return "(none)"
	}
	count := len(jobs)
	if count > maxCount {
		count = maxCount
	}
	var lines []string
	for i := 0; i < count; i++ {
		lines = append(lines, fmt.Sprintf("- %s: %s", jobs[i].JobTitle, jobs[i].Reason))
	}
	return strings.Join(lines, "\n")
}

func formatRejectedSample(rejected []reviewedJobResult, maxCount int) string {
	if len(rejected) == 0 {
		return "(none)"
	}
	count := len(rejected)
	if count > maxCount {
		count = maxCount
	}
	var lines []string
	for i := 0; i < count; i++ {
		lines = append(lines, fmt.Sprintf("- %s at %s: %s", rejected[i].Title, rejected[i].Company, rejected[i].Reason))
	}
	return strings.Join(lines, "\n")
}

func (s *AutomationService) generateSuggestionWithAnthropic(model, prompt string) string {
	if s.anthropicAPIKey == "" {
		return ""
	}

	client := anthropic.NewClient(option.WithAPIKey(s.anthropicAPIKey))
	params := anthropic.MessageNewParams{
		Model:     anthropic.Model(model),
		MaxTokens: 1024,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := client.Messages.New(ctx, params)
	if err != nil {
		log.Printf("Automation service: failed to generate suggestion: %v", err)
		return ""
	}

	if len(resp.Content) == 0 {
		return ""
	}
	text, ok := resp.Content[0].AsAny().(anthropic.TextBlock)
	if !ok {
		return ""
	}
	return strings.TrimSpace(text.Text)
}

func (s *AutomationService) generateSuggestionWithOpenAI(model, prompt string) string {
	if s.openAIAPIKey == "" {
		return ""
	}

	client := openai.NewClient(openaiOption.WithAPIKey(s.openAIAPIKey))
	params := openai.ChatCompletionNewParams{
		Model: model,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(prompt),
		},
		MaxCompletionTokens: openai.Int(1024),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := client.Chat.Completions.New(ctx, params)
	if err != nil {
		log.Printf("Automation service: failed to generate suggestion: %v", err)
		return ""
	}

	if len(resp.Choices) == 0 {
		return ""
	}
	return strings.TrimSpace(resp.Choices[0].Message.Content)
}

func (s *AutomationService) dtoToResponse(dto *repositories.AutomationConfigDTO) *serializers.AutomationConfigResponse {
	totalProposals, _ := s.automationRepo.GetActiveProposals(dto.ID)

	return &serializers.AutomationConfigResponse{
		ID:                  dto.ID,
		TenantID:            dto.TenantID,
		UserID:              dto.UserID,
		Name:                dto.Name,
		EntityType:          dto.EntityType,
		Enabled:             dto.Enabled,
		IntervalSeconds:     dto.IntervalSeconds,
		LastRunAt:           dto.LastRunAt,
		NextRunAt:           dto.NextRunAt,
		RunCount:          dto.RunCount,
		CompilationTarget: dto.CompilationTarget,
		DisableOnCompiled:   dto.DisableOnCompiled,
		ActiveProposals:     totalProposals,
		CompiledAt:          dto.CompiledAt,
		SystemPrompt:        dto.SystemPrompt,
		SuggestedPrompt:       dto.SuggestedPrompt,
		UseSuggestedPrompt:    dto.UseSuggestedPrompt,
		SuggestionThreshold:   dto.SuggestionThreshold,
		MinProspectsThreshold: dto.MinProspectsThreshold,
		SearchQuery:           dto.SearchQuery,
		SuggestedQuery:    dto.SuggestedQuery,
		UseSuggestedQuery: dto.UseSuggestedQuery,
		ATSMode:           dto.ATSMode,
		TimeFilter:          dto.TimeFilter,
		Location:            dto.Location,
		TargetType:          dto.TargetType,
		TargetIDs:           dto.TargetIDs,
		SourceDocumentIDs:   dto.SourceDocumentIDs,
		DigestEnabled:       dto.DigestEnabled,
		DigestEmails:        dto.DigestEmails,
		DigestTime:          dto.DigestTime,
		DigestModel:         dto.DigestModel,
		LastDigestAt:        dto.LastDigestAt,
		UseAgent:            dto.UseAgent,
		AgentModel:          dto.AgentModel,
		UseAnalytics:        dto.UseAnalytics,
		AnalyticsModel:      dto.AnalyticsModel,
		EmptyProposalLimit:      dto.EmptyProposalLimit,
		PromptCooldownRuns:      dto.PromptCooldownRuns,
		PromptCooldownProspects: dto.PromptCooldownProspects,
		CreatedAt:               dto.CreatedAt,
		UpdatedAt:               dto.UpdatedAt,
	}
}

// dedupData holds preloaded data for deduplication
type dedupData struct {
	mu               sync.RWMutex
	existingJobs     []repositories.JobDedupInfo
	pendingProposals []repositories.PendingJobProposal
	rejectedJobs     []repositories.RejectedJobInfo
	urlSource        map[string]string // URL -> source ("Job", "Proposal", "Rejected")
}

// loadDedupData loads existing jobs, pending proposals, and rejected jobs for deduplication (parallel)
func (s *AutomationService) loadDedupData(tenantID int64) *dedupData {
	data := &dedupData{
		urlSource: make(map[string]string),
	}

	var wg sync.WaitGroup
	var jobs []repositories.JobDedupInfo
	var proposals []repositories.PendingJobProposal
	var rejectedProposals []repositories.PendingJobProposal
	var rejected []repositories.RejectedJobInfo

	wg.Add(4)
	go func() {
		defer wg.Done()
		var err error
		jobs, err = s.jobRepo.GetJobsForDedup(tenantID)
		if err != nil {
			log.Printf("Automation service: failed to load jobs for dedup: %v", err)
		}
	}()
	go func() {
		defer wg.Done()
		var err error
		proposals, err = s.proposalRepo.GetPendingJobProposals(tenantID)
		if err != nil {
			log.Printf("Automation service: failed to load pending proposals for dedup: %v", err)
		}
	}()
	go func() {
		defer wg.Done()
		var err error
		rejectedProposals, err = s.proposalRepo.GetRejectedJobProposals(tenantID)
		if err != nil {
			log.Printf("Automation service: failed to load rejected proposals for dedup: %v", err)
		}
	}()
	go func() {
		defer wg.Done()
		var err error
		rejected, err = s.automationRepo.GetRejectedJobs(tenantID)
		if err != nil {
			log.Printf("Automation service: failed to load rejected jobs for dedup: %v", err)
		}
	}()
	wg.Wait()

	// Merge results (single-threaded, no lock needed)
	data.existingJobs = jobs
	for _, j := range jobs {
		if j.URL != "" {
			data.urlSource[j.URL] = "Job"
		}
	}
	log.Printf("🔍 Dedup [Job]: loaded %d existing jobs", len(jobs))

	data.pendingProposals = proposals
	for _, p := range proposals {
		if p.URL != "" {
			data.urlSource[p.URL] = "Proposal"
		}
	}
	log.Printf("🔍 Dedup [Proposal]: loaded %d pending proposals", len(proposals))

	for _, p := range rejectedProposals {
		if p.URL != "" {
			data.urlSource[p.URL] = "UserRejected"
		}
	}
	log.Printf("🔍 Dedup [UserRejected]: loaded %d user-rejected proposals", len(rejectedProposals))

	data.rejectedJobs = rejected
	for _, r := range rejected {
		if r.URL != "" {
			data.urlSource[r.URL] = "Rejected"
		}
	}
	log.Printf("🔍 Dedup [Rejected]: loaded %d agent-rejected jobs", len(rejected))

	log.Printf("✅ Dedup ready: %d total URLs to check against", len(data.urlSource))
	return data
}

// duplicateSource returns the source of the duplicate ("Job", "Proposal", "Rejected") or empty if not duplicate
func (d *dedupData) duplicateSource(job jobResult) string {
	d.mu.RLock()
	defer d.mu.RUnlock()

	// Check URL first (fast path)
	if job.URL != "" {
		if source, ok := d.urlSource[job.URL]; ok {
			return source
		}
	}

	// Check against existing jobs by company + title
	normalizedCompany := normalizeCompanyName(job.Company)
	normalizedTitle := normalizeTitleForMatch(job.Title)

	if len(normalizedCompany) >= 4 {
		for _, existing := range d.existingJobs {
			existingCompany := normalizeCompanyName(existing.Company)
			existingTitle := normalizeTitleForMatch(existing.JobTitle)

			if existingCompany == "" || len(existingCompany) < 4 {
				continue
			}

			companyMatch := normalizedCompany == existingCompany ||
				strings.Contains(normalizedCompany, existingCompany) ||
				strings.Contains(existingCompany, normalizedCompany)

			if companyMatch && (normalizedTitle == existingTitle ||
				strings.Contains(normalizedTitle, existingTitle) ||
				strings.Contains(existingTitle, normalizedTitle)) {
				return "Job"
			}
		}

		// Check against pending proposals by company + title
		for _, pending := range d.pendingProposals {
			pendingCompany := normalizeCompanyName(pending.Company)
			pendingTitle := normalizeTitleForMatch(pending.Title)

			if pendingCompany == "" || len(pendingCompany) < 4 {
				continue
			}

			companyMatch := normalizedCompany == pendingCompany ||
				strings.Contains(normalizedCompany, pendingCompany) ||
				strings.Contains(pendingCompany, normalizedCompany)

			if companyMatch && (normalizedTitle == pendingTitle ||
				strings.Contains(normalizedTitle, pendingTitle) ||
				strings.Contains(pendingTitle, normalizedTitle)) {
				return "Proposal"
			}
		}
	}

	return ""
}

// markURL adds a URL to the dedup set (thread-safe)
func (d *dedupData) markURL(url, source string) {
	if url == "" {
		return
	}
	d.mu.Lock()
	d.urlSource[url] = source
	d.mu.Unlock()
}

// normalizeCompanyName removes common suffixes and normalizes for comparison
func normalizeCompanyName(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	suffixes := []string{", inc.", ", inc", " inc.", " inc", ", llc", " llc", ", corp.", " corp.", ", corp", " corp", ", ltd.", " ltd.", ", ltd", " ltd", " co.", " co"}
	for _, suffix := range suffixes {
		name = strings.TrimSuffix(name, suffix)
	}
	return strings.TrimSpace(name)
}

// normalizeTitleForMatch removes seniority modifiers for fuzzy matching
func normalizeTitleForMatch(title string) string {
	title = strings.ToLower(strings.TrimSpace(title))
	prefixes := []string{"senior ", "sr. ", "sr ", "lead ", "principal ", "staff ", "junior ", "jr. ", "jr ", "associate ", "entry level ", "entry-level "}
	for _, prefix := range prefixes {
		title = strings.TrimPrefix(title, prefix)
	}
	return strings.TrimSpace(title)
}

// filterDuplicates removes jobs that already exist in the dedup data
func (s *AutomationService) filterDuplicates(results []jobResult, dedup *dedupData) []jobResult {
	if dedup == nil {
		return results
	}
	filtered := make([]jobResult, 0, len(results))
	for _, job := range results {
		if dupSource := dedup.duplicateSource(job); dupSource != "" {
			log.Printf("Automation: filtering duplicate [%s]: %s at %s", dupSource, job.Title, job.Company)
			continue
		}
		// Mark URL immediately to prevent duplicates across concurrent batches
		dedup.markURL(job.URL, "Pending")
		filtered = append(filtered, job)
	}
	return filtered
}

// validateURLs checks URLs with HEAD requests to filter broken links (concurrent)
func (s *AutomationService) validateURLs(results []jobResult) []jobResult {
	if len(results) == 0 {
		return results
	}

	const maxConcurrent = 5
	sem := make(chan struct{}, maxConcurrent)
	var wg sync.WaitGroup
	var mu sync.Mutex
	valid := make([]jobResult, 0, len(results))

	client := &http.Client{
		Timeout: 5 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return nil // follow redirects
		},
	}

	for _, r := range results {
		wg.Add(1)
		go func(job jobResult) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			if isURLValid(client, job.URL) {
				mu.Lock()
				valid = append(valid, job)
				mu.Unlock()
			}
		}(r)
	}
	wg.Wait()

	if len(valid) < len(results) {
		log.Printf("Automation: URL validation filtered %d/%d broken links", len(results)-len(valid), len(results))
	}

	return valid
}

func isURLValid(client *http.Client, url string) bool {
	resp, err := client.Head(url)
	if err != nil {
		log.Printf("Automation: URL validation failed for %s: %v", url, err)
		return false
	}
	resp.Body.Close()

	if resp.StatusCode >= 400 {
		log.Printf("Automation: skipping broken link %s (status %d)", url, resp.StatusCode)
		return false
	}
	return true
}

// extractExclusions pulls out -term patterns from a search query
func extractExclusions(query string) []string {
	var exclusions []string
	seen := make(map[string]bool)
	words := strings.Fields(query)
	for _, word := range words {
		if !strings.HasPrefix(word, "-") {
			continue
		}
		if seen[word] {
			continue
		}
		seen[word] = true
		exclusions = append(exclusions, word)
	}
	return exclusions
}

func (s *AutomationService) sendPushover(title, message string) {
	if s.pushoverUser == "" || s.pushoverToken == "" {
		return
	}
	if s.env == "development" {
		return
	}

	form := url.Values{
		"token":   {s.pushoverToken},
		"user":    {s.pushoverUser},
		"title":   {title},
		"message": {message},
	}

	resp, err := http.PostForm("https://api.pushover.net/1/messages.json", form)
	if err != nil {
		log.Printf("Pushover error: %v", err)
		return
	}
	defer resp.Body.Close()
}
