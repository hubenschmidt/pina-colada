package controllers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/pina-colada-co/agent-go/internal/agent/utils"
	apperrors "github.com/pina-colada-co/agent-go/internal/errors"
	"github.com/pina-colada-co/agent-go/internal/middleware"
	"github.com/pina-colada-co/agent-go/internal/serializers"
	"github.com/pina-colada-co/agent-go/internal/services"
)

type AgentConfigController struct {
	configService *services.AgentConfigService
	configCache   *utils.ConfigCache
}

func NewAgentConfigController(configService *services.AgentConfigService, configCache *utils.ConfigCache) *AgentConfigController {
	return &AgentConfigController{
		configService: configService,
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

	result, err := c.configService.GetAgentConfig(userID)
	if err != nil {
		writeConfigError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeConfigJSON(w, http.StatusOK, result)
}

// UpdateNodeConfigRequest represents the request body for updating a node config
type UpdateNodeConfigRequest struct {
	Model            string   `json:"model"`
	Temperature      *float64 `json:"temperature,omitempty"`
	MaxTokens        *int     `json:"max_tokens,omitempty"`
	TopP             *float64 `json:"top_p,omitempty"`
	TopK             *int     `json:"top_k,omitempty"`
	FrequencyPenalty *float64 `json:"frequency_penalty,omitempty"`
	PresencePenalty  *float64 `json:"presence_penalty,omitempty"`
}

// UpdateNodeConfig handles PUT /agent/config/{node_name}
func (c *AgentConfigController) UpdateNodeConfig(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeConfigError(w, http.StatusUnauthorized, "unauthorized")
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

	req := services.NodeConfigUpdateRequest{
		Model:            input.Model,
		Temperature:      input.Temperature,
		MaxTokens:        input.MaxTokens,
		TopP:             input.TopP,
		TopK:             input.TopK,
		FrequencyPenalty: input.FrequencyPenalty,
		PresencePenalty:  input.PresencePenalty,
	}
	result, err := c.configService.UpdateNodeConfig(userID, nodeName, req)
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
	_, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeConfigError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	result := c.configService.GetAvailableModels()
	writeConfigJSON(w, http.StatusOK, result)
}

// CreatePresetRequest represents the request body for creating a preset
type CreatePresetRequest struct {
	Name             string   `json:"name"`
	Temperature      *float64 `json:"temperature,omitempty"`
	MaxTokens        *int     `json:"max_tokens,omitempty"`
	TopP             *float64 `json:"top_p,omitempty"`
	TopK             *int     `json:"top_k,omitempty"`
	FrequencyPenalty *float64 `json:"frequency_penalty,omitempty"`
	PresencePenalty  *float64 `json:"presence_penalty,omitempty"`
}

// ListPresets handles GET /agent/config/presets
func (c *AgentConfigController) ListPresets(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeConfigError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	result, err := c.configService.ListPresets(userID)
	if err != nil {
		writeConfigError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeConfigJSON(w, http.StatusOK, result)
}

// CreatePreset handles POST /agent/config/presets
func (c *AgentConfigController) CreatePreset(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeConfigError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var input CreatePresetRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeConfigError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if input.Name == "" {
		writeConfigError(w, http.StatusBadRequest, "name is required")
		return
	}

	req := services.PresetCreateRequest{
		Name:             input.Name,
		Temperature:      input.Temperature,
		MaxTokens:        input.MaxTokens,
		TopP:             input.TopP,
		TopK:             input.TopK,
		FrequencyPenalty: input.FrequencyPenalty,
		PresencePenalty:  input.PresencePenalty,
	}
	result, err := c.configService.CreatePreset(userID, req)
	if err != nil {
		writeConfigError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeConfigJSON(w, http.StatusCreated, result)
}

// DeletePreset handles DELETE /agent/config/presets/{id}
func (c *AgentConfigController) DeletePreset(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeConfigError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	presetID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeConfigError(w, http.StatusBadRequest, "invalid preset id")
		return
	}

	err = c.configService.DeletePreset(userID, presetID)
	if errors.Is(err, apperrors.ErrNotFound) {
		writeConfigError(w, http.StatusNotFound, "preset not found")
		return
	}
	if err != nil {
		writeConfigError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ApplyPreset handles POST /agent/config/presets/{id}/apply
func (c *AgentConfigController) ApplyPreset(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeConfigError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	presetID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeConfigError(w, http.StatusBadRequest, "invalid preset id")
		return
	}

	result, err := c.configService.ApplyPreset(userID, presetID)
	if errors.Is(err, apperrors.ErrNotFound) {
		writeConfigError(w, http.StatusNotFound, "preset not found")
		return
	}
	if err != nil {
		writeConfigError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Invalidate cache so next chat uses new settings
	if c.configCache != nil {
		c.configCache.Invalidate(userID)
	}

	writeConfigJSON(w, http.StatusOK, result)
}

// GetCostTiers handles GET /agent/config/cost-tiers
func (c *AgentConfigController) GetCostTiers(w http.ResponseWriter, r *http.Request) {
	_, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeConfigError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	result := c.configService.GetCostTiers()
	writeConfigJSON(w, http.StatusOK, result)
}

// ApplyCostTierRequest represents the request body for applying a cost tier
type ApplyCostTierRequest struct {
	Tier     string `json:"tier"`
	Provider string `json:"provider,omitempty"`
}

// ApplyCostTier handles POST /agent/config/cost-tier
func (c *AgentConfigController) ApplyCostTier(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeConfigError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var input ApplyCostTierRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeConfigError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if input.Tier == "" {
		writeConfigError(w, http.StatusBadRequest, "tier is required")
		return
	}

	result, err := c.configService.ApplyCostTierWithProvider(userID, input.Tier, input.Provider)
	if errors.Is(err, services.ErrInvalidCostTier) {
		writeConfigError(w, http.StatusBadRequest, "invalid cost tier")
		return
	}
	if err != nil {
		writeConfigError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Invalidate cache so next chat uses new models
	if c.configCache != nil {
		c.configCache.Invalidate(userID)
	}

	writeConfigJSON(w, http.StatusOK, result)
}

func writeConfigJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeConfigError(w http.ResponseWriter, status int, message string) {
	writeConfigJSON(w, status, serializers.ErrorResponse{Error: message})
}
