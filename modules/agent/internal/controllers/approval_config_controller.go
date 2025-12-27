package controllers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/pina-colada-co/agent-go/internal/middleware"
	"github.com/pina-colada-co/agent-go/internal/serializers"
	"github.com/pina-colada-co/agent-go/internal/services"
)

// ApprovalConfigController handles approval config HTTP requests
type ApprovalConfigController struct {
	configService *services.ApprovalConfigService
}

// NewApprovalConfigController creates a new approval config controller
func NewApprovalConfigController(configService *services.ApprovalConfigService) *ApprovalConfigController {
	return &ApprovalConfigController{configService: configService}
}

// GetConfig handles GET /admin/approval-config
func (c *ApprovalConfigController) GetConfig(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		writeApprovalConfigError(w, http.StatusUnauthorized, "tenant not found")
		return
	}

	result, err := c.configService.GetConfig(tenantID)
	if err != nil {
		writeApprovalConfigError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeApprovalConfigJSON(w, http.StatusOK, result)
}

type updateConfigRequest struct {
	RequiresApproval bool `json:"requires_approval"`
}

// UpdateConfig handles PUT /admin/approval-config/{entityType}
func (c *ApprovalConfigController) UpdateConfig(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		writeApprovalConfigError(w, http.StatusUnauthorized, "tenant not found")
		return
	}

	entityType := chi.URLParam(r, "entityType")

	var req updateConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeApprovalConfigError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := c.configService.UpdateConfig(tenantID, entityType, req.RequiresApproval)
	if errors.Is(err, services.ErrInvalidEntityTypeConfig) {
		writeApprovalConfigError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err != nil {
		writeApprovalConfigError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeApprovalConfigJSON(w, http.StatusOK, result)
}

func writeApprovalConfigJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeApprovalConfigError(w http.ResponseWriter, status int, message string) {
	writeApprovalConfigJSON(w, status, serializers.ErrorResponse{Error: message})
}
