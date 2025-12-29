package controllers

import (
	"net/http"

	"agent/internal/middleware"
	"agent/internal/services"
)

// UsageController handles usage analytics HTTP requests
type UsageController struct {
	usageService *services.UsageService
}

// NewUsageController creates a new usage controller
func NewUsageController(usageService *services.UsageService) *UsageController {
	return &UsageController{usageService: usageService}
}

// GetUserUsage handles GET /usage/user
func (c *UsageController) GetUserUsage(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "user not found")
		return
	}

	period := r.URL.Query().Get("period")
	if period == "" {
		period = "monthly"
	}

	result, err := c.usageService.GetUserUsage(userID, period)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// GetTenantUsage handles GET /usage/tenant
func (c *UsageController) GetTenantUsage(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "tenant not found")
		return
	}

	period := r.URL.Query().Get("period")
	if period == "" {
		period = "monthly"
	}

	result, err := c.usageService.GetTenantUsage(tenantID, period)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// GetUsageTimeseries handles GET /usage/timeseries
func (c *UsageController) GetUsageTimeseries(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "tenant not found")
		return
	}

	period := r.URL.Query().Get("period")
	if period == "" {
		period = "monthly"
	}

	scope := r.URL.Query().Get("scope")
	var userID *int64
	uid, ok := middleware.GetUserID(r.Context())
	if (scope == "user" || scope == "") && ok {
		userID = &uid
	}

	result, err := c.usageService.GetUsageTimeseries(tenantID, period, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// GetDeveloperAnalytics handles GET /usage/analytics
func (c *UsageController) GetDeveloperAnalytics(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "user not found")
		return
	}

	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "tenant not found")
		return
	}

	period := r.URL.Query().Get("period")
	if period == "" {
		period = "monthly"
	}

	groupBy := r.URL.Query().Get("group_by")
	if groupBy == "" {
		groupBy = "node"
	}

	result, err := c.usageService.GetDeveloperAnalytics(userID, tenantID, period, groupBy)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// CheckDeveloperAccess handles GET /usage/developer-access
func (c *UsageController) CheckDeveloperAccess(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "user not found")
		return
	}

	result, err := c.usageService.CheckDeveloperAccess(userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, result)
}
