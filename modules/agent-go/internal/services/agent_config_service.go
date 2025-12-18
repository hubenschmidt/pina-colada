package services

import (
	"errors"

	"github.com/pina-colada-co/agent-go/internal/repositories"
)

var ErrInvalidNodeName = errors.New("invalid node name")
var ErrInvalidModel = errors.New("invalid model for provider")
var ErrInvalidCostTier = errors.New("invalid cost tier")

// LLMSettings contains configurable LLM parameters
type LLMSettings struct {
	Temperature      *float64
	MaxTokens        *int
	TopP             *float64
	TopK             *int     // Anthropic only
	FrequencyPenalty *float64 // OpenAI only
	PresencePenalty  *float64 // OpenAI only
}

// unsupportedModelParams maps models to parameters they don't support
// top_k is Anthropic-only, must be blocked for all OpenAI models
var unsupportedModelParams = map[string]map[string]bool{
	"claude-opus-4-5-20251101":   {"max_tokens": true, "top_p": true, "temperature": true},
	"claude-sonnet-4-5-20250929": {"top_p": true, "temperature": true, "max_tokens": true},
	"gpt-5":                      {"top_p": true, "temperature": true, "top_k": true},
	"gpt-5.1":                    {"top_p": true, "temperature": true, "top_k": true},
	"gpt-5.2":                    {"top_p": true, "temperature": true, "top_k": true},
	"gpt-5-mini":                 {"top_p": true, "temperature": true, "top_k": true},
	"gpt-5-nano":                 {"temperature": true, "top_p": true, "top_k": true},
	"gpt-4.1":                    {"top_k": true},
	"gpt-4.1-mini":               {"top_k": true},
	"gpt-4o":                     {"top_k": true},
	"gpt-4o-mini":                {"top_k": true},
	"o3":                         {"top_p": true, "temperature": true, "top_k": true},
	"o4-mini":                    {"temperature": true, "top_p": true, "top_k": true},
}

// FilterSettingsForModel removes unsupported parameters for the given model
func FilterSettingsForModel(settings LLMSettings, model string) LLMSettings {
	blocked, ok := unsupportedModelParams[model]
	if !ok {
		return settings
	}

	filtered := settings
	if blocked["temperature"] {
		filtered.Temperature = nil
	}
	if blocked["max_tokens"] {
		filtered.MaxTokens = nil
	}
	if blocked["top_p"] {
		filtered.TopP = nil
	}
	if blocked["top_k"] {
		filtered.TopK = nil
	}
	if blocked["frequency_penalty"] {
		filtered.FrequencyPenalty = nil
	}
	if blocked["presence_penalty"] {
		filtered.PresencePenalty = nil
	}
	return filtered
}

type AgentConfigService struct {
	configRepo    *repositories.AgentConfigRepository
	modelRepo     *repositories.AvailableModelRepository
	modelCache    map[string]string // modelName -> provider cache
	modelCacheErr error
}

func NewAgentConfigService(configRepo *repositories.AgentConfigRepository, modelRepo *repositories.AvailableModelRepository) *AgentConfigService {
	return &AgentConfigService{
		configRepo: configRepo,
		modelRepo:  modelRepo,
	}
}

// NodeConfigResponse represents a single node's configuration
type NodeConfigResponse struct {
	NodeName         string   `json:"node_name"`
	DisplayName      string   `json:"display_name"`
	Description      string   `json:"description"`
	Model            string   `json:"model"`
	Provider         string   `json:"provider"`
	Temperature      *float64 `json:"temperature,omitempty"`
	MaxTokens        *int     `json:"max_tokens,omitempty"`
	TopP             *float64 `json:"top_p,omitempty"`
	TopK             *int     `json:"top_k,omitempty"`
	FrequencyPenalty *float64 `json:"frequency_penalty,omitempty"`
	PresencePenalty  *float64 `json:"presence_penalty,omitempty"`
	FallbackChain    []byte   `json:"fallback_chain,omitempty"`
	IsDefault        bool     `json:"is_default"`
}

// AgentConfigResponse contains all node configurations
type AgentConfigResponse struct {
	Nodes                 []NodeConfigResponse `json:"nodes"`
	SelectedParamPresetID *int64               `json:"selected_param_preset_id,omitempty"`
	SelectedCostTier      string               `json:"selected_cost_tier"`
}

// AvailableModelsResponse contains models grouped by provider
type AvailableModelsResponse struct {
	OpenAI    []string `json:"openai"`
	Anthropic []string `json:"anthropic"`
}

// GetAgentConfig returns all node configurations for a user, merged with defaults
func (s *AgentConfigService) GetAgentConfig(userID int64) (*AgentConfigResponse, error) {
	userConfigs, err := s.configRepo.GetUserConfigs(userID)
	if err != nil {
		return nil, err
	}

	selection, err := s.configRepo.GetUserSelection(userID)
	if err != nil {
		return nil, err
	}

	// Build a map of user overrides
	overrides := make(map[string]repositories.AgentNodeConfigDTO)
	for _, cfg := range userConfigs {
		overrides[cfg.NodeName] = cfg
	}

	// Build response with all nodes
	nodes := make([]NodeConfigResponse, 0, len(repositories.AllNodeNames))
	for _, nodeName := range repositories.AllNodeNames {
		defaultCfg := repositories.DefaultModels[nodeName]
		isDefault := true
		model := defaultCfg.Model
		provider := defaultCfg.Provider

		var temperature *float64
		var maxTokens *int
		var topP *float64
		var topK *int
		var frequencyPenalty *float64
		var presencePenalty *float64
		var fallbackChain []byte

		if override, ok := overrides[nodeName]; ok {
			model = override.Model
			provider = override.Provider
			temperature = override.Temperature
			maxTokens = override.MaxTokens
			topP = override.TopP
			topK = override.TopK
			frequencyPenalty = override.FrequencyPenalty
			presencePenalty = override.PresencePenalty
			fallbackChain = override.FallbackChain
			isDefault = false
		}

		nodes = append(nodes, NodeConfigResponse{
			NodeName:         nodeName,
			DisplayName:      repositories.NodeDisplayNames[nodeName],
			Description:      repositories.NodeDescriptions[nodeName],
			Model:            model,
			Provider:         provider,
			Temperature:      temperature,
			MaxTokens:        maxTokens,
			TopP:             topP,
			TopK:             topK,
			FrequencyPenalty: frequencyPenalty,
			PresencePenalty:  presencePenalty,
			FallbackChain:    fallbackChain,
			IsDefault:        isDefault,
		})
	}

	return &AgentConfigResponse{
		Nodes:                 nodes,
		SelectedParamPresetID: selection.SelectedParamPresetID,
		SelectedCostTier:      selection.SelectedCostTier,
	}, nil
}

// NodeConfigUpdateRequest contains fields for updating a node configuration
type NodeConfigUpdateRequest struct {
	Model            string
	Temperature      *float64
	MaxTokens        *int
	TopP             *float64
	TopK             *int
	FrequencyPenalty *float64
	PresencePenalty  *float64
}

// UpdateNodeConfig updates a single node's model configuration
func (s *AgentConfigService) UpdateNodeConfig(userID int64, nodeName string, req NodeConfigUpdateRequest) (*NodeConfigResponse, error) {
	// Validate node name
	if !repositories.IsValidNodeName(nodeName) {
		return nil, ErrInvalidNodeName
	}

	// Determine provider from model (from DB)
	provider := s.GetProviderForModel(req.Model)
	if provider == "" {
		return nil, ErrInvalidModel
	}

	// Upsert the config
	update := repositories.NodeConfigUpdate{
		Model:            req.Model,
		Provider:         provider,
		Temperature:      req.Temperature,
		MaxTokens:        req.MaxTokens,
		TopP:             req.TopP,
		TopK:             req.TopK,
		FrequencyPenalty: req.FrequencyPenalty,
		PresencePenalty:  req.PresencePenalty,
	}
	err := s.configRepo.UpsertNodeConfig(userID, nodeName, update)
	if err != nil {
		return nil, err
	}

	return &NodeConfigResponse{
		NodeName:         nodeName,
		DisplayName:      repositories.NodeDisplayNames[nodeName],
		Description:      repositories.NodeDescriptions[nodeName],
		Model:            req.Model,
		Provider:         provider,
		Temperature:      req.Temperature,
		MaxTokens:        req.MaxTokens,
		TopP:             req.TopP,
		TopK:             req.TopK,
		FrequencyPenalty: req.FrequencyPenalty,
		PresencePenalty:  req.PresencePenalty,
		IsDefault:        false,
	}, nil
}

// ResetNodeConfig removes user override and reverts to default
func (s *AgentConfigService) ResetNodeConfig(userID int64, nodeName string) (*NodeConfigResponse, error) {
	// Validate node name
	if !repositories.IsValidNodeName(nodeName) {
		return nil, ErrInvalidNodeName
	}

	// Delete the override (ignore not found)
	_ = s.configRepo.DeleteNodeConfig(userID, nodeName)

	// Return default config
	defaultCfg := repositories.DefaultModels[nodeName]
	return &NodeConfigResponse{
		NodeName:    nodeName,
		DisplayName: repositories.NodeDisplayNames[nodeName],
		Description: repositories.NodeDescriptions[nodeName],
		Model:       defaultCfg.Model,
		Provider:    defaultCfg.Provider,
		IsDefault:   true,
	}, nil
}

// GetAvailableModels returns all available models grouped by provider
func (s *AgentConfigService) GetAvailableModels() *AvailableModelsResponse {
	resp := &AvailableModelsResponse{
		OpenAI:    []string{},
		Anthropic: []string{},
	}

	if s.modelRepo == nil {
		return resp
	}

	models, err := s.modelRepo.GetActiveModels()
	if err != nil {
		return resp
	}

	for _, m := range models {
		if m.Provider == "openai" {
			resp.OpenAI = append(resp.OpenAI, m.ModelName)
		}
		if m.Provider == "anthropic" {
			resp.Anthropic = append(resp.Anthropic, m.ModelName)
		}
	}

	return resp
}

// GetProviderForModel returns the provider for a model from DB, or empty string if not found
func (s *AgentConfigService) GetProviderForModel(modelName string) string {
	if s.modelRepo == nil {
		return ""
	}

	model, err := s.modelRepo.GetModelByName(modelName)
	if err != nil || model == nil {
		return ""
	}

	return model.Provider
}

// ModelTierDTO represents a model tier for fallback chains
type ModelTierDTO struct {
	Model          string
	TimeoutSeconds int
}

// GetDefaultFallbackChain returns the default fallback chain for a provider from DB
// Models are ordered by sort_order (ascending), so fastest models come last
func (s *AgentConfigService) GetDefaultFallbackChain(provider string) []ModelTierDTO {
	if s.modelRepo == nil {
		return nil
	}

	models, err := s.modelRepo.GetModelsByProvider(provider)
	if err != nil || len(models) == 0 {
		return nil
	}

	// Build chain with models ordered by sort_order (slowest/most capable first)
	chain := make([]ModelTierDTO, len(models))
	for i, m := range models {
		chain[i] = ModelTierDTO{
			Model:          m.ModelName,
			TimeoutSeconds: m.DefaultTimeoutSeconds,
		}
	}

	return chain
}

// GetModelForNode returns the model for a specific node for a user
func (s *AgentConfigService) GetModelForNode(userID int64, nodeName string) (model, provider string) {
	cfg, err := s.configRepo.GetNodeConfig(userID, nodeName)
	if err != nil || cfg == nil {
		// Return default
		defaultCfg := repositories.DefaultModels[nodeName]
		return defaultCfg.Model, defaultCfg.Provider
	}
	return cfg.Model, cfg.Provider
}

// PresetResponse represents a preset for API responses
type PresetResponse struct {
	ID               int64    `json:"id"`
	Name             string   `json:"name"`
	Temperature      *float64 `json:"temperature,omitempty"`
	MaxTokens        *int     `json:"max_tokens,omitempty"`
	TopP             *float64 `json:"top_p,omitempty"`
	TopK             *int     `json:"top_k,omitempty"`
	FrequencyPenalty *float64 `json:"frequency_penalty,omitempty"`
	PresencePenalty  *float64 `json:"presence_penalty,omitempty"`
	IsGlobal         bool     `json:"is_global"`
}

// PresetCreateRequest contains fields for creating a preset
type PresetCreateRequest struct {
	Name             string
	Temperature      *float64
	MaxTokens        *int
	TopP             *float64
	TopK             *int
	FrequencyPenalty *float64
	PresencePenalty  *float64
}

// ListPresets returns all presets for a user
func (s *AgentConfigService) ListPresets(userID int64) ([]PresetResponse, error) {
	presets, err := s.configRepo.GetUserPresets(userID)
	if err != nil {
		return nil, err
	}

	result := make([]PresetResponse, len(presets))
	for i, p := range presets {
		result[i] = PresetResponse{
			ID:               p.ID,
			Name:             p.Name,
			Temperature:      p.Temperature,
			MaxTokens:        p.MaxTokens,
			TopP:             p.TopP,
			TopK:             p.TopK,
			FrequencyPenalty: p.FrequencyPenalty,
			PresencePenalty:  p.PresencePenalty,
			IsGlobal:         p.UserID == nil,
		}
	}
	return result, nil
}

// CreatePreset creates a new preset
func (s *AgentConfigService) CreatePreset(userID int64, req PresetCreateRequest) (*PresetResponse, error) {
	input := repositories.PresetCreateInput{
		UserID:           userID,
		Name:             req.Name,
		Temperature:      req.Temperature,
		MaxTokens:        req.MaxTokens,
		TopP:             req.TopP,
		TopK:             req.TopK,
		FrequencyPenalty: req.FrequencyPenalty,
		PresencePenalty:  req.PresencePenalty,
	}
	presetID, err := s.configRepo.CreatePreset(input)
	if err != nil {
		return nil, err
	}

	return &PresetResponse{
		ID:               presetID,
		Name:             req.Name,
		Temperature:      req.Temperature,
		MaxTokens:        req.MaxTokens,
		TopP:             req.TopP,
		TopK:             req.TopK,
		FrequencyPenalty: req.FrequencyPenalty,
		PresencePenalty:  req.PresencePenalty,
		IsGlobal:         false,
	}, nil
}

// DeletePreset deletes a preset
func (s *AgentConfigService) DeletePreset(userID, presetID int64) error {
	return s.configRepo.DeletePreset(userID, presetID)
}

// ApplyPreset applies a preset to all nodes and returns updated config
func (s *AgentConfigService) ApplyPreset(userID, presetID int64) (*AgentConfigResponse, error) {
	preset, err := s.configRepo.GetPresetByID(userID, presetID)
	if err != nil {
		return nil, err
	}

	err = s.configRepo.ApplyPresetToNodes(userID, *preset)
	if err != nil {
		return nil, err
	}

	err = s.configRepo.SetSelectedParamPreset(userID, &presetID)
	if err != nil {
		return nil, err
	}

	return s.GetAgentConfig(userID)
}

// CostTierResponse represents a cost tier for API responses
type CostTierResponse struct {
	Tier        string `json:"tier"`
	Description string `json:"description"`
	OpenAI      string `json:"openai_model"`
	Anthropic   string `json:"anthropic_model"`
}

// GetCostTiers returns all available cost tiers
func (s *AgentConfigService) GetCostTiers() []CostTierResponse {
	result := make([]CostTierResponse, len(repositories.AllCostTiers))
	for i, tier := range repositories.AllCostTiers {
		models := repositories.CostTierModels[tier]
		result[i] = CostTierResponse{
			Tier:        tier,
			Description: repositories.CostTierDescriptions[tier],
			OpenAI:      models.OpenAI,
			Anthropic:   models.Anthropic,
		}
	}
	return result
}

// ApplyCostTier applies a cost tier to all nodes and returns updated config
func (s *AgentConfigService) ApplyCostTier(userID int64, tier string) (*AgentConfigResponse, error) {
	if !repositories.IsValidCostTier(tier) {
		return nil, ErrInvalidCostTier
	}

	err := s.configRepo.ApplyCostTierToNodes(userID, tier)
	if err != nil {
		return nil, err
	}

	err = s.configRepo.SetSelectedCostTier(userID, tier)
	if err != nil {
		return nil, err
	}

	return s.GetAgentConfig(userID)
}
