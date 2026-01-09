package controllers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"agent/internal/middleware"
	"agent/internal/repositories"
	"agent/internal/services"
	"agent/internal/sse"
)

// TestRunner interface for running automation tests
type TestRunner interface {
	ExecuteForConfig(configID int64) error
}

// DigestTester interface for sending test digests
type DigestTester interface {
	SendTestDigest(configID int64) error
}

// AutomationController handles automation HTTP requests
type AutomationController struct {
	automationService *services.AutomationService
	testRunner        TestRunner
	digestTester      DigestTester
}

// NewAutomationController creates a new automation controller
func NewAutomationController(automationService *services.AutomationService, testRunner TestRunner, digestTester DigestTester) *AutomationController {
	return &AutomationController{
		automationService: automationService,
		testRunner:        testRunner,
		digestTester:      digestTester,
	}
}

// CrawlerRequest represents the request body for creating/updating a crawler
type CrawlerRequest struct {
	Name               *string  `json:"name,omitempty"`
	EntityType         *string  `json:"entity_type,omitempty"`
	Enabled            *bool    `json:"enabled,omitempty"`
	IntervalSeconds    *int     `json:"interval_seconds,omitempty"`
	ConcurrentSearches *int     `json:"concurrent_searches,omitempty"`
	CompilationTarget  *int     `json:"compilation_target,omitempty"`
	DisableOnCompiled  *bool    `json:"disable_on_compiled,omitempty"`
	SystemPrompt       *string    `json:"system_prompt,omitempty"`
	SearchSlots        []string `json:"search_slots,omitempty"`
	ATSMode            *bool      `json:"ats_mode,omitempty"`
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
	EmptyProposalLimit *int     `json:"empty_proposal_limit,omitempty"`
}

// ToggleRequest represents the request body for toggling a crawler
type ToggleRequest struct {
	Enabled bool `json:"enabled"`
}

// GetCrawlers handles GET /automation/crawlers
func (c *AutomationController) GetCrawlers(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		writeAutomationError(w, http.StatusUnauthorized, "tenant not found")
		return
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeAutomationError(w, http.StatusUnauthorized, "user not found")
		return
	}

	result, err := c.automationService.GetCrawlers(tenantID, userID)
	if err != nil {
		writeAutomationError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeAutomationJSON(w, http.StatusOK, map[string]interface{}{"crawlers": result})
}

// CreateCrawler handles POST /automation/crawlers
func (c *AutomationController) CreateCrawler(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		writeAutomationError(w, http.StatusUnauthorized, "tenant not found")
		return
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeAutomationError(w, http.StatusUnauthorized, "user not found")
		return
	}

	var req CrawlerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAutomationError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	input := c.requestToInput(req)

	result, err := c.automationService.CreateCrawler(tenantID, userID, input)
	if err != nil {
		writeAutomationError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeAutomationJSON(w, http.StatusCreated, result)
}

// GetCrawler handles GET /automation/crawlers/{id}
func (c *AutomationController) GetCrawler(w http.ResponseWriter, r *http.Request) {
	configID, err := c.parseConfigID(r)
	if err != nil {
		writeAutomationError(w, http.StatusBadRequest, "invalid crawler ID")
		return
	}

	result, err := c.automationService.GetCrawler(configID)
	if err == services.ErrCrawlerNotFound {
		writeAutomationError(w, http.StatusNotFound, "crawler not found")
		return
	}
	if err != nil {
		writeAutomationError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeAutomationJSON(w, http.StatusOK, result)
}

// UpdateCrawler handles PUT /automation/crawlers/{id}
func (c *AutomationController) UpdateCrawler(w http.ResponseWriter, r *http.Request) {
	configID, err := c.parseConfigID(r)
	if err != nil {
		writeAutomationError(w, http.StatusBadRequest, "invalid crawler ID")
		return
	}

	var req CrawlerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAutomationError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	input := c.requestToInput(req)

	result, err := c.automationService.UpdateCrawler(configID, input)
	if err != nil {
		writeAutomationError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeAutomationJSON(w, http.StatusOK, result)
}

// DeleteCrawler handles DELETE /automation/crawlers/{id}
func (c *AutomationController) DeleteCrawler(w http.ResponseWriter, r *http.Request) {
	configID, err := c.parseConfigID(r)
	if err != nil {
		writeAutomationError(w, http.StatusBadRequest, "invalid crawler ID")
		return
	}

	if err := c.automationService.DeleteCrawler(configID); err != nil {
		writeAutomationError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ClearRejectedJobs handles DELETE /automation/crawlers/{id}/rejected-jobs
func (c *AutomationController) ClearRejectedJobs(w http.ResponseWriter, r *http.Request) {
	configID, err := c.parseConfigID(r)
	if err != nil {
		writeAutomationError(w, http.StatusBadRequest, "invalid crawler ID")
		return
	}

	if err := c.automationService.ClearRejectedJobs(configID); err != nil {
		writeAutomationError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeAutomationJSON(w, http.StatusOK, map[string]string{"status": "rejected jobs cleared"})
}

// ToggleCrawler handles POST /automation/crawlers/{id}/toggle
func (c *AutomationController) ToggleCrawler(w http.ResponseWriter, r *http.Request) {
	configID, err := c.parseConfigID(r)
	if err != nil {
		writeAutomationError(w, http.StatusBadRequest, "invalid crawler ID")
		return
	}

	var req ToggleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAutomationError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := c.automationService.ToggleCrawler(configID, req.Enabled)
	if err == services.ErrCrawlerNotFound {
		writeAutomationError(w, http.StatusNotFound, "crawler not found")
		return
	}
	if err != nil {
		writeAutomationError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeAutomationJSON(w, http.StatusOK, result)
}

// GetCrawlerRuns handles GET /automation/crawlers/{id}/runs
func (c *AutomationController) GetCrawlerRuns(w http.ResponseWriter, r *http.Request) {
	configID, err := c.parseConfigID(r)
	if err != nil {
		writeAutomationError(w, http.StatusBadRequest, "invalid crawler ID")
		return
	}

	page := parseIntQueryParam(r, "page", 1)
	limit := parseIntQueryParam(r, "limit", 10)

	result, err := c.automationService.GetCrawlerRuns(configID, page, limit)
	if err != nil {
		writeAutomationError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeAutomationJSON(w, http.StatusOK, result)
}

// TestCrawler handles POST /automation/crawlers/{id}/test
func (c *AutomationController) TestCrawler(w http.ResponseWriter, r *http.Request) {
	configID, err := c.parseConfigID(r)
	if err != nil {
		writeAutomationError(w, http.StatusBadRequest, "invalid crawler ID")
		return
	}

	if c.testRunner == nil {
		writeAutomationError(w, http.StatusServiceUnavailable, "automation worker not configured")
		return
	}

	log.Printf("Automation: test run requested for crawler=%d", configID)

	if err := c.testRunner.ExecuteForConfig(configID); err != nil {
		log.Printf("Automation: test run failed: %v", err)
		writeAutomationError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeAutomationJSON(w, http.StatusOK, map[string]string{"status": "test run initiated"})
}

// SendTestDigest handles POST /automation/crawlers/{id}/test-digest
func (c *AutomationController) SendTestDigest(w http.ResponseWriter, r *http.Request) {
	configID, err := c.parseConfigID(r)
	if err != nil {
		writeAutomationError(w, http.StatusBadRequest, "invalid crawler ID")
		return
	}

	if c.digestTester == nil {
		writeAutomationError(w, http.StatusServiceUnavailable, "digest service not configured")
		return
	}

	if err := c.digestTester.SendTestDigest(configID); err != nil {
		writeAutomationError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeAutomationJSON(w, http.StatusOK, map[string]string{"status": "test digest sent"})
}

// StreamCrawlerRuns handles GET /automation/crawlers/{id}/runs/stream (SSE)
func (c *AutomationController) StreamCrawlerRuns(w http.ResponseWriter, r *http.Request) {
	configID, err := c.parseConfigID(r)
	if err != nil {
		writeAutomationError(w, http.StatusBadRequest, "invalid crawler ID")
		return
	}

	sw := sse.NewWriter(w)
	if sw == nil {
		writeAutomationError(w, http.StatusInternalServerError, "streaming not supported")
		return
	}

	// Disable timeouts for SSE streaming
	rc := http.NewResponseController(w)
	_ = rc.SetWriteDeadline(time.Time{})
	_ = rc.SetReadDeadline(time.Time{})

	page := parseIntQueryParam(r, "page", 1)
	limit := parseIntQueryParam(r, "limit", 10)

	// Send initial data
	result, err := c.automationService.GetCrawlerRuns(configID, page, limit)
	if err != nil {
		_ = sw.Send(sse.Event{Type: "error", Data: map[string]string{"error": err.Error()}})
		return
	}
	_ = sw.Send(sse.Event{Type: "init", Data: result})

	// Subscribe to updates
	topic := sse.CrawlerTopic(configID)
	eventCh := sse.Subscribe(topic)
	defer sse.Unsubscribe(topic, eventCh)

	sse.StreamWithKeepAlive(r.Context(), sw, 10*time.Second, eventCh)
}

func (c *AutomationController) parseConfigID(r *http.Request) (int64, error) {
	idStr := chi.URLParam(r, "id")
	return strconv.ParseInt(idStr, 10, 64)
}

func (c *AutomationController) requestToInput(req CrawlerRequest) repositories.AutomationConfigInput {
	return repositories.AutomationConfigInput{
		Name:               req.Name,
		EntityType:         req.EntityType,
		Enabled:            req.Enabled,
		IntervalSeconds:    req.IntervalSeconds,
		ConcurrentSearches: req.ConcurrentSearches,
		CompilationTarget:  req.CompilationTarget,
		DisableOnCompiled:  req.DisableOnCompiled,
		SystemPrompt:       req.SystemPrompt,
		SearchSlots:        req.SearchSlots,
		ATSMode:            req.ATSMode,
		TimeFilter:         req.TimeFilter,
		Location:           req.Location,
		TargetType:         req.TargetType,
		TargetIDs:          req.TargetIDs,
		SourceDocumentIDs:  req.SourceDocumentIDs,
		DigestEnabled:      req.DigestEnabled,
		DigestEmails:       req.DigestEmails,
		DigestTime:         req.DigestTime,
		DigestModel:        req.DigestModel,
		UseAgent:           req.UseAgent,
		AgentModel:         req.AgentModel,
		EmptyProposalLimit: req.EmptyProposalLimit,
	}
}

func writeAutomationJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeAutomationError(w http.ResponseWriter, status int, message string) {
	writeAutomationJSON(w, status, map[string]string{"error": message})
}

func parseIntQueryParam(r *http.Request, key string, defaultVal int) int {
	val := r.URL.Query().Get(key)
	if val == "" {
		return defaultVal
	}
	parsed, err := strconv.Atoi(val)
	if err != nil || parsed < 1 {
		return defaultVal
	}
	return parsed
}
