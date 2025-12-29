package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"agent/internal/middleware"
	"agent/internal/services"
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

// GetTenantConversations handles GET /conversations/all
func (c *ConversationController) GetTenantConversations(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "tenant required")
		return
	}

	search := r.URL.Query().Get("search")
	includeArchived := r.URL.Query().Get("include_archived") == "true"

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 100
	}

	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	results, err := c.convService.GetTenantConversations(tenantID, search, includeArchived, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, results)
}

// GetConversation handles GET /conversations/{thread_id}
func (c *ConversationController) GetConversation(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	threadID := chi.URLParam(r, "thread_id")
	if threadID == "" {
		writeError(w, http.StatusBadRequest, "thread_id required")
		return
	}

	result, err := c.convService.GetConversation(threadID, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if result == nil {
		writeError(w, http.StatusNotFound, "conversation not found")
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// UpdateConversationTitle handles PATCH /conversations/{thread_id}
func (c *ConversationController) UpdateConversationTitle(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	threadID := chi.URLParam(r, "thread_id")
	if threadID == "" {
		writeError(w, http.StatusBadRequest, "thread_id required")
		return
	}

	var input struct {
		Title string `json:"title"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := c.convService.UpdateTitle(threadID, userID, input.Title)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if result == nil {
		writeError(w, http.StatusNotFound, "conversation not found")
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// ArchiveConversation handles DELETE /conversations/{thread_id}
func (c *ConversationController) ArchiveConversation(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	threadID := chi.URLParam(r, "thread_id")
	if threadID == "" {
		writeError(w, http.StatusBadRequest, "thread_id required")
		return
	}

	if err := c.convService.ArchiveConversation(threadID, userID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}

// UnarchiveConversation handles POST /conversations/{thread_id}/unarchive
func (c *ConversationController) UnarchiveConversation(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	threadID := chi.URLParam(r, "thread_id")
	if threadID == "" {
		writeError(w, http.StatusBadRequest, "thread_id required")
		return
	}

	if err := c.convService.UnarchiveConversation(threadID, userID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}

// DeleteConversationPermanent handles DELETE /conversations/{thread_id}/permanent
func (c *ConversationController) DeleteConversationPermanent(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	threadID := chi.URLParam(r, "thread_id")
	if threadID == "" {
		writeError(w, http.StatusBadRequest, "thread_id required")
		return
	}

	if err := c.convService.DeleteConversation(threadID, userID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}
