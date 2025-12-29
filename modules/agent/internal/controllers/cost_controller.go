package controllers

import (
	"net/http"

	"agent/internal/services"
)

// CostController handles costs HTTP requests
type CostController struct {
	costsService *services.CostService
}

// NewCostController creates a new costs controller
func NewCostController(costsService *services.CostService) *CostController {
	return &CostController{costsService: costsService}
}

// GetCostsSummary handles GET /costs/summary
func (c *CostController) GetCostsSummary(w http.ResponseWriter, r *http.Request) {
	period := r.URL.Query().Get("period")
	if period == "" {
		period = "monthly"
	}

	result, err := c.costsService.GetCombinedCosts(period)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// GetOrgCosts handles GET /costs/org
func (c *CostController) GetOrgCosts(w http.ResponseWriter, r *http.Request) {
	period := r.URL.Query().Get("period")
	if period == "" {
		period = "monthly"
	}

	result, err := c.costsService.GetOrgCosts(period)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, result)
}
