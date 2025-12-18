package controllers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/pina-colada-co/agent-go/internal/middleware"
	"github.com/pina-colada-co/agent-go/internal/serializers"
	"github.com/pina-colada-co/agent-go/internal/services"
)

// MetricController handles HTTP requests for metrics
type MetricController struct {
	metricService *services.MetricService
}

// NewMetricController creates a new metric controller
func NewMetricController(metricService *services.MetricService) *MetricController {
	return &MetricController{metricService: metricService}
}

// StartRecording handles POST /metrics/recording/start
func (c *MetricController) StartRecording(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeMetricError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		writeMetricError(w, http.StatusBadRequest, "tenant not found")
		return
	}

	result, err := c.metricService.StartRecording(userID, tenantID)
	if errors.Is(err, services.ErrSessionAlreadyActive) {
		writeMetricError(w, http.StatusConflict, "recording session already active")
		return
	}
	if err != nil {
		writeMetricError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeMetricJSON(w, http.StatusCreated, result)
}

// StopRecording handles POST /metrics/recording/stop
func (c *MetricController) StopRecording(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeMetricError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	result, err := c.metricService.StopRecording(userID)
	if errors.Is(err, services.ErrNoActiveSession) {
		writeMetricError(w, http.StatusNotFound, "no active recording session")
		return
	}
	if err != nil {
		writeMetricError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeMetricJSON(w, http.StatusOK, result)
}

// GetActiveSession handles GET /metrics/recording/active
func (c *MetricController) GetActiveSession(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeMetricError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	result, err := c.metricService.GetActiveSession(userID)
	if err != nil {
		writeMetricError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if result == nil {
		writeMetricJSON(w, http.StatusOK, map[string]interface{}{"active": false})
		return
	}

	writeMetricJSON(w, http.StatusOK, map[string]interface{}{
		"active":  true,
		"session": result,
	})
}

// ListSessions handles GET /metrics/sessions
func (c *MetricController) ListSessions(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeMetricError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	limit := 50
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	result, err := c.metricService.ListSessions(userID, limit)
	if err != nil {
		writeMetricError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeMetricJSON(w, http.StatusOK, result)
}

// GetSession handles GET /metrics/sessions/{id}
func (c *MetricController) GetSession(w http.ResponseWriter, r *http.Request) {
	sessionID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeMetricError(w, http.StatusBadRequest, "invalid session id")
		return
	}

	result, err := c.metricService.GetSessionWithMetrics(sessionID)
	if err != nil {
		writeMetricError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if result == nil {
		writeMetricError(w, http.StatusNotFound, "session not found")
		return
	}

	writeMetricJSON(w, http.StatusOK, result)
}

// Compare handles GET /metrics/compare?ids=1,2,3
func (c *MetricController) Compare(w http.ResponseWriter, r *http.Request) {
	idsStr := r.URL.Query().Get("ids")
	if idsStr == "" {
		writeMetricError(w, http.StatusBadRequest, "ids parameter required")
		return
	}

	sessionIDs, err := parseSessionIDs(idsStr)
	if err != nil {
		writeMetricError(w, http.StatusBadRequest, "invalid session ids")
		return
	}

	if len(sessionIDs) < 2 {
		writeMetricError(w, http.StatusBadRequest, "at least 2 session ids required for comparison")
		return
	}

	result, err := c.metricService.CompareSessions(sessionIDs)
	if err != nil {
		writeMetricError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeMetricJSON(w, http.StatusOK, result)
}

func parseSessionIDs(idsStr string) ([]int64, error) {
	parts := strings.Split(idsStr, ",")
	result := make([]int64, 0, len(parts))

	for _, part := range parts {
		id, err := strconv.ParseInt(strings.TrimSpace(part), 10, 64)
		if err != nil {
			return nil, err
		}
		result = append(result, id)
	}

	return result, nil
}

func writeMetricJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeMetricError(w http.ResponseWriter, status int, message string) {
	writeMetricJSON(w, status, serializers.ErrorResponse{Error: message})
}
