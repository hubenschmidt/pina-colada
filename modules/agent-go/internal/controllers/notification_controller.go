package controllers

import (
	"net/http"
)

// NotificationController handles notification HTTP requests
type NotificationController struct{}

// NewNotificationController creates a new notification controller
func NewNotificationController() *NotificationController {
	return &NotificationController{}
}

// GetNotificationCount handles GET /notifications/count
func (c *NotificationController) GetNotificationCount(w http.ResponseWriter, r *http.Request) {
	// Stub - returns 0 unread notifications
	writeJSON(w, http.StatusOK, map[string]int{"count": 0})
}
