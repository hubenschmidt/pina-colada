package controllers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/pina-colada-co/agent-go/internal/middleware"
	"github.com/pina-colada-co/agent-go/internal/services"
)

// LeadController handles lead-related HTTP requests (opportunities, partnerships, job leads)
type LeadController struct {
	leadService *services.LeadService
	jobService  *services.JobService
}

// NewLeadController creates a new lead controller
func NewLeadController(leadService *services.LeadService, jobService *services.JobService) *LeadController {
	return &LeadController{leadService: leadService, jobService: jobService}
}

// GetOpportunities returns paginated opportunities
func (c *LeadController) GetOpportunities(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	search := r.URL.Query().Get("search")

	var tenantID *int64
	if tid, ok := middleware.GetTenantID(r.Context()); ok {
		tenantID = &tid
	}

	projectID := parseOptionalInt64(r.URL.Query().Get("projectId"))

	result, err := c.leadService.GetOpportunities(
		page, limit,
		r.URL.Query().Get("orderBy"),
		r.URL.Query().Get("order"),
		search, tenantID, projectID,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// GetOpportunity returns a single opportunity by ID
func (c *LeadController) GetOpportunity(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid opportunity id")
		return
	}

	result, err := c.leadService.GetOpportunity(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if result == nil {
		writeError(w, http.StatusNotFound, "opportunity not found")
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// GetPartnerships returns paginated partnerships
func (c *LeadController) GetPartnerships(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	search := r.URL.Query().Get("search")

	var tenantID *int64
	if tid, ok := middleware.GetTenantID(r.Context()); ok {
		tenantID = &tid
	}

	result, err := c.leadService.GetPartnerships(
		page, limit,
		r.URL.Query().Get("orderBy"),
		r.URL.Query().Get("order"),
		search, tenantID,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// GetPartnership returns a single partnership by ID
func (c *LeadController) GetPartnership(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid partnership id")
		return
	}

	result, err := c.leadService.GetPartnership(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if result == nil {
		writeError(w, http.StatusNotFound, "partnership not found")
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// GetLeads handles GET /leads
func (c *LeadController) GetLeads(w http.ResponseWriter, r *http.Request) {
	var tenantID *int64
	if tid, ok := middleware.GetTenantID(r.Context()); ok {
		tenantID = &tid
	}

	statusNames := parseStatusNames(r.URL.Query().Get("statuses"))

	leads, err := c.jobService.GetLeads(statusNames, tenantID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, leads)
}

// GetStatuses handles GET /leads/statuses
func (c *LeadController) GetStatuses(w http.ResponseWriter, r *http.Request) {
	statuses, err := c.jobService.GetLeadStatuses()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, statuses)
}

// MarkAsApplied handles POST /leads/{job_id}/apply
func (c *LeadController) MarkAsApplied(w http.ResponseWriter, r *http.Request) {
	jobID, err := parseIDParam(r, "job_id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid job_id")
		return
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "user not authenticated")
		return
	}

	job, err := c.jobService.MarkLeadAsApplied(jobID, userID)
	if errors.Is(err, services.ErrJobNotFound) {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, job)
}

// MarkAsDoNotApply handles POST /leads/{job_id}/do-not-apply
func (c *LeadController) MarkAsDoNotApply(w http.ResponseWriter, r *http.Request) {
	jobID, err := parseIDParam(r, "job_id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid job_id")
		return
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "user not authenticated")
		return
	}

	job, err := c.jobService.MarkLeadAsDoNotApply(jobID, userID)
	if errors.Is(err, services.ErrJobNotFound) {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, job)
}

func parseStatusNames(statuses string) []string {
	if statuses == "" {
		return nil
	}
	var result []string
	for _, s := range strings.Split(statuses, ",") {
		trimmed := strings.TrimSpace(s)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}