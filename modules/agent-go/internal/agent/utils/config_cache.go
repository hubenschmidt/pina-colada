package utils

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/pina-colada-co/agent-go/internal/agent/promoter"
	"github.com/pina-colada-co/agent-go/internal/models"
	"github.com/pina-colada-co/agent-go/internal/repositories"
	"github.com/pina-colada-co/agent-go/internal/services"
)

// ConfigCache provides cached access to agent node configurations per user
type ConfigCache struct {
	service *services.AgentConfigService
	cache   map[int64]map[string]nodeConfig // userID -> nodeName -> config
	mu      sync.RWMutex
}

type nodeConfig struct {
	model            string
	provider         string
	temperature      *float64
	maxTokens        *int
	topP             *float64
	topK             *int
	frequencyPenalty *float64
	presencePenalty  *float64
	fallbackChain    []promoter.ModelTier
}

// NewConfigCache creates a new config cache
func NewConfigCache(service *services.AgentConfigService) *ConfigCache {
	return &ConfigCache{
		service: service,
		cache:   make(map[int64]map[string]nodeConfig),
	}
}

// GetModel returns the model for a specific node for a user
// Falls back to default if no custom config exists
func (c *ConfigCache) GetModel(userID int64, nodeName string) string {
	if model := c.getCachedModel(userID, nodeName); model != "" {
		return model
	}

	c.loadUser(userID)

	if model := c.getCachedModel(userID, nodeName); model != "" {
		return model
	}

	return repositories.DefaultModels[nodeName].Model
}

func (c *ConfigCache) getCachedModel(userID int64, nodeName string) string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	userCache, ok := c.cache[userID]
	if !ok {
		return ""
	}

	cfg, ok := userCache[nodeName]
	if !ok {
		return ""
	}

	return cfg.model
}

// GetProvider returns the provider for a specific node for a user
func (c *ConfigCache) GetProvider(userID int64, nodeName string) string {
	if provider := c.getCachedProvider(userID, nodeName); provider != "" {
		return provider
	}

	c.loadUser(userID)

	if provider := c.getCachedProvider(userID, nodeName); provider != "" {
		return provider
	}

	return repositories.DefaultModels[nodeName].Provider
}

func (c *ConfigCache) getCachedProvider(userID int64, nodeName string) string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	userCache, ok := c.cache[userID]
	if !ok {
		return ""
	}

	cfg, ok := userCache[nodeName]
	if !ok {
		return ""
	}

	return cfg.provider
}

// GetSettings returns the LLM settings for a specific node for a user
func (c *ConfigCache) GetSettings(userID int64, nodeName string) services.LLMSettings {
	settings := c.getCachedSettings(userID, nodeName)
	if settings != nil {
		return *settings
	}

	c.loadUser(userID)

	settings = c.getCachedSettings(userID, nodeName)
	if settings != nil {
		return *settings
	}

	return services.LLMSettings{}
}

func (c *ConfigCache) getCachedSettings(userID int64, nodeName string) *services.LLMSettings {
	c.mu.RLock()
	defer c.mu.RUnlock()

	userCache, ok := c.cache[userID]
	if !ok {
		return nil
	}

	cfg, ok := userCache[nodeName]
	if !ok {
		return nil
	}

	return &services.LLMSettings{
		Temperature:      cfg.temperature,
		MaxTokens:        cfg.maxTokens,
		TopP:             cfg.topP,
		TopK:             cfg.topK,
		FrequencyPenalty: cfg.frequencyPenalty,
		PresencePenalty:  cfg.presencePenalty,
	}
}

// Invalidate removes a user's cached config (call after updates)
func (c *ConfigCache) Invalidate(userID int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.cache, userID)
}

// GetModelChain returns the fallback chain for a specific node for a user.
// Returns nil if no custom fallback chain is configured (promotion disabled by default).
func (c *ConfigCache) GetModelChain(userID int64, nodeName string) []promoter.ModelTier {
	chain := c.getCachedModelChain(userID, nodeName)
	if len(chain) > 0 {
		return chain
	}

	c.loadUser(userID)

	// Only return explicitly configured chains - no default promotion
	return c.getCachedModelChain(userID, nodeName)
}

func (c *ConfigCache) getCachedModelChain(userID int64, nodeName string) []promoter.ModelTier {
	c.mu.RLock()
	defer c.mu.RUnlock()

	userCache, ok := c.cache[userID]
	if !ok {
		return nil
	}

	cfg, ok := userCache[nodeName]
	if !ok {
		return nil
	}

	return cfg.fallbackChain
}

// loadUser loads a user's config from the service into the cache
func (c *ConfigCache) loadUser(userID int64) {
	if c.service == nil {
		return
	}

	config, err := c.service.GetAgentConfig(userID)
	if err != nil {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	userCache := make(map[string]nodeConfig)
	for _, node := range config.Nodes {
		userCache[node.NodeName] = nodeConfig{
			model:            node.Model,
			provider:         node.Provider,
			temperature:      node.Temperature,
			maxTokens:        node.MaxTokens,
			topP:             node.TopP,
			topK:             node.TopK,
			frequencyPenalty: node.FrequencyPenalty,
			presencePenalty:  node.PresencePenalty,
			fallbackChain:    parseFallbackChain(node.FallbackChain),
		}
	}
	c.cache[userID] = userCache
}

// parseFallbackChain parses JSONB fallback chain data into ModelTier slice
func parseFallbackChain(data []byte) []promoter.ModelTier {
	if len(data) == 0 {
		return nil
	}

	var tiers []models.FallbackTier
	if err := json.Unmarshal(data, &tiers); err != nil {
		return nil
	}

	result := make([]promoter.ModelTier, len(tiers))
	for i, t := range tiers {
		result[i] = promoter.ModelTier{
			Model:             t.Model,
			FirstTokenTimeout: time.Duration(t.TimeoutSeconds) * time.Second,
		}
	}
	return result
}
