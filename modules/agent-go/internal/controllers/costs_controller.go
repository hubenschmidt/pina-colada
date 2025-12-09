package controllers

import (
	"net/http"

	"github.com/pina-colada-co/agent-go/internal/services"
)

// CostsController handles costs HTTP requests
type CostsController struct {
	costsService *services.CostsService
}

// NewCostsController creates a new costs controller
func NewCostsController(costsService *services.CostsService) *CostsController {
	return &CostsController{costsService: costsService}
}

// GetCostsSummary handles GET /costs/summary
func (c *CostsController) GetCostsSummary(w http.ResponseWriter, r *http.Request) {
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
func (c *CostsController) GetOrgCosts(w http.ResponseWriter, r *http.Request) {
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
