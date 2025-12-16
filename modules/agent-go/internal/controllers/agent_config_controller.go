package controllers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/pina-colada-co/agent-go/internal/agent"
	"github.com/pina-colada-co/agent-go/internal/middleware"
	"github.com/pina-colada-co/agent-go/internal/repositories"
	"github.com/pina-colada-co/agent-go/internal/serializers"
	"github.com/pina-colada-co/agent-go/internal/services"
)

type AgentConfigController struct {
	configService *services.AgentConfigService
	userRepo      *repositories.UserRepository
	configCache   *agent.ConfigCache
}

func NewAgentConfigController(configService *services.AgentConfigService, userRepo *repositories.UserRepository, configCache *agent.ConfigCache) *AgentConfigController {
	return &AgentConfigController{
		configService: configService,
		userRepo:      userRepo,
		configCache:   configCache,
	}
}

// GetConfig handles GET /agent/config
func (c *AgentConfigController) GetConfig(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeConfigError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	if !c.hasDeveloperAccess(userID) {
		writeConfigError(w, http.StatusForbidden, "developer access required")
		return
	}

	result, err := c.configService.GetAgentConfig(userID)
	if err != nil {
		writeConfigError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeConfigJSON(w, http.StatusOK, result)
}

// UpdateNodeConfigRequest represents the request body for updating a node config
type UpdateNodeConfigRequest struct {
	Model string `json:"model"`
}

// UpdateNodeConfig handles PUT /agent/config/{node_name}
func (c *AgentConfigController) UpdateNodeConfig(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeConfigError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	if !c.hasDeveloperAccess(userID) {
		writeConfigError(w, http.StatusForbidden, "developer access required")
		return
	}

	nodeName := chi.URLParam(r, "node_name")
	if nodeName == "" {
		writeConfigError(w, http.StatusBadRequest, "node_name is required")
		return
	}

	var input UpdateNodeConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeConfigError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if input.Model == "" {
		writeConfigError(w, http.StatusBadRequest, "model is required")
		return
	}

	result, err := c.configService.UpdateNodeConfig(userID, nodeName, input.Model)
	if errors.Is(err, services.ErrInvalidNodeName) {
		writeConfigError(w, http.StatusBadRequest, "invalid node name")
		return
	}
	if errors.Is(err, services.ErrInvalidModel) {
		writeConfigError(w, http.StatusBadRequest, "invalid model")
		return
	}
	if err != nil {
		writeConfigError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Invalidate cache so next chat uses new model
	if c.configCache != nil {
		c.configCache.Invalidate(userID)
	}

	writeConfigJSON(w, http.StatusOK, result)
}

// ResetNodeConfig handles DELETE /agent/config/{node_name}
func (c *AgentConfigController) ResetNodeConfig(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeConfigError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	if !c.hasDeveloperAccess(userID) {
		writeConfigError(w, http.StatusForbidden, "developer access required")
		return
	}

	nodeName := chi.URLParam(r, "node_name")
	if nodeName == "" {
		writeConfigError(w, http.StatusBadRequest, "node_name is required")
		return
	}

	result, err := c.configService.ResetNodeConfig(userID, nodeName)
	if errors.Is(err, services.ErrInvalidNodeName) {
		writeConfigError(w, http.StatusBadRequest, "invalid node name")
		return
	}
	if err != nil {
		writeConfigError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Invalidate cache so next chat uses default model
	if c.configCache != nil {
		c.configCache.Invalidate(userID)
	}

	writeConfigJSON(w, http.StatusOK, result)
}

// GetAvailableModels handles GET /agent/config/models
func (c *AgentConfigController) GetAvailableModels(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeConfigError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	if !c.hasDeveloperAccess(userID) {
		writeConfigError(w, http.StatusForbidden, "developer access required")
		return
	}

	result := c.configService.GetAvailableModels()
	writeConfigJSON(w, http.StatusOK, result)
}

func (c *AgentConfigController) hasDeveloperAccess(userID int64) bool {
	hasRole, err := c.userRepo.HasRole(userID, "developer")
	if err != nil {
		return false
	}
	return hasRole
}

func writeConfigJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeConfigError(w http.ResponseWriter, status int, message string) {
	writeConfigJSON(w, status, serializers.ErrorResponse{Error: message})
}
