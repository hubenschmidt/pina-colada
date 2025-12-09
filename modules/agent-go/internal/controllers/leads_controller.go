package controllers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/pina-colada-co/agent-go/internal/middleware"
	"github.com/pina-colada-co/agent-go/internal/services"
)

// LeadsController handles leads HTTP requests (job leads)
type LeadsController struct {
	jobService *services.JobService
}

// NewLeadsController creates a new leads controller
func NewLeadsController(jobService *services.JobService) *LeadsController {
	return &LeadsController{jobService: jobService}
}

// GetLeads handles GET /leads
func (c *LeadsController) GetLeads(w http.ResponseWriter, r *http.Request) {
	var tenantID *int64
	if tid, ok := middleware.GetTenantID(r.Context()); ok {
		tenantID = &tid
	}

	var statusNames []string
	if statuses := r.URL.Query().Get("statuses"); statuses != "" {
		for _, s := range strings.Split(statuses, ",") {
			trimmed := strings.TrimSpace(s)
			if trimmed != "" {
				statusNames = append(statusNames, trimmed)
			}
		}
	}

	leads, err := c.jobService.GetLeads(statusNames, tenantID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, leads)
}

// GetStatuses handles GET /leads/statuses
func (c *LeadsController) GetStatuses(w http.ResponseWriter, r *http.Request) {
	statuses, err := c.jobService.GetLeadStatuses()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, statuses)
}

// MarkAsApplied handles POST /leads/{job_id}/apply
func (c *LeadsController) MarkAsApplied(w http.ResponseWriter, r *http.Request) {
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
func (c *LeadsController) MarkAsDoNotApply(w http.ResponseWriter, r *http.Request) {
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

// parseIDParam is defined in job_controller.go but we need it here too
func init() {
	// ensure chi is imported for URLParam
	_ = chi.URLParam
}
