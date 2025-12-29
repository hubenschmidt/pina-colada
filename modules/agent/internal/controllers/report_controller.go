package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"agent/internal/middleware"
	"agent/internal/schemas"
	"agent/internal/services"
)

// ReportController handles report HTTP requests
type ReportController struct {
	reportService *services.ReportService
}

// NewReportController creates a new report controller
func NewReportController(reportService *services.ReportService) *ReportController {
	return &ReportController{reportService: reportService}
}

// GetLeadPipeline handles GET /reports/canned/lead-pipeline
func (c *ReportController) GetLeadPipeline(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "tenant required")
		return
	}

	dateFrom := r.URL.Query().Get("date_from")
	dateTo := r.URL.Query().Get("date_to")
	projectID := parseOptionalInt64(r.URL.Query().Get("project_id"))

	var dateFromPtr, dateToPtr *string
	if dateFrom != "" {
		dateFromPtr = &dateFrom
	}
	if dateTo != "" {
		dateToPtr = &dateTo
	}

	result, err := c.reportService.GetLeadPipelineReport(tenantID, dateFromPtr, dateToPtr, projectID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// GetAccountOverview handles GET /reports/canned/account-overview
func (c *ReportController) GetAccountOverview(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "tenant required")
		return
	}

	result, err := c.reportService.GetAccountOverviewReport(tenantID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// GetContactCoverage handles GET /reports/canned/contact-coverage
func (c *ReportController) GetContactCoverage(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "tenant required")
		return
	}

	result, err := c.reportService.GetContactCoverageReport(tenantID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// GetNotesActivity handles GET /reports/canned/notes-activity
func (c *ReportController) GetNotesActivity(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "tenant required")
		return
	}

	projectID := parseOptionalInt64(r.URL.Query().Get("project_id"))

	result, err := c.reportService.GetNotesActivityReport(tenantID, projectID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// GetUserAudit handles GET /reports/canned/user-audit
func (c *ReportController) GetUserAudit(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "tenant required")
		return
	}

	userID := parseOptionalInt64(r.URL.Query().Get("user_id"))

	result, err := c.reportService.GetUserAuditReport(tenantID, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// GetEntityFields handles GET /reports/fields/{entity}
func (c *ReportController) GetEntityFields(w http.ResponseWriter, r *http.Request) {
	entity := chi.URLParam(r, "entity")
	if entity == "" {
		writeError(w, http.StatusBadRequest, "entity required")
		return
	}

	result, err := c.reportService.GetEntityFields(entity)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if result == nil {
		writeError(w, http.StatusNotFound, "unknown entity: "+entity)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// PreviewCustomReport handles POST /reports/custom/preview
func (c *ReportController) PreviewCustomReport(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "tenant required")
		return
	}

	var query schemas.ReportQueryRequest
	if err := json.NewDecoder(r.Body).Decode(&query); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := c.reportService.PreviewCustomReport(tenantID, query)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// RunCustomReport handles POST /reports/custom/run
func (c *ReportController) RunCustomReport(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "tenant required")
		return
	}

	var query schemas.ReportQueryRequest
	if err := json.NewDecoder(r.Body).Decode(&query); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := c.reportService.RunCustomReport(tenantID, query)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// ExportCustomReport handles POST /reports/custom/export
func (c *ReportController) ExportCustomReport(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "tenant required")
		return
	}

	var query schemas.ReportQueryRequest
	if err := json.NewDecoder(r.Body).Decode(&query); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := c.reportService.ExportCustomReport(tenantID, query)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Return JSON data (actual Excel generation would be done client-side or with additional library)
	writeJSON(w, http.StatusOK, result)
}

// ListSavedReports handles GET /reports/saved
func (c *ReportController) ListSavedReports(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "tenant required")
		return
	}

	projectID := parseOptionalInt64(r.URL.Query().Get("project_id"))
	includeGlobal := r.URL.Query().Get("include_global") != "false"
	search := r.URL.Query().Get("q")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	sortBy := r.URL.Query().Get("sort_by")
	order := r.URL.Query().Get("order")

	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}

	result, err := c.reportService.GetSavedReports(tenantID, projectID, includeGlobal, search, page, limit, sortBy, order)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// CreateSavedReport handles POST /reports/saved
func (c *ReportController) CreateSavedReport(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "tenant required")
		return
	}

	var userID *int64
	if uid, ok := middleware.GetUserID(r.Context()); ok {
		userID = &uid
	}

	var input schemas.SavedReportCreate
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := c.reportService.CreateSavedReport(tenantID, userID, input)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, result)
}

// GetSavedReport handles GET /reports/saved/{report_id}
func (c *ReportController) GetSavedReport(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "tenant required")
		return
	}

	reportID, err := strconv.ParseInt(chi.URLParam(r, "report_id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid report_id")
		return
	}

	result, err := c.reportService.GetSavedReport(reportID, tenantID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if result == nil {
		writeError(w, http.StatusNotFound, "report not found")
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// UpdateSavedReport handles PUT /reports/saved/{report_id}
func (c *ReportController) UpdateSavedReport(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "tenant required")
		return
	}

	reportID, err := strconv.ParseInt(chi.URLParam(r, "report_id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid report_id")
		return
	}

	var input schemas.SavedReportUpdate
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := c.reportService.UpdateSavedReport(reportID, tenantID, input)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if result == nil {
		writeError(w, http.StatusNotFound, "report not found")
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// DeleteSavedReport handles DELETE /reports/saved/{report_id}
func (c *ReportController) DeleteSavedReport(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "tenant required")
		return
	}

	reportID, err := strconv.ParseInt(chi.URLParam(r, "report_id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid report_id")
		return
	}

	deleted, err := c.reportService.DeleteSavedReport(reportID, tenantID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !deleted {
		writeError(w, http.StatusNotFound, "report not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Report deleted"})
}
