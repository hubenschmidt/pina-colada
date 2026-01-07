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
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/ledongthuc/pdf"
	"github.com/openai/openai-go/v2"
	openaiOption "github.com/openai/openai-go/v2/option"

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

// GetCrawlers returns all crawlers for a user
func (s *AutomationService) GetCrawlers(tenantID, userID int64) ([]serializers.AutomationConfigResponse, error) {
	configs, err := s.automationRepo.GetUserConfigs(tenantID, userID)
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

	// Set next run time if enabling - run immediately
	now := time.Now()
	result, err := s.automationRepo.UpdateConfig(configID, input)
	if err != nil {
		return nil, err
	}

	_ = s.automationRepo.SetNextRun(result.ID, now)
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

	go s.executeAutomation(cfg)
	return nil
}

func (s *AutomationService) executeAutomation(cfg *repositories.AutomationConfigDTO) {
	s.logAutomationConfig(cfg)

	deficit := s.calculateDeficit(cfg)
	if deficit <= 0 {
		s.pauseCrawler(cfg)
		return
	}

	log.Printf("Automation service: config=%d needs %d more proposals", cfg.ID, deficit)

	if len(cfg.SearchSlots) == 0 {
		log.Printf("Automation service: no search slots for config %d", cfg.ID)
		return
	}

	// Load dedup data before starting searches
	dedup := s.loadDedupData(cfg.TenantID)

	combinedQuery := s.buildCombinedQuery(cfg.SearchSlots)
	logID, err := s.automationRepo.CreateRunLog(cfg.ID, combinedQuery)
	if err != nil {
		log.Printf("Automation service: failed to create run log: %v", err)
		return
	}

	// Notify SSE subscribers that run started
	now := time.Now()
	sse.Publish(sse.CrawlerTopic(cfg.ID), sse.Event{
		Type: "run_started",
		Data: serializers.AutomationRunLogResponse{
			ID:          logID,
			StartedAt:   now,
			Status:      "running",
			SearchQuery: &combinedQuery,
		},
	})

	keywordBatches := s.buildSearchQueries(cfg)
	totalProspects, totalProposals, suggestedQueries := s.executeSearches(cfg, keywordBatches, deficit, combinedQuery, dedup)

	// Check if this run crossed the compilation threshold (from below to at-or-above)
	currentActive, _ := s.automationRepo.GetActiveProposals(cfg.ID)
	previousActive := currentActive - int(totalProposals)
	compiled := previousActive < cfg.CompilationTarget && currentActive >= cfg.CompilationTarget

	// Disable crawler when compilation target is reached (if configured)
	shouldDisable := compiled && cfg.DisableOnCompiled
	if shouldDisable {
		s.automationRepo.DisableConfig(cfg.ID)
	}

	// Format suggested queries for storage
	var suggestedQueriesStr *string
	if len(suggestedQueries) > 0 {
		joined := strings.Join(suggestedQueries, "; ")
		suggestedQueriesStr = &joined
	}

	s.automationRepo.CompleteRunLog(logID, "done", int(totalProspects), int(totalProposals), nil, suggestedQueriesStr, compiled)

	// Notify SSE subscribers that run completed
	completedAt := time.Now()
	sseEvent := serializers.AutomationRunLogResponse{
		ID:                    logID,
		StartedAt:             now,
		CompletedAt:           &completedAt,
		Status:                "done",
		ProspectsFound:        int(totalProspects),
		ProposalsCreated:      int(totalProposals),
		SearchQuery:           &combinedQuery,
		SuggestedQueries:      suggestedQueriesStr,
		Compiled:              compiled,
		ConfigActiveProposals: &currentActive,
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
	log.Printf("  interval_seconds=%d, prospects_per_run=%d, concurrent_searches=%d", cfg.IntervalSeconds, cfg.ProspectsPerRun, cfg.ConcurrentSearches)
	log.Printf("  compilation_target=%d, ats_mode=%v, time_filter=%v", cfg.CompilationTarget, cfg.ATSMode, cfg.TimeFilter)
	log.Printf("  search_slots=%v", cfg.SearchSlots)
	log.Printf("  target_type=%v, target_ids=%v", cfg.TargetType, cfg.TargetIDs)
	log.Printf("  source_document_ids=%v", cfg.SourceDocumentIDs)
	log.Printf("  use_agent=%v, agent_model=%v", cfg.UseAgent, cfg.AgentModel)
	log.Printf("  system_prompt=%v", cfg.SystemPrompt)
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

func (s *AutomationService) executeSearches(cfg *repositories.AutomationConfigDTO, batches [][]string, deficit int64, combinedQuery string, dedup *dedupData) (int32, int32, []string) {
	var searchWg sync.WaitGroup
	var totalProspectsFound int32
	resultsChan := make(chan []jobResult, len(batches))

	for i, batch := range batches {
		searchWg.Add(1)
		go s.searchWorker(i, batch, cfg, resultsChan, &searchWg)
	}

	go func() {
		searchWg.Wait()
		close(resultsChan)
	}()

	if !cfg.UseAgent {
		prospects, proposals := s.processSearchResultsSequential(cfg, resultsChan, &totalProspectsFound, deficit, combinedQuery, dedup)
		return prospects, proposals, nil
	}
	return s.processSearchResultsWithConcurrentAgent(cfg, resultsChan, &totalProspectsFound, deficit, combinedQuery, dedup)
}

func (s *AutomationService) processSearchResultsSequential(cfg *repositories.AutomationConfigDTO, resultsChan <-chan []jobResult, totalProspectsFound *int32, deficit int64, combinedQuery string, dedup *dedupData) (int32, int32) {
	var totalProposalsCreated int32
	source := "automation"
	configID := cfg.ID

	for results := range resultsChan {
		atomic.AddInt32(totalProspectsFound, int32(len(results)))
		created := s.processRawResults(cfg, results, combinedQuery, source, configID, deficit, &totalProposalsCreated, dedup)
		atomic.AddInt32(&totalProposalsCreated, int32(created))
	}

	return *totalProspectsFound, totalProposalsCreated
}

func (s *AutomationService) processSearchResultsWithConcurrentAgent(cfg *repositories.AutomationConfigDTO, resultsChan <-chan []jobResult, totalProspectsFound *int32, deficit int64, combinedQuery string, dedup *dedupData) (int32, int32, []string) {
	documentContext := s.loadDocumentContext(cfg.SourceDocumentIDs)

	var reviewWg sync.WaitGroup
	reviewedChan := make(chan []reviewedJobResult, 10)
	querySuggestionsUsed := false
	var suggestedQueries []string

	// Spawn concurrent agent review goroutines for each batch
	for results := range resultsChan {
		atomic.AddInt32(totalProspectsFound, int32(len(results)))

		// Filter duplicates BEFORE sending to LLM
		filtered := s.filterDuplicates(results, dedup)

		// Has new results - send to LLM review
		if len(filtered) > 0 {
			log.Printf("Automation: %d/%d results are new, sending to LLM review", len(filtered), len(results))
			reviewWg.Add(1)
			go s.agentReviewWorker(cfg, filtered, documentContext, reviewedChan, &reviewWg)
			continue
		}

		// All duplicates
		log.Printf("Automation: all %d results are duplicates, skipping LLM review", len(results))

		// Already tried query suggestions this run
		if querySuggestionsUsed {
			continue
		}

		// Ask LLM for new query suggestions (once per run)
		querySuggestionsUsed = true
		suggestedQueries = s.searchWithSuggestedQueries(cfg, documentContext, dedup, &reviewWg, reviewedChan)
	}

	go func() {
		reviewWg.Wait()
		close(reviewedChan)
	}()

	prospectsFound, proposalsCreated := s.collectAndCreateProposals(cfg, reviewedChan, totalProspectsFound, deficit, combinedQuery, dedup)
	return prospectsFound, proposalsCreated, suggestedQueries
}

func (s *AutomationService) agentReviewWorker(cfg *repositories.AutomationConfigDTO, results []jobResult, documentContext string, reviewedChan chan<- []reviewedJobResult, wg *sync.WaitGroup) {
	defer wg.Done()
	reviewed := s.reviewResultsWithAgent(cfg, results, documentContext)
	reviewedChan <- reviewed
}

func (s *AutomationService) collectAndCreateProposals(cfg *repositories.AutomationConfigDTO, reviewedChan <-chan []reviewedJobResult, totalProspectsFound *int32, deficit int64, combinedQuery string, dedup *dedupData) (int32, int32) {
	var totalProposalsCreated int32
	source := "automation"
	configID := cfg.ID

	for reviewed := range reviewedChan {
		created := s.createProposalsFromReviewed(cfg, reviewed, combinedQuery, source, configID, deficit, &totalProposalsCreated, dedup)
		atomic.AddInt32(&totalProposalsCreated, int32(created))
	}

	return *totalProspectsFound, totalProposalsCreated
}

func (s *AutomationService) createProposalsFromReviewed(cfg *repositories.AutomationConfigDTO, reviewed []reviewedJobResult, combinedQuery, source string, configID, deficit int64, totalCreated *int32, dedup *dedupData) int {
	created := 0
	for _, job := range reviewed {
		if s.deficitReached(totalCreated, created, deficit) {
			return created
		}
		created += s.tryCreateApprovedProposal(cfg, job, combinedQuery, source, configID, dedup)
	}
	return created
}

func (s *AutomationService) processRawResults(cfg *repositories.AutomationConfigDTO, results []jobResult, combinedQuery, source string, configID, deficit int64, totalCreated *int32, dedup *dedupData) int {
	created := 0
	for _, job := range results {
		if s.deficitReached(totalCreated, created, deficit) {
			return created
		}
		created += s.tryCreateProposal(cfg, job, "", combinedQuery, source, configID, dedup)
	}
	return created
}

func (s *AutomationService) deficitReached(totalCreated *int32, created int, deficit int64) bool {
	return int64(atomic.LoadInt32(totalCreated))+int64(created) >= deficit
}

func (s *AutomationService) tryCreateProposal(cfg *repositories.AutomationConfigDTO, job jobResult, matchReason, combinedQuery, source string, configID int64, dedup *dedupData) int {
	// Check for duplicate before creating
	if dedup != nil && dedup.isDuplicate(job) {
		log.Printf("Automation service: skipping duplicate job: %s at %s", job.Title, job.Company)
		return 0
	}
	if !s.createProposalFromJob(cfg, job, matchReason, combinedQuery, source, configID) {
		return 0
	}
	// Add to dedup set to prevent duplicates within the same run
	if dedup != nil && job.URL != "" {
		dedup.urlSet[job.URL] = true
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

func (s *AutomationService) searchWorker(workerID int, keywords []string, cfg *repositories.AutomationConfigDTO, results chan<- []jobResult, wg *sync.WaitGroup) {
	defer wg.Done()

	query := strings.Join(keywords, " ")
	log.Printf("Automation service[%d]: searching for '%s'", workerID, query)

	jobs, err := s.searchJobs(query, cfg.ProspectsPerRun, cfg.ATSMode, cfg.TimeFilter)
	if err != nil {
		log.Printf("Automation service[%d]: search failed: %v", workerID, err)
		return
	}

	log.Printf("Automation service[%d]: found %d results", workerID, len(jobs))
	results <- jobs
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

func (s *AutomationService) buildSearchQueries(cfg *repositories.AutomationConfigDTO) [][]string {
	var batches [][]string
	for _, slot := range cfg.SearchSlots {
		if len(slot) == 0 {
			continue
		}
		batches = append(batches, slot)
	}
	return batches
}

func (s *AutomationService) buildCombinedQuery(slots [][]string) string {
	var allKeywords []string
	for _, slot := range slots {
		allKeywords = append(allKeywords, slot...)
	}
	return strings.Join(allKeywords, ", ")
}

// jobResult represents a job search result
type jobResult struct {
	Title   string
	Company string
	URL     string
	Snippet string
}

func (s *AutomationService) searchJobs(query string, maxResults int, atsMode bool, timeFilter *string) ([]jobResult, error) {
	if s.serperAPIKey == "" {
		return nil, fmt.Errorf("SERPER_API_KEY not configured")
	}

	enhancedQuery := query
	if atsMode {
		enhancedQuery = buildATSQuery(query)
	}

	tbs := mapTimeFilter(timeFilter)

	log.Printf("🔧 [Serper] Calling API - query=%q, maxResults=%d, atsMode=%v, timeFilter=%v",
		enhancedQuery, maxResults, atsMode, timeFilter)

	reqBody := map[string]interface{}{
		"q":   enhancedQuery,
		"num": maxResults,
	}
	if tbs != "" {
		reqBody["tbs"] = tbs
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("error encoding request: %w", err)
	}

	results, err := s.executeSerperRequest(jsonBody)
	if err != nil {
		log.Printf("🔧 [Serper] API error: %v", err)
		return nil, err
	}
	log.Printf("🔧 [Serper] API returned %d results", len(results))
	return results, nil
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

func (s *AutomationService) executeSerperRequest(jsonBody []byte) ([]jobResult, error) {
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

	return s.parseSerperResponse(resp.Body)
}

func (s *AutomationService) parseSerperResponse(body io.Reader) ([]jobResult, error) {
	var serperResp struct {
		Organic []struct {
			Title   string `json:"title"`
			Link    string `json:"link"`
			Snippet string `json:"snippet"`
		} `json:"organic"`
	}

	if err := json.NewDecoder(body).Decode(&serperResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	var results []jobResult
	for _, item := range serperResp.Organic {
		results = append(results, jobResult{
			Title:   item.Title,
			Company: extractCompanyFromTitle(item.Title),
			URL:     item.Link,
			Snippet: item.Snippet,
		})
	}
	return results, nil
}

func buildATSQuery(baseQuery string) string {
	atsKeywords := []string{
		"site:greenhouse.io",
		"site:lever.co",
		"site:jobs.ashbyhq.com",
		"site:boards.eu.greenhouse.io",
	}
	return baseQuery + " (" + strings.Join(atsKeywords, " OR ") + ")"
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
	model := "claude-sonnet-4-20250514"
	if cfg.AgentModel != nil && *cfg.AgentModel != "" {
		model = *cfg.AgentModel
	}

	s.logAgentReviewContext(model, cfg.SystemPrompt, documentContext, results)

	if strings.HasPrefix(model, "gpt-") {
		return s.reviewWithOpenAI(model, cfg.SystemPrompt, documentContext, results)
	}
	return s.reviewWithAnthropic(model, cfg.SystemPrompt, documentContext, results)
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

func (s *AutomationService) dtoToResponse(dto *repositories.AutomationConfigDTO) *serializers.AutomationConfigResponse {
	totalProposals, _ := s.automationRepo.GetActiveProposals(dto.ID)

	return &serializers.AutomationConfigResponse{
		ID:                    dto.ID,
		TenantID:              dto.TenantID,
		UserID:                dto.UserID,
		Name:                  dto.Name,
		EntityType:            dto.EntityType,
		Enabled:               dto.Enabled,
		IntervalSeconds:       dto.IntervalSeconds,
		LastRunAt:             dto.LastRunAt,
		NextRunAt:             dto.NextRunAt,
		RunCount:              dto.RunCount,
		ProspectsPerRun:       dto.ProspectsPerRun,
		ConcurrentSearches:    dto.ConcurrentSearches,
		CompilationTarget:     dto.CompilationTarget,
		DisableOnCompiled:     dto.DisableOnCompiled,
		ActiveProposals:       totalProposals,
		CompiledAt:            dto.CompiledAt,
		SystemPrompt:          dto.SystemPrompt,
		SearchSlots:           dto.SearchSlots,
		ATSMode:               dto.ATSMode,
		TimeFilter:            dto.TimeFilter,
		Location:              dto.Location,
		TargetType:            dto.TargetType,
		TargetIDs:             dto.TargetIDs,
		SourceDocumentIDs:     dto.SourceDocumentIDs,
		DigestEnabled:         dto.DigestEnabled,
		DigestEmails:          dto.DigestEmails,
		DigestTime:            dto.DigestTime,
		DigestModel:           dto.DigestModel,
		LastDigestAt:          dto.LastDigestAt,
		UseAgent:              dto.UseAgent,
		AgentModel:            dto.AgentModel,
		CreatedAt:             dto.CreatedAt,
		UpdatedAt:             dto.UpdatedAt,
	}
}

// dedupData holds preloaded data for deduplication
type dedupData struct {
	existingJobs     []repositories.JobDedupInfo
	pendingProposals []repositories.PendingJobProposal
	rejectedJobs     []repositories.RejectedJobInfo
	urlSet           map[string]bool
}

// loadDedupData loads existing jobs, pending proposals, and rejected jobs for deduplication
func (s *AutomationService) loadDedupData(tenantID int64) *dedupData {
	data := &dedupData{
		urlSet: make(map[string]bool),
	}

	// Load existing jobs
	jobs, err := s.jobRepo.GetJobsForDedup(tenantID)
	if err != nil {
		log.Printf("Automation service: failed to load jobs for dedup: %v", err)
	}
	data.existingJobs = jobs
	for _, j := range jobs {
		if j.URL != "" {
			data.urlSet[j.URL] = true
		}
	}
	log.Printf("🔍 Dedup [Job]: loaded %d existing jobs", len(jobs))

	// Load pending proposals
	proposals, err := s.proposalRepo.GetPendingJobProposals(tenantID)
	if err != nil {
		log.Printf("Automation service: failed to load pending proposals for dedup: %v", err)
	}
	data.pendingProposals = proposals
	for _, p := range proposals {
		if p.URL != "" {
			data.urlSet[p.URL] = true
		}
	}
	log.Printf("🔍 Dedup [Proposal]: loaded %d pending proposals", len(proposals))

	// Load agent-rejected jobs
	rejected, err := s.automationRepo.GetRejectedJobs(tenantID)
	if err != nil {
		log.Printf("Automation service: failed to load rejected jobs for dedup: %v", err)
	}
	data.rejectedJobs = rejected
	for _, r := range rejected {
		if r.URL != "" {
			data.urlSet[r.URL] = true
		}
	}
	log.Printf("🔍 Dedup [Rejected]: loaded %d agent-rejected jobs", len(rejected))

	log.Printf("✅ Dedup ready: %d total URLs to check against", len(data.urlSet))
	return data
}

// isDuplicate checks if a job result is a duplicate
func (d *dedupData) isDuplicate(job jobResult) bool {
	// Check URL first (fast path)
	if job.URL != "" && d.urlSet[job.URL] {
		return true
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
				return true
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
				return true
			}
		}
	}

	return false
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
		if dedup.isDuplicate(job) {
			log.Printf("Automation: filtering duplicate: %s at %s", job.Title, job.Company)
			continue
		}
		filtered = append(filtered, job)
	}
	return filtered
}

// searchWithSuggestedQueries asks LLM for new queries and searches with them
// Returns the suggested queries for logging
func (s *AutomationService) searchWithSuggestedQueries(cfg *repositories.AutomationConfigDTO, docContext string, dedup *dedupData, reviewWg *sync.WaitGroup, reviewedChan chan<- []reviewedJobResult) []string {
	newQueries := s.suggestNewQueries(cfg, docContext, dedup)
	if len(newQueries) == 0 {
		return nil
	}

	for _, query := range newQueries {
		newResults, err := s.searchJobs(query, cfg.ProspectsPerRun, cfg.ATSMode, cfg.TimeFilter)
		if err != nil {
			log.Printf("Automation: suggested query search failed: %v", err)
			continue
		}

		newFiltered := s.filterDuplicates(newResults, dedup)
		if len(newFiltered) == 0 {
			log.Printf("Automation: suggested query '%s' returned %d results, all duplicates", query, len(newResults))
			continue
		}

		log.Printf("Automation: found %d new jobs from suggested query '%s'", len(newFiltered), query)
		reviewWg.Add(1)
		go s.agentReviewWorker(cfg, newFiltered, docContext, reviewedChan, reviewWg)
	}

	return newQueries
}

// suggestNewQueries asks LLM to suggest alternative job titles for search
// Returns exactly ConcurrentSearches number of queries, each with location appended
func (s *AutomationService) suggestNewQueries(cfg *repositories.AutomationConfigDTO, docContext string, dedup *dedupData) []string {
	if s.anthropicAPIKey == "" {
		log.Printf("Automation: no Anthropic API key, skipping query suggestions")
		return nil
	}

	numQueries := cfg.ConcurrentSearches
	if numQueries < 1 {
		numQueries = 1
	}

	// Build context of seen companies (limit to 15)
	seenCompanies := make([]string, 0, 15)
	for _, job := range dedup.existingJobs {
		if job.Company == "" {
			continue
		}
		seenCompanies = append(seenCompanies, job.Company)
		if len(seenCompanies) >= 15 {
			break
		}
	}

	// Extract just job titles from current search slots
	currentTitles := make([]string, 0, len(cfg.SearchSlots))
	for _, slot := range cfg.SearchSlots {
		if len(slot) > 0 {
			currentTitles = append(currentTitles, slot[0])
		}
	}

	// Build system prompt context
	systemPromptContext := ""
	if cfg.SystemPrompt != nil && *cfg.SystemPrompt != "" {
		systemPromptContext = fmt.Sprintf(`
Original search instructions:
%s
`, *cfg.SystemPrompt)
	}

	prompt := fmt.Sprintf(`Based on this resume/context:
%s
%s
Current job titles being searched:
%v

Companies already found (avoid similar):
%v

Suggest exactly %d NEW job title(s) to search for.
Requirements:
- Different from current titles but matching the skills in the resume
- Follow the original search instructions above
- Focus on: alternative job titles, niche roles, or different industries
- Return ONLY job titles, one per line
- No explanations, no numbering, no location info`, docContext, systemPromptContext, currentTitles, seenCompanies, numQueries)

	log.Printf("🤖 [Anthropic] Requesting %d query suggestions (docContext: %d chars, systemPrompt: %v)",
		numQueries, len(docContext), cfg.SystemPrompt != nil && *cfg.SystemPrompt != "")

	client := anthropic.NewClient(option.WithAPIKey(s.anthropicAPIKey))
	resp, err := client.Messages.New(context.Background(), anthropic.MessageNewParams{
		Model:     anthropic.ModelClaude3_5Haiku20241022,
		MaxTokens: 200,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
		},
	})
	if err != nil {
		log.Printf("Automation: failed to get query suggestions: %v", err)
		return nil
	}

	if len(resp.Content) == 0 {
		return nil
	}

	text := resp.Content[0].Text
	lines := strings.Split(strings.TrimSpace(text), "\n")

	// Parse job titles and append location
	location := ""
	if cfg.Location != nil {
		location = *cfg.Location
	}

	queries := make([]string, 0, numQueries)
	for _, line := range lines {
		if len(queries) >= numQueries {
			break
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Remove common prefixes
		line = strings.TrimPrefix(line, "- ")
		line = strings.TrimPrefix(line, "* ")
		if len(line) > 2 && line[1] == '.' {
			line = strings.TrimSpace(line[2:])
		}
		// Append location if configured
		if location != "" {
			line = line + " " + location
		}
		queries = append(queries, line)
	}

	log.Printf("Automation: LLM suggested %d new queries: %v", len(queries), queries)
	return queries
}
