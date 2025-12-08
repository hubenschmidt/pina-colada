package controllers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/pina-colada-co/agent-go/internal/middleware"
	"github.com/pina-colada-co/agent-go/internal/serializers"
	"github.com/pina-colada-co/agent-go/internal/services"
)

type PreferencesController struct {
	prefsService *services.PreferencesService
}

func NewPreferencesController(prefsService *services.PreferencesService) *PreferencesController {
	return &PreferencesController{prefsService: prefsService}
}

func (c *PreferencesController) GetUserPreferences(w http.ResponseWriter, r *http.Request) {
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

func (c *PreferencesController) UpdateUserPreferences(w http.ResponseWriter, r *http.Request) {
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
