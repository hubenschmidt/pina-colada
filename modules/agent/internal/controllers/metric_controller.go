package controllers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"agent/internal/middleware"
	"agent/internal/serializers"
	"agent/internal/services"
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

	params := parseSessionQueryParams(r)
	result, err := c.metricService.ListSessionsPaginated(userID, params)
	if err != nil {
		writeMetricError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeMetricJSON(w, http.StatusOK, result)
}

func parseSessionQueryParams(r *http.Request) *services.SessionQueryParams {
	query := r.URL.Query()

	page := 1
	if parsed, err := strconv.Atoi(query.Get("page")); err == nil && parsed > 0 {
		page = parsed
	}

	pageSize := 10
	if parsed, err := strconv.Atoi(query.Get("page_size")); err == nil && parsed > 0 && parsed <= 100 {
		pageSize = parsed
	}

	return &services.SessionQueryParams{
		Page:          page,
		PageSize:      pageSize,
		SortBy:        query.Get("sort_by"),
		SortDirection: query.Get("sort_direction"),
	}
}

// GetSession handles GET /metrics/sessions/{id}
// Query params: page, page_size, sort_by, sort_direction
func (c *MetricController) GetSession(w http.ResponseWriter, r *http.Request) {
	sessionID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeMetricError(w, http.StatusBadRequest, "invalid session id")
		return
	}

	// Parse pagination params
	params := parseMetricQueryParams(r)

	result, err := c.metricService.GetSessionWithMetrics(sessionID, params)
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

func parseMetricQueryParams(r *http.Request) *services.MetricQueryParams {
	query := r.URL.Query()

	pageStr := query.Get("page")
	pageSizeStr := query.Get("page_size")

	// Return nil if no pagination params (backwards compatible)
	if pageStr == "" && pageSizeStr == "" {
		return nil
	}

	page := 1
	if parsed, err := strconv.Atoi(pageStr); err == nil && parsed > 0 {
		page = parsed
	}

	pageSize := 10
	if parsed, err := strconv.Atoi(pageSizeStr); err == nil && parsed > 0 && parsed <= 100 {
		pageSize = parsed
	}

	return &services.MetricQueryParams{
		Page:          page,
		PageSize:      pageSize,
		SortBy:        query.Get("sort_by"),
		SortDirection: query.Get("sort_direction"),
	}
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
