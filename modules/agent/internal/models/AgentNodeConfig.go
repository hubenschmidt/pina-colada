package models

import (
	"time"
)

type AgentNodeConfig struct {
	ID               int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID           int64          `gorm:"not null" json:"user_id"`
	NodeName         string         `gorm:"not null" json:"node_name"`
	Model            string         `gorm:"not null" json:"model"`
	Provider         string         `gorm:"not null;default:openai" json:"provider"`
	Temperature      *float64       `gorm:"column:temperature" json:"temperature,omitempty"`
	MaxTokens        *int           `gorm:"column:max_tokens" json:"max_tokens,omitempty"`
	TopP             *float64       `gorm:"column:top_p" json:"top_p,omitempty"`
	TopK             *int           `gorm:"column:top_k" json:"top_k,omitempty"`
	FrequencyPenalty *float64       `gorm:"column:frequency_penalty" json:"frequency_penalty,omitempty"`
	PresencePenalty *float64  `gorm:"column:presence_penalty" json:"presence_penalty,omitempty"`
	CreatedAt       time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
}

func (AgentNodeConfig) TableName() string {
	return "Agent_Node_Config"
}

// Node name constants
const (
	NodeTriageOrchestrator = "triage_orchestrator"
	NodeJobSearchWorker    = "job_search_worker"
	NodeCRMWorker          = "crm_worker"
	NodeGeneralWorker      = "general_worker"
	NodeWriterWorker       = "writer_worker"
	NodeEvaluator          = "evaluator"
	NodeTitleGenerator     = "title_generator"
)

// AllNodeNames returns all configurable node names
var AllNodeNames = []string{
	NodeTriageOrchestrator,
	NodeJobSearchWorker,
	NodeCRMWorker,
	NodeGeneralWorker,
	NodeWriterWorker,
	NodeEvaluator,
	NodeTitleGenerator,
}

// DefaultModels defines the default model for each node
var DefaultModels = map[string]struct {
	Model    string
	Provider string
}{
	NodeTriageOrchestrator: {"gpt-5.2", "openai"},
	NodeJobSearchWorker:    {"gpt-5.2", "openai"},
	NodeCRMWorker:          {"gpt-5.2", "openai"},
	NodeGeneralWorker:      {"gpt-5.2", "openai"},
	NodeWriterWorker:       {"gpt-5.2", "openai"},
	NodeEvaluator:          {"claude-sonnet-4-5-20250929", "anthropic"},
	NodeTitleGenerator:     {"claude-haiku-4-5-20251001", "anthropic"},
}

// NodeDisplayNames provides human-readable names for nodes
var NodeDisplayNames = map[string]string{
	NodeTriageOrchestrator: "Triage Orchestrator",
	NodeJobSearchWorker:    "Job Search Worker",
	NodeCRMWorker:          "CRM Worker",
	NodeGeneralWorker:      "General Worker",
	NodeWriterWorker:       "Writer Worker",
	NodeEvaluator:          "Evaluator",
	NodeTitleGenerator:     "Title Generator",
}

// NodeDescriptions provides descriptions for each node
var NodeDescriptions = map[string]string{
	NodeTriageOrchestrator: "Routes requests to specialized workers",
	NodeJobSearchWorker:    "Searches for job listings and career opportunities",
	NodeCRMWorker:          "Handles CRM lookups and data queries",
	NodeGeneralWorker:      "Handles general questions and analysis",
	NodeWriterWorker:       "Generates documents using existing examples as templates",
	NodeEvaluator:          "Evaluates agent responses for quality",
	NodeTitleGenerator:     "Generates conversation titles",
}
