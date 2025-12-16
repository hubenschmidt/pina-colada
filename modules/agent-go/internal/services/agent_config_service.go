package services

import (
	"errors"
	"slices"

	"github.com/pina-colada-co/agent-go/internal/models"
	"github.com/pina-colada-co/agent-go/internal/repositories"
)

var ErrInvalidNodeName = errors.New("invalid node name")
var ErrInvalidModel = errors.New("invalid model for provider")

type AgentConfigService struct {
	configRepo *repositories.AgentConfigRepository
}

func NewAgentConfigService(configRepo *repositories.AgentConfigRepository) *AgentConfigService {
	return &AgentConfigService{configRepo: configRepo}
}

// NodeConfigResponse represents a single node's configuration
type NodeConfigResponse struct {
	NodeName    string `json:"node_name"`
	DisplayName string `json:"display_name"`
	Description string `json:"description"`
	Model       string `json:"model"`
	Provider    string `json:"provider"`
	IsDefault   bool   `json:"is_default"`
}

// AgentConfigResponse contains all node configurations
type AgentConfigResponse struct {
	Nodes []NodeConfigResponse `json:"nodes"`
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

	// Build a map of user overrides
	overrides := make(map[string]repositories.AgentNodeConfigDTO)
	for _, cfg := range userConfigs {
		overrides[cfg.NodeName] = cfg
	}

	// Build response with all nodes
	nodes := make([]NodeConfigResponse, 0, len(models.AllNodeNames))
	for _, nodeName := range models.AllNodeNames {
		defaultCfg := models.DefaultModels[nodeName]
		isDefault := true
		model := defaultCfg.Model
		provider := defaultCfg.Provider

		if override, ok := overrides[nodeName]; ok {
			model = override.Model
			provider = override.Provider
			isDefault = false
		}

		nodes = append(nodes, NodeConfigResponse{
			NodeName:    nodeName,
			DisplayName: models.NodeDisplayNames[nodeName],
			Description: models.NodeDescriptions[nodeName],
			Model:       model,
			Provider:    provider,
			IsDefault:   isDefault,
		})
	}

	return &AgentConfigResponse{Nodes: nodes}, nil
}

// UpdateNodeConfig updates a single node's model configuration
func (s *AgentConfigService) UpdateNodeConfig(userID int64, nodeName, model string) (*NodeConfigResponse, error) {
	// Validate node name
	if !slices.Contains(models.AllNodeNames, nodeName) {
		return nil, ErrInvalidNodeName
	}

	// Determine provider from model
	provider := s.getProviderForModel(model)
	if provider == "" {
		return nil, ErrInvalidModel
	}

	// Upsert the config
	err := s.configRepo.UpsertNodeConfig(userID, nodeName, model, provider)
	if err != nil {
		return nil, err
	}

	return &NodeConfigResponse{
		NodeName:    nodeName,
		DisplayName: models.NodeDisplayNames[nodeName],
		Description: models.NodeDescriptions[nodeName],
		Model:       model,
		Provider:    provider,
		IsDefault:   false,
	}, nil
}

// ResetNodeConfig removes user override and reverts to default
func (s *AgentConfigService) ResetNodeConfig(userID int64, nodeName string) (*NodeConfigResponse, error) {
	// Validate node name
	if !slices.Contains(models.AllNodeNames, nodeName) {
		return nil, ErrInvalidNodeName
	}

	// Delete the override (ignore not found)
	_ = s.configRepo.DeleteNodeConfig(userID, nodeName)

	// Return default config
	defaultCfg := models.DefaultModels[nodeName]
	return &NodeConfigResponse{
		NodeName:    nodeName,
		DisplayName: models.NodeDisplayNames[nodeName],
		Description: models.NodeDescriptions[nodeName],
		Model:       defaultCfg.Model,
		Provider:    defaultCfg.Provider,
		IsDefault:   true,
	}, nil
}

// GetAvailableModels returns all available models grouped by provider
func (s *AgentConfigService) GetAvailableModels() *AvailableModelsResponse {
	return &AvailableModelsResponse{
		OpenAI:    models.AvailableModels["openai"],
		Anthropic: models.AvailableModels["anthropic"],
	}
}

// GetModelForNode returns the model for a specific node for a user
func (s *AgentConfigService) GetModelForNode(userID int64, nodeName string) (model, provider string) {
	cfg, err := s.configRepo.GetNodeConfig(userID, nodeName)
	if err != nil || cfg == nil {
		// Return default
		defaultCfg := models.DefaultModels[nodeName]
		return defaultCfg.Model, defaultCfg.Provider
	}
	return cfg.Model, cfg.Provider
}

// getProviderForModel determines which provider a model belongs to
func (s *AgentConfigService) getProviderForModel(model string) string {
	for provider, modelList := range models.AvailableModels {
		if slices.Contains(modelList, model) {
			return provider
		}
	}
	return ""
}
