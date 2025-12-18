# Spec: Automatic Model Promotion for Slow Requests

## Overview
When an LLM takes too long to respond, automatically cancel and retry with a faster model. Uses first-token latency as the promotion trigger to avoid mid-stream disruption.

## Code Rules Compliance
- **Go Code Rules**: Guard clauses, early returns, no nested conditionals, KISS/YAGNI
- **CSR Pattern**: Repository defines DTOs, Services call repositories, no direct model imports in services

## Problem
- gpt-5 can take 50-70+ seconds to process complex requests
- Users experience unacceptable wait times
- Faster models (gpt-5.1, gpt-5.2) have lower latency but may be more expensive

## Design: First-Token Timeout Promotion

### Key Insight
For streaming responses, we can't switch mid-stream without disrupting the user experience. Instead:
1. Set a **first-token timeout** (e.g., 15 seconds)
2. If no tokens received within timeout, cancel and try faster model
3. Once tokens start flowing, let the response complete

### Model Escalation Chain
```go
var defaultModelChain = []ModelTier{
    {Model: "gpt-4o", FirstTokenTimeout: 15 * time.Second},
    {Model: "gpt-4o-mini", FirstTokenTimeout: 10 * time.Second},
    {Model: "gpt-3.5-turbo", FirstTokenTimeout: 5 * time.Second},
}
```

### Generic Utility: `ModelPromoter` (internal/agent/promoter/promoter.go)

```go
package promoter

import (
    "context"
    "errors"
    "fmt"
    "log"
    "time"
)

var ErrFirstTokenTimeout = errors.New("first token timeout exceeded")

type ModelTier struct {
    Model             string
    FirstTokenTimeout time.Duration
}

type StreamEvent struct {
    Type string
    // ... other fields from agent.StreamEvent
}

type RunFunc func(ctx context.Context, model string) (<-chan StreamEvent, <-chan error, error)

type ModelPromoter struct {
    tiers []ModelTier
}

func NewModelPromoter(tiers []ModelTier) *ModelPromoter {
    return &ModelPromoter{tiers: tiers}
}

// RunStreamWithPromotion runs streaming with automatic model promotion.
// Uses guard clauses per go-code-rules.
func (p *ModelPromoter) RunStreamWithPromotion(ctx context.Context, runFn RunFunc, eventCh chan<- StreamEvent) error {
    for i, tier := range p.tiers {
        err := p.tryTier(ctx, tier, runFn, eventCh)

        // Success - return immediately (guard clause)
        if err == nil {
            return nil
        }

        // Non-timeout error - don't promote (guard clause)
        if !errors.Is(err, ErrFirstTokenTimeout) {
            return err
        }

        // Last tier exhausted - return error (guard clause)
        if i == len(p.tiers)-1 {
            return fmt.Errorf("all model tiers exhausted: %w", err)
        }

        log.Printf("Model %s timed out, promoting to %s", tier.Model, p.tiers[i+1].Model)
    }
    return fmt.Errorf("no model tiers configured")
}

func (p *ModelPromoter) tryTier(ctx context.Context, tier ModelTier, runFn RunFunc, eventCh chan<- StreamEvent) error {
    timeoutCtx, cancel := context.WithCancel(ctx)
    defer cancel()

    streamCh, errCh, err := runFn(timeoutCtx, tier.Model)
    if err != nil {
        return err
    }

    timer := time.NewTimer(tier.FirstTokenTimeout)
    defer timer.Stop()

    return p.processStream(ctx, timeoutCtx, cancel, streamCh, errCh, eventCh, timer)
}

func (p *ModelPromoter) processStream(
    ctx, timeoutCtx context.Context,
    cancel context.CancelFunc,
    streamCh <-chan StreamEvent,
    errCh <-chan error,
    eventCh chan<- StreamEvent,
    timer *time.Timer,
) error {
    firstTokenReceived := false

    for {
        select {
        case evt, ok := <-streamCh:
            if !ok {
                return <-errCh // Stream completed
            }
            firstTokenReceived = p.handleEvent(evt, firstTokenReceived, timer, eventCh)

        case <-timer.C:
            if firstTokenReceived {
                continue // Already receiving tokens, ignore timer
            }
            cancel()
            return ErrFirstTokenTimeout

        case <-ctx.Done():
            return ctx.Err()
        }
    }
}

func (p *ModelPromoter) handleEvent(evt StreamEvent, firstTokenReceived bool, timer *time.Timer, eventCh chan<- StreamEvent) bool {
    eventCh <- evt

    if firstTokenReceived {
        return true
    }
    if evt.Type != "text" {
        return false
    }

    // First text token received - stop timeout
    timer.Stop()
    return true
}
```

## Database-Driven Available Models

### Problem
`AvailableModels` is hardcoded in `AgentNodeConfig.go`:
```go
var AvailableModels = map[string][]string{
    "openai": {"gpt-5.2", "gpt-5.1", ...},
    "anthropic": {"claude-sonnet-4-5-20250929", ...},
}
```
Adding/removing models requires code changes and redeployment.

### Solution: `Available_Model` table
```sql
CREATE TABLE "Available_Model" (
    id BIGSERIAL PRIMARY KEY,
    provider VARCHAR(50) NOT NULL,        -- "openai", "anthropic"
    model_name VARCHAR(100) NOT NULL,     -- "gpt-5.2"
    display_name VARCHAR(100),            -- "GPT 5.2" (for UI)
    sort_order INT NOT NULL DEFAULT 0,    -- For ordering in dropdowns
    is_active BOOLEAN DEFAULT true,       -- Soft disable without deletion
    default_timeout_seconds INT DEFAULT 20, -- Default first-token timeout
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(provider, model_name)
);

-- Seed data
INSERT INTO "Available_Model" (provider, model_name, display_name, sort_order, default_timeout_seconds) VALUES
('openai', 'gpt-5.2', 'GPT 5.2', 1, 10),
('openai', 'gpt-5.1', 'GPT 5.1', 2, 15),
('openai', 'gpt-5', 'GPT 5', 3, 20),
('openai', 'gpt-5-mini', 'GPT 5 Mini', 4, 10),
('openai', 'gpt-5-nano', 'GPT 5 Nano', 5, 5),
('openai', 'gpt-4.1', 'GPT 4.1', 6, 15),
('openai', 'gpt-4.1-mini', 'GPT 4.1 Mini', 7, 10),
('openai', 'gpt-4o', 'GPT 4o', 8, 15),
('openai', 'gpt-4o-mini', 'GPT 4o Mini', 9, 10),
('openai', 'o3', 'O3', 10, 30),
('openai', 'o4-mini', 'O4 Mini', 11, 15),
('anthropic', 'claude-sonnet-4-5-20250929', 'Claude Sonnet 4.5', 1, 15),
('anthropic', 'claude-opus-4-5-20251101', 'Claude Opus 4.5', 2, 30),
('anthropic', 'claude-haiku-4-5-20251001', 'Claude Haiku 4.5', 3, 5);
```

### New Model: `AvailableModel` (models/AvailableModel.go)
```go
package models

import "time"

type AvailableModel struct {
    ID                    int64     `gorm:"primaryKey" json:"id"`
    Provider              string    `gorm:"not null" json:"provider"`
    ModelName             string    `gorm:"not null" json:"model_name"`
    DisplayName           string    `json:"display_name"`
    SortOrder             int       `gorm:"default:0" json:"sort_order"`
    IsActive              bool      `gorm:"default:true" json:"is_active"`
    DefaultTimeoutSeconds int       `gorm:"default:20" json:"default_timeout_seconds"`
    CreatedAt             time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (AvailableModel) TableName() string {
    return "Available_Model"
}
```

### New Repository: `AvailableModelRepository` (repositories/available_model_repository.go)
```go
package repositories

import (
    "github.com/pina-colada-co/agent-go/internal/models"
    "gorm.io/gorm"
)

// DTO for service layer (services don't import models)
type AvailableModelDTO struct {
    ID                    int64
    Provider              string
    ModelName             string
    DisplayName           string
    SortOrder             int
    IsActive              bool
    DefaultTimeoutSeconds int
}

type AvailableModelRepository struct {
    db *gorm.DB
}

func NewAvailableModelRepository(db *gorm.DB) *AvailableModelRepository {
    return &AvailableModelRepository{db: db}
}

func (r *AvailableModelRepository) GetActiveModels() ([]AvailableModelDTO, error) {
    var models []models.AvailableModel
    err := r.db.Where("is_active = ?", true).Order("provider, sort_order").Find(&models).Error
    if err != nil {
        return nil, err
    }
    return toAvailableModelDTOs(models), nil
}

func (r *AvailableModelRepository) GetModelsByProvider(provider string) ([]AvailableModelDTO, error) {
    var models []models.AvailableModel
    err := r.db.Where("provider = ? AND is_active = ?", provider, true).Order("sort_order").Find(&models).Error
    if err != nil {
        return nil, err
    }
    return toAvailableModelDTOs(models), nil
}

func (r *AvailableModelRepository) GetModel(provider, modelName string) (*AvailableModelDTO, error) {
    var model models.AvailableModel
    err := r.db.Where("provider = ? AND model_name = ?", provider, modelName).First(&model).Error
    if err == gorm.ErrRecordNotFound {
        return nil, nil // Not found = nil, nil per CSR rules
    }
    if err != nil {
        return nil, err
    }
    return toAvailableModelDTO(&model), nil
}

func toAvailableModelDTO(m *models.AvailableModel) *AvailableModelDTO {
    return &AvailableModelDTO{
        ID: m.ID, Provider: m.Provider, ModelName: m.ModelName,
        DisplayName: m.DisplayName, SortOrder: m.SortOrder,
        IsActive: m.IsActive, DefaultTimeoutSeconds: m.DefaultTimeoutSeconds,
    }
}

func toAvailableModelDTOs(models []models.AvailableModel) []AvailableModelDTO {
    dtos := make([]AvailableModelDTO, len(models))
    for i, m := range models {
        dtos[i] = *toAvailableModelDTO(&m)
    }
    return dtos
}
```

### Refactor `AgentNodeConfig.go`
Remove hardcoded maps, use DB lookups:
```go
// REMOVE:
// var AvailableModels = map[string][]string{...}

// KEEP (as fallback/validation):
var DefaultModels = map[string]struct{Model, Provider string}{...}
```

---

## Database-Driven Model Chains

### Extend `Agent_Node_Config`
```sql
ALTER TABLE "Agent_Node_Config"
ADD COLUMN fallback_chain JSONB;
```

```go
type AgentNodeConfig struct {
    // ... existing fields ...
    FallbackChain datatypes.JSON `json:"fallback_chain,omitempty"` // NEW
}

// FallbackChain format:
// [{"model": "gpt-5.1", "timeout_seconds": 15}, {"model": "gpt-5.2", "timeout_seconds": 10}]
```

### Default Chains (when DB is null)
```go
var DefaultFallbackChains = map[string][]ModelTier{
    NodeJobSearchWorker: {
        {Model: "gpt-5", FirstTokenTimeout: 20 * time.Second},
        {Model: "gpt-5.1", FirstTokenTimeout: 15 * time.Second},
        {Model: "gpt-5.2", FirstTokenTimeout: 10 * time.Second},
    },
}
```

### ConfigCache Integration
```go
func (c *ConfigCache) GetModelChain(userID int64, nodeName string) []ModelTier {
    config := c.getConfig(userID, nodeName)
    if config == nil || config.FallbackChain == nil {
        return DefaultFallbackChains[nodeName]
    }
    return parseFallbackChain(config.FallbackChain)
}
```

## Modified `runAgentStream` Flow

```go
func (o *Orchestrator) runAgentStreamWithPromotion(
    ctx context.Context,
    userID int64,
    input string,
    useEvaluator bool,
    sendEvent func(StreamEvent),
) (*streamState, error) {
    modelChain := o.getModelChain(userID)
    promoter := NewModelPromoter(modelChain)

    // Wrap the actual run function
    runFn := func(ctx context.Context, model string) (<-chan StreamEvent, <-chan error, error) {
        // Build agent with specific model
        agent := o.buildAgentWithModel(userID, model)
        runner := agents.Runner{Config: agents.RunConfig{MaxTurns: 20}}
        return runner.RunStreamedChan(ctx, agent, input)
    }

    internalCh := make(chan StreamEvent, 100)
    ss := &streamState{useEvaluator: useEvaluator, sendEvent: sendEvent}

    go func() {
        defer close(internalCh)
        promoter.RunStreamWithPromotion(ctx, runFn, internalCh)
    }()

    for evt := range internalCh {
        ss.handleStreamEvent(evt)
    }

    return ss, nil
}
```

## Files to Create/Modify (CSR Compliant)

| Layer | File | Action |
|-------|------|--------|
| **Migrations** | `modules/agent/migrations/070_available_model.sql` | NEW - Create Available_Model table with seed data |
| **Migrations** | `modules/agent/migrations/071_agent_config_fallback_chain.sql` | NEW - Add fallback_chain column |
| **Models** | `internal/models/AvailableModel.go` | NEW - AvailableModel struct with TableName() |
| **Models** | `internal/models/AgentNodeConfig.go` | Add FallbackChain field |
| **Repository** | `internal/repositories/available_model_repository.go` | NEW - CRUD with DTOs (services don't import models) |
| **Service** | `internal/services/available_model_service.go` | NEW - Business logic, calls repository |
| **Agent** | `internal/agent/promoter/promoter.go` | NEW - ModelPromoter utility (guard clauses) |
| **Agent** | `internal/agent/utils/config_cache.go` | Add GetModelChain method |
| **Agent** | `internal/agent/orchestrator.go` | Integrate promoter into streaming flow |
| **Main** | `cmd/agent/main.go` | Initialize AvailableModelRepository + Service |

### CSR Import Rules Compliance
```
orchestrator.go → services, promoter (NOT models)
services/available_model_service.go → repositories (uses DTOs, NOT models)
repositories/available_model_repository.go → models
```

## Metrics
Track promotion events:
- Which model was promoted from/to
- Time spent before promotion
- Final model used for completion

## Edge Cases
1. **All models timeout**: Return error after exhausting chain
2. **Partial tokens before timeout**: Once any text token received, no promotion
3. **Tool calls without text**: Tool events don't reset timeout (waiting for LLM response)
4. **User-configured fast model**: If user already configured fast model, chain has 1 tier

## Benefits
1. **Automatic latency optimization** - Slow requests automatically upgrade
2. **User experience** - No 70+ second waits
3. **Generic** - Any worker can use the promoter
4. **Observable** - Metrics show promotion patterns
5. **Configurable** - Timeouts and chains tunable per user/node
