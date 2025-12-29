package controllers

import (
	"encoding/json"
	"errors"
	"net/http"

	"agent/internal/middleware"
	"agent/internal/serializers"
	"agent/internal/services"
)

type PreferenceController struct {
	prefsService *services.PreferenceService
}

func NewPreferenceController(prefsService *services.PreferenceService) *PreferenceController {
	return &PreferenceController{prefsService: prefsService}
}

func (c *PreferenceController) GetUserPreferences(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writePrefsError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	result, err := c.prefsService.GetUserPreferences(userID)
	if errors.Is(err, services.ErrPreferencesUserNotFound) {
		writePrefsError(w, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		writePrefsError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writePrefsJSON(w, http.StatusOK, result)
}

type UpdatePreferencesRequest struct {
	Theme    *string `json:"theme"`
	Timezone *string `json:"timezone"`
}

func (c *PreferenceController) UpdateUserPreferences(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writePrefsError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var input UpdatePreferencesRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writePrefsError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := c.prefsService.UpdateUserPreferences(userID, input.Theme, input.Timezone)
	if err != nil {
		writePrefsError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writePrefsJSON(w, http.StatusOK, result)
}

func writePrefsJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writePrefsError(w http.ResponseWriter, status int, message string) {
	writePrefsJSON(w, status, serializers.ErrorResponse{Error: message})
}

// GetTimezones handles GET /preferences/timezones
func (c *PreferenceController) GetTimezones(w http.ResponseWriter, r *http.Request) {
	timezones := []map[string]string{
		{"value": "UTC", "label": "UTC"},
		{"value": "America/New_York", "label": "Eastern Time (ET)"},
		{"value": "America/Chicago", "label": "Central Time (CT)"},
		{"value": "America/Denver", "label": "Mountain Time (MT)"},
		{"value": "America/Los_Angeles", "label": "Pacific Time (PT)"},
		{"value": "America/Phoenix", "label": "Arizona (no DST)"},
		{"value": "America/Anchorage", "label": "Alaska (AKT)"},
		{"value": "Pacific/Honolulu", "label": "Hawaii (HST)"},
		{"value": "Europe/London", "label": "London (GMT/BST)"},
		{"value": "Europe/Paris", "label": "Central European (CET)"},
		{"value": "Europe/Berlin", "label": "Berlin (CET)"},
		{"value": "Asia/Tokyo", "label": "Japan (JST)"},
		{"value": "Asia/Shanghai", "label": "China (CST)"},
		{"value": "Asia/Singapore", "label": "Singapore (SGT)"},
		{"value": "Australia/Sydney", "label": "Sydney (AEST)"},
	}
	writePrefsJSON(w, http.StatusOK, timezones)
}

// GetTenantPreferences handles GET /preferences/tenant
func (c *PreferenceController) GetTenantPreferences(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		writePrefsError(w, http.StatusBadRequest, "tenant not set")
		return
	}

	result, err := c.prefsService.GetTenantPreferences(tenantID)
	if err != nil {
		writePrefsError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writePrefsJSON(w, http.StatusOK, result)
}

// UpdateTenantPreferencesRequest represents the request body for updating tenant preferences
type UpdateTenantPreferencesRequest struct {
	Theme string `json:"theme"`
}

// UpdateTenantPreferences handles PATCH /preferences/tenant
func (c *PreferenceController) UpdateTenantPreferences(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		writePrefsError(w, http.StatusBadRequest, "tenant not set")
		return
	}

	var input UpdateTenantPreferencesRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writePrefsError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if input.Theme == "" {
		writePrefsError(w, http.StatusBadRequest, "theme is required")
		return
	}

	result, err := c.prefsService.UpdateTenantPreferences(tenantID, input.Theme)
	if err != nil {
		writePrefsError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writePrefsJSON(w, http.StatusOK, result)
}
