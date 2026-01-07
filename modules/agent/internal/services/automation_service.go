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
	proposalService ProposalCreator,
	docLoader DocumentLoader,
	serperAPIKey string,
	anthropicAPIKey string,
	openAIAPIKey string,
) *AutomationService {
	return &AutomationService{
		automationRepo:  automationRepo,
		proposalRepo:    proposalRepo,
		proposalService: proposalService,
		docLoader:       docLoader,
		serperAPIKey:    serperAPIKey,
		anthropicAPIKey: anthropicAPIKey,
		openAIAPIKey:    openAIAPIKey,
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
	cfg, err := s.automationRepo.UpdateConfig(configID, input)
	if err != nil {
		return nil, err
	}
	return s.dtoToResponse(cfg), nil
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

// GetCrawlerRuns returns recent run logs for a crawler
func (s *AutomationService) GetCrawlerRuns(configID int64, limit int) ([]serializers.AutomationRunLogResponse, error) {
	logs, err := s.automationRepo.GetRunLogs(configID, limit)
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
			LeadsFound:       logs[i].LeadsFound,
			ProposalsCreated: logs[i].ProposalsCreated,
			ErrorMessage:     logs[i].ErrorMessage,
			SearchQuery:      logs[i].SearchQuery,
		}
	}
	return result, nil
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
		log.Printf("Automation service: config=%d compilation target reached, skipping", cfg.ID)
		return
	}

	log.Printf("Automation service: config=%d needs %d more proposals", cfg.ID, deficit)

	keywords := cfg.SearchKeywords
	if len(keywords) == 0 {
		log.Printf("Automation service: no search keywords for config %d", cfg.ID)
		return
	}

	combinedQuery := strings.Join(keywords, ", ")
	logID, err := s.automationRepo.CreateRunLog(cfg.ID, combinedQuery)
	if err != nil {
		log.Printf("Automation service: failed to create run log: %v", err)
		return
	}

	keywordBatches := s.buildSearchQueries(cfg)
	totalLeads, totalProposals := s.executeSearches(cfg, keywordBatches, deficit, combinedQuery)

	s.automationRepo.CompleteRunLog(logID, "completed", int(totalLeads), int(totalProposals), nil)

	nextRun := time.Now().Add(time.Duration(cfg.IntervalMinutes) * time.Minute)
	s.automationRepo.UpdateLastRun(cfg.ID, time.Now(), nextRun)

	log.Printf("Automation service: completed config=%d - found=%d, proposals=%d",
		cfg.ID, totalLeads, totalProposals)
}

func (s *AutomationService) logAutomationConfig(cfg *repositories.AutomationConfigDTO) {
	log.Printf("Automation service: === Starting config=%d '%s' ===", cfg.ID, cfg.Name)
	log.Printf("  entity_type=%s, enabled=%v", cfg.EntityType, cfg.Enabled)
	log.Printf("  interval_minutes=%d, leads_per_run=%d, concurrent_searches=%d", cfg.IntervalMinutes, cfg.LeadsPerRun, cfg.ConcurrentSearches)
	log.Printf("  compilation_target=%d, ats_mode=%v, time_filter=%v", cfg.CompilationTarget, cfg.ATSMode, cfg.TimeFilter)
	log.Printf("  search_keywords=%v", cfg.SearchKeywords)
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

func (s *AutomationService) executeSearches(cfg *repositories.AutomationConfigDTO, batches [][]string, deficit int64, combinedQuery string) (int32, int32) {
	var searchWg sync.WaitGroup
	var totalLeadsFound int32
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
		return s.processSearchResultsSequential(cfg, resultsChan, &totalLeadsFound, deficit, combinedQuery)
	}
	return s.processSearchResultsWithConcurrentAgent(cfg, resultsChan, &totalLeadsFound, deficit, combinedQuery)
}

func (s *AutomationService) processSearchResultsSequential(cfg *repositories.AutomationConfigDTO, resultsChan <-chan []jobResult, totalLeadsFound *int32, deficit int64, combinedQuery string) (int32, int32) {
	var totalProposalsCreated int32
	source := "automation"
	configID := cfg.ID

	for results := range resultsChan {
		atomic.AddInt32(totalLeadsFound, int32(len(results)))
		created := s.processRawResults(cfg, results, combinedQuery, source, configID, deficit, &totalProposalsCreated)
		atomic.AddInt32(&totalProposalsCreated, int32(created))
	}

	return *totalLeadsFound, totalProposalsCreated
}

func (s *AutomationService) processSearchResultsWithConcurrentAgent(cfg *repositories.AutomationConfigDTO, resultsChan <-chan []jobResult, totalLeadsFound *int32, deficit int64, combinedQuery string) (int32, int32) {
	documentContext := s.loadDocumentContext(cfg.SourceDocumentIDs)

	var reviewWg sync.WaitGroup
	reviewedChan := make(chan []reviewedJobResult, 10)

	// Spawn concurrent agent review goroutines for each batch
	for results := range resultsChan {
		atomic.AddInt32(totalLeadsFound, int32(len(results)))
		reviewWg.Add(1)
		go s.agentReviewWorker(cfg, results, documentContext, reviewedChan, &reviewWg)
	}

	go func() {
		reviewWg.Wait()
		close(reviewedChan)
	}()

	return s.collectAndCreateProposals(cfg, reviewedChan, totalLeadsFound, deficit, combinedQuery)
}

func (s *AutomationService) agentReviewWorker(cfg *repositories.AutomationConfigDTO, results []jobResult, documentContext string, reviewedChan chan<- []reviewedJobResult, wg *sync.WaitGroup) {
	defer wg.Done()
	reviewed := s.reviewResultsWithAgent(cfg, results, documentContext)
	reviewedChan <- reviewed
}

func (s *AutomationService) collectAndCreateProposals(cfg *repositories.AutomationConfigDTO, reviewedChan <-chan []reviewedJobResult, totalLeadsFound *int32, deficit int64, combinedQuery string) (int32, int32) {
	var totalProposalsCreated int32
	source := "automation"
	configID := cfg.ID

	for reviewed := range reviewedChan {
		created := s.createProposalsFromReviewed(cfg, reviewed, combinedQuery, source, configID, deficit, &totalProposalsCreated)
		atomic.AddInt32(&totalProposalsCreated, int32(created))
	}

	return *totalLeadsFound, totalProposalsCreated
}

func (s *AutomationService) createProposalsFromReviewed(cfg *repositories.AutomationConfigDTO, reviewed []reviewedJobResult, combinedQuery, source string, configID, deficit int64, totalCreated *int32) int {
	created := 0
	for _, job := range reviewed {
		if s.deficitReached(totalCreated, created, deficit) {
			return created
		}
		created += s.tryCreateApprovedProposal(cfg, job, combinedQuery, source, configID)
	}
	return created
}

func (s *AutomationService) processRawResults(cfg *repositories.AutomationConfigDTO, results []jobResult, combinedQuery, source string, configID, deficit int64, totalCreated *int32) int {
	created := 0
	for _, job := range results {
		if s.deficitReached(totalCreated, created, deficit) {
			return created
		}
		created += s.tryCreateProposal(cfg, job, "", combinedQuery, source, configID)
	}
	return created
}

func (s *AutomationService) deficitReached(totalCreated *int32, created int, deficit int64) bool {
	return int64(atomic.LoadInt32(totalCreated))+int64(created) >= deficit
}

func (s *AutomationService) tryCreateProposal(cfg *repositories.AutomationConfigDTO, job jobResult, matchReason, combinedQuery, source string, configID int64) int {
	if !s.createProposalFromJob(cfg, job, matchReason, combinedQuery, source, configID) {
		return 0
	}
	return 1
}

func (s *AutomationService) tryCreateApprovedProposal(cfg *repositories.AutomationConfigDTO, job reviewedJobResult, combinedQuery, source string, configID int64) int {
	if !job.Approved {
		return 0
	}
	return s.tryCreateProposal(cfg, job.jobResult, job.Reason, combinedQuery, source, configID)
}

func (s *AutomationService) searchWorker(workerID int, keywords []string, cfg *repositories.AutomationConfigDTO, results chan<- []jobResult, wg *sync.WaitGroup) {
	defer wg.Done()

	query := strings.Join(keywords, " ")
	log.Printf("Automation service[%d]: searching for '%s'", workerID, query)

	jobs, err := s.searchJobs(query, cfg.LeadsPerRun, cfg.ATSMode, cfg.TimeFilter)
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
	keywords := cfg.SearchKeywords
	if len(keywords) == 0 {
		return nil
	}

	if len(cfg.SearchSlots) == 0 {
		return s.autoDistributeKeywords(cfg)
	}

	if len(cfg.SearchSlots) == 1 && len(cfg.SearchSlots[0]) == 0 {
		return [][]string{keywords}
	}

	batches := s.buildBatchesFromSlots(cfg.SearchSlots, keywords)
	if len(batches) == 0 {
		return s.autoDistributeKeywords(cfg)
	}
	return batches
}

func (s *AutomationService) autoDistributeKeywords(cfg *repositories.AutomationConfigDTO) [][]string {
	concurrency := cfg.ConcurrentSearches
	if concurrency < 1 {
		concurrency = 1
	}
	return s.distributeKeywords(cfg.SearchKeywords, concurrency)
}

func (s *AutomationService) distributeKeywords(keywords []string, n int) [][]string {
	if n <= 0 {
		n = 1
	}
	if n > len(keywords) {
		n = len(keywords)
	}

	batches := make([][]string, n)
	for i, kw := range keywords {
		batches[i%n] = append(batches[i%n], kw)
	}

	var result [][]string
	for _, batch := range batches {
		result = appendNonEmpty(result, batch)
	}
	return result
}

func (s *AutomationService) buildBatchesFromSlots(slots [][]int, keywords []string) [][]string {
	var batches [][]string
	for _, slot := range slots {
		batch := s.buildBatchFromSlot(slot, keywords)
		batches = appendNonEmpty(batches, batch)
	}
	return batches
}

func (s *AutomationService) buildBatchFromSlot(slot []int, keywords []string) []string {
	var batch []string
	for _, idx := range slot {
		batch = appendIfValidIndex(batch, idx, keywords)
	}
	return batch
}

func appendNonEmpty(batches [][]string, batch []string) [][]string {
	if len(batch) == 0 {
		return batches
	}
	return append(batches, batch)
}

func appendIfValidIndex(batch []string, idx int, keywords []string) []string {
	if idx < 0 || idx >= len(keywords) {
		return batch
	}
	return append(batch, keywords[idx])
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

	return s.executeSerperRequest(jsonBody)
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

	resp, err := client.Messages.New(ctx, params)
	if err != nil {
		log.Printf("Automation service: Anthropic agent review failed: %v", err)
		return approveAllResults(results)
	}

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

	resp, err := client.Chat.Completions.New(ctx, params)
	if err != nil {
		log.Printf("Automation service: OpenAI agent review failed: %v", err)
		return approveAllResults(results)
	}

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
	totalProposals, _ := s.automationRepo.GetTotalProposalsCreated(dto.ID)

	return &serializers.AutomationConfigResponse{
		ID:                    dto.ID,
		TenantID:              dto.TenantID,
		UserID:                dto.UserID,
		Name:                  dto.Name,
		EntityType:            dto.EntityType,
		Enabled:               dto.Enabled,
		IntervalMinutes:       dto.IntervalMinutes,
		LastRunAt:             dto.LastRunAt,
		NextRunAt:             dto.NextRunAt,
		RunCount:              dto.RunCount,
		LeadsPerRun:           dto.LeadsPerRun,
		ConcurrentSearches:    dto.ConcurrentSearches,
		CompilationTarget:     dto.CompilationTarget,
		TotalProposalsCreated: totalProposals,
		SystemPrompt:          dto.SystemPrompt,
		SearchKeywords:        dto.SearchKeywords,
		SearchSlots:           dto.SearchSlots,
		ATSMode:               dto.ATSMode,
		TimeFilter:            dto.TimeFilter,
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
