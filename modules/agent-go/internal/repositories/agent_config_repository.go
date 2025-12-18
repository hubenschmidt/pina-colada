package repositories

import (
	"errors"
	"slices"

	"gorm.io/gorm"

	apperrors "github.com/pina-colada-co/agent-go/internal/errors"
	"github.com/pina-colada-co/agent-go/internal/models"
)

// Node name constants
const (
	NodeTriageOrchestrator = "triage_orchestrator"
	NodeJobSearchWorker    = "job_search_worker"
	NodeCRMWorker          = "crm_worker"
	NodeGeneralWorker      = "general_worker"
	NodeEvaluator          = "evaluator"
	NodeTitleGenerator     = "title_generator"
)

// AllNodeNames returns all configurable node names
var AllNodeNames = []string{
	NodeTriageOrchestrator,
	NodeJobSearchWorker,
	NodeCRMWorker,
	NodeGeneralWorker,
	NodeEvaluator,
	NodeTitleGenerator,
}

// DefaultModelConfig holds default model and provider for a node
type DefaultModelConfig struct {
	Model    string
	Provider string
}

// DefaultModels defines the default model for each node
var DefaultModels = map[string]DefaultModelConfig{
	NodeTriageOrchestrator: {"gpt-5.2", "openai"},
	NodeJobSearchWorker:    {"gpt-5.2", "openai"},
	NodeCRMWorker:          {"gpt-5.2", "openai"},
	NodeGeneralWorker:      {"gpt-5.2", "openai"},
	NodeEvaluator:          {"claude-sonnet-4-5-20250929", "anthropic"},
	NodeTitleGenerator:     {"claude-haiku-4-5-20251001", "anthropic"},
}

// AvailableModels lists available models by provider
var AvailableModels = map[string][]string{
	"openai": {
		"gpt-5.2",
		"gpt-5.1",
		"gpt-5",
		"gpt-5-mini",
		"gpt-5-nano",
		"gpt-4.1",
		"gpt-4.1-mini",
		"gpt-4o",
		"gpt-4o-mini",
		"o3",
		"o4-mini",
	},
	"anthropic": {
		"claude-sonnet-4-5-20250929",
		"claude-opus-4-5-20251101",
		"claude-haiku-4-5-20251001",
	},
}

// NodeDisplayNames provides human-readable names for nodes
var NodeDisplayNames = map[string]string{
	NodeTriageOrchestrator: "Triage Orchestrator",
	NodeJobSearchWorker:    "Job Search Worker",
	NodeCRMWorker:          "CRM Worker",
	NodeGeneralWorker:      "General Worker",
	NodeEvaluator:          "Evaluator",
	NodeTitleGenerator:     "Title Generator",
}

// NodeDescriptions provides descriptions for each node
var NodeDescriptions = map[string]string{
	NodeTriageOrchestrator: "Routes requests to specialized workers",
	NodeJobSearchWorker:    "Searches for job listings and career opportunities",
	NodeCRMWorker:          "Handles CRM lookups and data queries",
	NodeGeneralWorker:      "Handles general questions and analysis",
	NodeEvaluator:          "Evaluates agent responses for quality",
	NodeTitleGenerator:     "Generates conversation titles",
}

// CostTier represents a cost tier level
const (
	CostTierBasic    = "basic"
	CostTierEconomy  = "economy"
	CostTierStandard = "standard"
	CostTierPremium  = "premium"
	CostTierMax      = "max"
)

// AllCostTiers lists all valid cost tiers in order
var AllCostTiers = []string{
	CostTierBasic,
	CostTierEconomy,
	CostTierStandard,
	CostTierPremium,
	CostTierMax,
}

// CostTierModelConfig defines models for a cost tier
type CostTierModelConfig struct {
	OpenAI    string
	Anthropic string
}

// CostTierModels maps cost tiers to model selections
var CostTierModels = map[string]CostTierModelConfig{
	CostTierBasic:    {"gpt-5-nano", "claude-haiku-4-5-20251001"},
	CostTierEconomy:  {"gpt-5-mini", "claude-haiku-4-5-20251001"},
	CostTierStandard: {"gpt-5", "claude-sonnet-4-5-20250929"},
	CostTierPremium:  {"gpt-5.1", "claude-sonnet-4-5-20250929"},
	CostTierMax:      {"gpt-5.2", "claude-opus-4-5-20251101"},
}

// CostTierDescriptions provides human-readable descriptions for cost tiers
var CostTierDescriptions = map[string]string{
	CostTierBasic:    "Cheapest, fast responses",
	CostTierEconomy:  "Budget-friendly",
	CostTierStandard: "Balanced cost and quality",
	CostTierPremium:  "Higher quality output",
	CostTierMax:      "Best available models",
}

// IsValidCostTier checks if a cost tier is valid
func IsValidCostTier(tier string) bool {
	for _, t := range AllCostTiers {
		if t == tier {
			return true
		}
	}
	return false
}

// IsValidNodeName checks if a node name is valid
func IsValidNodeName(nodeName string) bool {
	return slices.Contains(AllNodeNames, nodeName)
}

// GetProviderForModel returns the provider for a given model, or empty string if invalid
func GetProviderForModel(model string) string {
	for provider, modelList := range AvailableModels {
		if slices.Contains(modelList, model) {
			return provider
		}
	}
	return ""
}

type AgentConfigRepository struct {
	db *gorm.DB
}

func NewAgentConfigRepository(db *gorm.DB) *AgentConfigRepository {
	return &AgentConfigRepository{db: db}
}

// AgentNodeConfigDTO represents a single node configuration
type AgentNodeConfigDTO struct {
	NodeName         string   `json:"node_name"`
	Model            string   `json:"model"`
	Provider         string   `json:"provider"`
	Temperature      *float64 `json:"temperature,omitempty"`
	MaxTokens        *int     `json:"max_tokens,omitempty"`
	TopP             *float64 `json:"top_p,omitempty"`
	TopK             *int     `json:"top_k,omitempty"`
	FrequencyPenalty *float64 `json:"frequency_penalty,omitempty"`
	PresencePenalty  *float64 `json:"presence_penalty,omitempty"`
}

// GetUserConfigs returns all agent node configurations for a user
func (r *AgentConfigRepository) GetUserConfigs(userID int64) ([]AgentNodeConfigDTO, error) {
	var configs []models.AgentNodeConfig
	err := r.db.Where("user_id = ?", userID).Find(&configs).Error
	if err != nil {
		return nil, err
	}

	result := make([]AgentNodeConfigDTO, len(configs))
	for i, cfg := range configs {
		result[i] = AgentNodeConfigDTO{
			NodeName:         cfg.NodeName,
			Model:            cfg.Model,
			Provider:         cfg.Provider,
			Temperature:      cfg.Temperature,
			MaxTokens:        cfg.MaxTokens,
			TopP:             cfg.TopP,
			TopK:             cfg.TopK,
			FrequencyPenalty: cfg.FrequencyPenalty,
			PresencePenalty:  cfg.PresencePenalty,
		}
	}
	return result, nil
}

// GetNodeConfig returns a single node configuration for a user
func (r *AgentConfigRepository) GetNodeConfig(userID int64, nodeName string) (*AgentNodeConfigDTO, error) {
	var cfg models.AgentNodeConfig
	err := r.db.Where("user_id = ? AND node_name = ?", userID, nodeName).First(&cfg).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperrors.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &AgentNodeConfigDTO{
		NodeName:         cfg.NodeName,
		Model:            cfg.Model,
		Provider:         cfg.Provider,
		Temperature:      cfg.Temperature,
		MaxTokens:        cfg.MaxTokens,
		TopP:             cfg.TopP,
		TopK:             cfg.TopK,
		FrequencyPenalty: cfg.FrequencyPenalty,
		PresencePenalty:  cfg.PresencePenalty,
	}, nil
}

// NodeConfigUpdate contains fields to update for a node configuration
type NodeConfigUpdate struct {
	Model            string
	Provider         string
	Temperature      *float64
	MaxTokens        *int
	TopP             *float64
	TopK             *int
	FrequencyPenalty *float64
	PresencePenalty  *float64
}

// UpsertNodeConfig creates or updates a node configuration for a user
func (r *AgentConfigRepository) UpsertNodeConfig(userID int64, nodeName string, update NodeConfigUpdate) error {
	cfg := models.AgentNodeConfig{
		UserID:           userID,
		NodeName:         nodeName,
		Model:            update.Model,
		Provider:         update.Provider,
		Temperature:      update.Temperature,
		MaxTokens:        update.MaxTokens,
		TopP:             update.TopP,
		TopK:             update.TopK,
		FrequencyPenalty: update.FrequencyPenalty,
		PresencePenalty:  update.PresencePenalty,
	}

	return r.db.Where("user_id = ? AND node_name = ?", userID, nodeName).
		Assign(models.AgentNodeConfig{
			Model:            update.Model,
			Provider:         update.Provider,
			Temperature:      update.Temperature,
			MaxTokens:        update.MaxTokens,
			TopP:             update.TopP,
			TopK:             update.TopK,
			FrequencyPenalty: update.FrequencyPenalty,
			PresencePenalty:  update.PresencePenalty,
		}).
		FirstOrCreate(&cfg).Error
}

// DeleteNodeConfig removes a node configuration (reverts to default)
func (r *AgentConfigRepository) DeleteNodeConfig(userID int64, nodeName string) error {
	result := r.db.Where("user_id = ? AND node_name = ?", userID, nodeName).Delete(&models.AgentNodeConfig{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

// PresetDTO represents a preset configuration
type PresetDTO struct {
	ID               int64
	UserID           *int64 // NULL = global preset
	Name             string
	Temperature      *float64
	MaxTokens        *int
	TopP             *float64
	TopK             *int
	FrequencyPenalty *float64
	PresencePenalty  *float64
}

// PresetCreateInput contains fields to create a preset
type PresetCreateInput struct {
	UserID           int64
	Name             string
	Temperature      *float64
	MaxTokens        *int
	TopP             *float64
	TopK             *int
	FrequencyPenalty *float64
	PresencePenalty  *float64
}

// GetUserPresets returns global presets and user's own presets
func (r *AgentConfigRepository) GetUserPresets(userID int64) ([]PresetDTO, error) {
	var presets []models.AgentConfigPreset
	err := r.db.Where("user_id IS NULL OR user_id = ?", userID).Order("user_id NULLS FIRST, name").Find(&presets).Error
	if err != nil {
		return nil, err
	}

	result := make([]PresetDTO, len(presets))
	for i, p := range presets {
		result[i] = PresetDTO{
			ID:               p.ID,
			UserID:           p.UserID,
			Name:             p.Name,
			Temperature:      p.Temperature,
			MaxTokens:        p.MaxTokens,
			TopP:             p.TopP,
			TopK:             p.TopK,
			FrequencyPenalty: p.FrequencyPenalty,
			PresencePenalty:  p.PresencePenalty,
		}
	}
	return result, nil
}

// GetPresetByID returns a preset by ID (allows global or user's own)
func (r *AgentConfigRepository) GetPresetByID(userID, presetID int64) (*PresetDTO, error) {
	var preset models.AgentConfigPreset
	err := r.db.Where("id = ? AND (user_id IS NULL OR user_id = ?)", presetID, userID).First(&preset).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperrors.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &PresetDTO{
		ID:               preset.ID,
		UserID:           preset.UserID,
		Name:             preset.Name,
		Temperature:      preset.Temperature,
		MaxTokens:        preset.MaxTokens,
		TopP:             preset.TopP,
		TopK:             preset.TopK,
		FrequencyPenalty: preset.FrequencyPenalty,
		PresencePenalty:  preset.PresencePenalty,
	}, nil
}

// CreatePreset creates a new preset
func (r *AgentConfigRepository) CreatePreset(input PresetCreateInput) (int64, error) {
	preset := models.AgentConfigPreset{
		UserID:           &input.UserID,
		Name:             input.Name,
		Temperature:      input.Temperature,
		MaxTokens:        input.MaxTokens,
		TopP:             input.TopP,
		TopK:             input.TopK,
		FrequencyPenalty: input.FrequencyPenalty,
		PresencePenalty:  input.PresencePenalty,
	}
	err := r.db.Create(&preset).Error
	if err != nil {
		return 0, err
	}
	return preset.ID, nil
}

// DeletePreset deletes a preset by ID (validates user ownership)
func (r *AgentConfigRepository) DeletePreset(userID, presetID int64) error {
	result := r.db.Where("id = ? AND user_id = ?", presetID, userID).Delete(&models.AgentConfigPreset{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

// ApplyPresetToNodes applies preset settings to all nodes for a user
func (r *AgentConfigRepository) ApplyPresetToNodes(userID int64, preset PresetDTO) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for _, nodeName := range AllNodeNames {
			defaultCfg := DefaultModels[nodeName]
			cfg := models.AgentNodeConfig{
				UserID:   userID,
				NodeName: nodeName,
				Model:    defaultCfg.Model,
				Provider: defaultCfg.Provider,
			}
			presetUpdate := models.AgentNodeConfig{
				Temperature:      preset.Temperature,
				MaxTokens:        preset.MaxTokens,
				TopP:             preset.TopP,
				TopK:             preset.TopK,
				FrequencyPenalty: preset.FrequencyPenalty,
				PresencePenalty:  preset.PresencePenalty,
			}
			err := tx.Where("user_id = ? AND node_name = ?", userID, nodeName).
				Assign(presetUpdate).
				FirstOrCreate(&cfg).Error
			if err != nil {
				return err
			}
		}
		return nil
	})
}

// UserSelectionDTO represents user's selected presets
type UserSelectionDTO struct {
	SelectedParamPresetID *int64
	SelectedCostTier      string
}

// GetUserSelection returns user's selected presets
func (r *AgentConfigRepository) GetUserSelection(userID int64) (*UserSelectionDTO, error) {
	var selection models.AgentConfigUserSelection
	err := r.db.Where("user_id = ?", userID).First(&selection).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &UserSelectionDTO{SelectedCostTier: CostTierStandard}, nil
	}
	if err != nil {
		return nil, err
	}
	return &UserSelectionDTO{
		SelectedParamPresetID: selection.SelectedParamPresetID,
		SelectedCostTier:      selection.SelectedCostTier,
	}, nil
}

// SetUserSelection updates user's selected presets
func (r *AgentConfigRepository) SetUserSelection(userID int64, presetID *int64, costTier string) error {
	selection := models.AgentConfigUserSelection{
		UserID:                userID,
		SelectedParamPresetID: presetID,
		SelectedCostTier:      costTier,
	}
	return r.db.Save(&selection).Error
}

// SetSelectedParamPreset updates only the selected param preset
func (r *AgentConfigRepository) SetSelectedParamPreset(userID int64, presetID *int64) error {
	selection := models.AgentConfigUserSelection{
		UserID:                userID,
		SelectedParamPresetID: presetID,
		SelectedCostTier:      CostTierStandard,
	}
	return r.db.Where("user_id = ?", userID).
		Assign(models.AgentConfigUserSelection{SelectedParamPresetID: presetID}).
		FirstOrCreate(&selection).Error
}

// SetSelectedCostTier updates only the selected cost tier
func (r *AgentConfigRepository) SetSelectedCostTier(userID int64, costTier string) error {
	selection := models.AgentConfigUserSelection{
		UserID:           userID,
		SelectedCostTier: costTier,
	}
	return r.db.Where("user_id = ?", userID).
		Assign(models.AgentConfigUserSelection{SelectedCostTier: costTier}).
		FirstOrCreate(&selection).Error
}

// getModelForTier returns the model and provider for a given tier and base provider
func getModelForTier(tierConfig CostTierModelConfig, baseProvider string) (model, provider string) {
	providerModels := map[string]string{
		"anthropic": tierConfig.Anthropic,
		"openai":    tierConfig.OpenAI,
	}
	model = providerModels[baseProvider]
	provider = baseProvider
	return model, provider
}

// ApplyCostTierToNodes applies cost tier model selections to all nodes for a user
func (r *AgentConfigRepository) ApplyCostTierToNodes(userID int64, tier string) error {
	tierConfig := CostTierModels[tier]

	return r.db.Transaction(func(tx *gorm.DB) error {
		for _, nodeName := range AllNodeNames {
			defaultCfg := DefaultModels[nodeName]
			model, provider := getModelForTier(tierConfig, defaultCfg.Provider)

			cfg := models.AgentNodeConfig{
				UserID:   userID,
				NodeName: nodeName,
				Model:    model,
				Provider: provider,
			}
			err := tx.Where("user_id = ? AND node_name = ?", userID, nodeName).
				Assign(models.AgentNodeConfig{Model: model, Provider: provider}).
				FirstOrCreate(&cfg).Error
			if err != nil {
				return err
			}
		}
		return nil
	})
}
