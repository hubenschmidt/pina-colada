package repositories

import (
	"errors"

	"gorm.io/gorm"

	apperrors "github.com/pina-colada-co/agent-go/internal/errors"
	"github.com/pina-colada-co/agent-go/internal/models"
)

type AgentConfigRepository struct {
	db *gorm.DB
}

func NewAgentConfigRepository(db *gorm.DB) *AgentConfigRepository {
	return &AgentConfigRepository{db: db}
}

// AgentNodeConfigDTO represents a single node configuration
type AgentNodeConfigDTO struct {
	NodeName string `json:"node_name"`
	Model    string `json:"model"`
	Provider string `json:"provider"`
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
			NodeName: cfg.NodeName,
			Model:    cfg.Model,
			Provider: cfg.Provider,
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
		NodeName: cfg.NodeName,
		Model:    cfg.Model,
		Provider: cfg.Provider,
	}, nil
}

// UpsertNodeConfig creates or updates a node configuration for a user
func (r *AgentConfigRepository) UpsertNodeConfig(userID int64, nodeName, model, provider string) error {
	cfg := models.AgentNodeConfig{
		UserID:   userID,
		NodeName: nodeName,
		Model:    model,
		Provider: provider,
	}

	return r.db.Where("user_id = ? AND node_name = ?", userID, nodeName).
		Assign(models.AgentNodeConfig{Model: model, Provider: provider}).
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
