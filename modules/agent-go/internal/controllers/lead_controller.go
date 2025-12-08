package controllers

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/pina-colada-co/agent-go/internal/middleware"
	"github.com/pina-colada-co/agent-go/internal/services"
)

// LeadController handles lead-related HTTP requests (opportunities, partnerships)
type LeadController struct {
	leadService *services.LeadService
}

// NewLeadController creates a new lead controller
func NewLeadController(leadService *services.LeadService) *LeadController {
	return &LeadController{leadService: leadService}
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

	result, err := c.leadService.GetOpportunities(
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
