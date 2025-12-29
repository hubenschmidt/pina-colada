package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"agent/internal/middleware"
	"agent/internal/schemas"
	"agent/internal/services"
)

// NotificationController handles notification HTTP requests
type NotificationController struct {
	notifService *services.NotificationService
}

// NewNotificationController creates a new notification controller
func NewNotificationController(notifService *services.NotificationService) *NotificationController {
	return &NotificationController{notifService: notifService}
}

// GetNotifications handles GET /notifications
func (c *NotificationController) GetNotifications(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	tenantID, _ := middleware.GetTenantID(r.Context())

	limit := parseLimit(r.URL.Query().Get("limit"), 20, 100)

	resp, err := c.notifService.GetNotifications(userID, tenantID, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// GetNotificationCount handles GET /notifications/count
func (c *NotificationController) GetNotificationCount(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	tenantID, _ := middleware.GetTenantID(r.Context())

	count, err := c.notifService.GetUnreadCount(userID, tenantID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]int64{"unread_count": count})
}

// MarkAsRead handles POST /notifications/mark-read
func (c *NotificationController) MarkAsRead(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	tenantID, _ := middleware.GetTenantID(r.Context())

	var req schemas.MarkReadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if len(req.NotificationIDs) == 0 {
		writeError(w, http.StatusBadRequest, "notification_ids required")
		return
	}

	updated, err := c.notifService.MarkAsRead(req.NotificationIDs, userID, tenantID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "ok",
		"updated": updated,
	})
}

// MarkEntityAsRead handles POST /notifications/mark-entity-read
func (c *NotificationController) MarkEntityAsRead(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	tenantID, _ := middleware.GetTenantID(r.Context())

	var req schemas.MarkEntityReadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.EntityType == "" || req.EntityID == 0 {
		writeError(w, http.StatusBadRequest, "entity_type and entity_id required")
		return
	}

	updated, err := c.notifService.MarkEntityAsRead(userID, tenantID, req.EntityType, req.EntityID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "ok",
		"updated": updated,
	})
}

func parseLimit(limitStr string, defaultVal, maxVal int) int {
	if limitStr == "" {
		return defaultVal
	}
	parsed, err := strconv.Atoi(limitStr)
	if err != nil {
		return defaultVal
	}
	if parsed <= 0 || parsed > maxVal {
		return defaultVal
	}
	return parsed
}
