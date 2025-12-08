package controllers

import (
	"net/http"
	"strconv"

	"github.com/pina-colada-co/agent-go/internal/middleware"
	"github.com/pina-colada-co/agent-go/internal/services"
)

// ConversationController handles conversation HTTP requests
type ConversationController struct {
	convService *services.ConversationService
}

// NewConversationController creates a new conversation controller
func NewConversationController(convService *services.ConversationService) *ConversationController {
	return &ConversationController{convService: convService}
}

// GetConversations handles GET /conversations
func (c *ConversationController) GetConversations(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var tenantID *int64
	if tid, ok := middleware.GetTenantID(r.Context()); ok {
		tenantID = &tid
	}

	includeArchived := r.URL.Query().Get("include_archived") == "true"
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 50
	}

	results, err := c.convService.GetConversations(userID, tenantID, includeArchived, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, results)
}
