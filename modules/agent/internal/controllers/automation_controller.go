package controllers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"agent/internal/middleware"
	"agent/internal/repositories"
	"agent/internal/services"
)

// TestRunner interface for running automation tests
type TestRunner interface {
	ExecuteForConfig(configID int64) error
}

// AutomationController handles automation HTTP requests
type AutomationController struct {
	automationService *services.AutomationService
	testRunner        TestRunner
}

// NewAutomationController creates a new automation controller
func NewAutomationController(automationService *services.AutomationService, testRunner TestRunner) *AutomationController {
	return &AutomationController{
		automationService: automationService,
		testRunner:        testRunner,
	}
}

// CrawlerRequest represents the request body for creating/updating a crawler
type CrawlerRequest struct {
	Name               *string  `json:"name,omitempty"`
	EntityType         *string  `json:"entity_type,omitempty"`
	Enabled            *bool    `json:"enabled,omitempty"`
	IntervalMinutes    *int     `json:"interval_minutes,omitempty"`
	LeadsPerRun        *int     `json:"leads_per_run,omitempty"`
	ConcurrentSearches *int     `json:"concurrent_searches,omitempty"`
	CompilationTarget  *int     `json:"compilation_target,omitempty"`
	SystemPrompt       *string  `json:"system_prompt,omitempty"`
	SearchKeywords     []string `json:"search_keywords,omitempty"`
	SearchSlots        [][]int  `json:"search_slots,omitempty"`
	ATSMode            *bool    `json:"ats_mode,omitempty"`
	TimeFilter         *string  `json:"time_filter,omitempty"`
	TargetType         *string  `json:"target_type,omitempty"`
	TargetIDs          []int64  `json:"target_ids,omitempty"`
	SourceDocumentIDs  []int64  `json:"source_document_ids,omitempty"`
	DigestEnabled      *bool    `json:"digest_enabled,omitempty"`
	DigestEmails       *string  `json:"digest_emails,omitempty"`
	DigestTime         *string  `json:"digest_time,omitempty"`
	DigestModel        *string  `json:"digest_model,omitempty"`
	UseAgent           *bool    `json:"use_agent,omitempty"`
	AgentModel         *string  `json:"agent_model,omitempty"`
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

	limit := 10
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	result, err := c.automationService.GetCrawlerRuns(configID, limit)
	if err != nil {
		writeAutomationError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeAutomationJSON(w, http.StatusOK, map[string]interface{}{"runs": result})
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

func (c *AutomationController) parseConfigID(r *http.Request) (int64, error) {
	idStr := chi.URLParam(r, "id")
	return strconv.ParseInt(idStr, 10, 64)
}

func (c *AutomationController) requestToInput(req CrawlerRequest) repositories.AutomationConfigInput {
	return repositories.AutomationConfigInput{
		Name:               req.Name,
		EntityType:         req.EntityType,
		Enabled:            req.Enabled,
		IntervalMinutes:    req.IntervalMinutes,
		LeadsPerRun:        req.LeadsPerRun,
		ConcurrentSearches: req.ConcurrentSearches,
		CompilationTarget:  req.CompilationTarget,
		SystemPrompt:       req.SystemPrompt,
		SearchKeywords:     req.SearchKeywords,
		SearchSlots:        req.SearchSlots,
		ATSMode:            req.ATSMode,
		TimeFilter:         req.TimeFilter,
		TargetType:         req.TargetType,
		TargetIDs:          req.TargetIDs,
		SourceDocumentIDs:  req.SourceDocumentIDs,
		DigestEnabled:      req.DigestEnabled,
		DigestEmails:       req.DigestEmails,
		DigestTime:         req.DigestTime,
		DigestModel:        req.DigestModel,
		UseAgent:           req.UseAgent,
		AgentModel:         req.AgentModel,
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
