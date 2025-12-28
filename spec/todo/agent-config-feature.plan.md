# Agent Node Model Configuration Feature

## Overview
Allow developers to configure which LLM model each agent node uses via a UI on the /chat page.

## Requirements
- Config menu (gear icon) in chat header, visible only to users with "developer" role
- Per-user configuration stored in database
- Config fetched on page load, NOT on every chat message
- Support all 6 agent nodes with their respective model providers

## Agent Nodes

| Node | Default Model | Provider | Location |
|------|---------------|----------|----------|
| Triage Orchestrator | gpt-5.2 | OpenAI | orchestrator.go:55 |
| Job Search Worker | gpt-5.2 | OpenAI | workers/job_search_worker.go |
| CRM Worker | gpt-5.2 | OpenAI | workers/crm_worker.go |
| General Worker | gpt-5.2 | OpenAI | workers/general_worker.go |
| Evaluator | claude-sonnet-4-5 | Anthropic | evaluator.go:83 |
| Title Generator | claude-haiku-4-5 | Anthropic | orchestrator.go:715 |

---

## Implementation Steps

### 1. Database Migration
**File:** `migrations/061_agent_node_config.sql`

```sql
CREATE TABLE IF NOT EXISTS "Agent_Node_Config" (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES "User"(id) ON DELETE CASCADE,
    node_name TEXT NOT NULL,
    model TEXT NOT NULL,
    provider TEXT NOT NULL DEFAULT 'openai',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT agent_node_config_user_node_unique UNIQUE (user_id, node_name)
);
CREATE INDEX IF NOT EXISTS ix_agent_node_config_user_id ON "Agent_Node_Config"(user_id);
```

### 2. Backend Model
**File:** `internal/models/AgentNodeConfig.go`

- Define `AgentNodeConfig` struct with fields: ID, UserID, NodeName, Model, Provider, timestamps
- Define node name constants: `NodeTriageOrchestrator`, `NodeJobSearchWorker`, etc.
- Define `AvailableModels` map by provider
- Define `DefaultModels` map with defaults per node

### 3. Backend Repository
**File:** `internal/repositories/agent_config_repository.go`

Methods:
- `GetUserConfig(userID int64) ([]AgentNodeConfig, error)`
- `GetNodeConfig(userID int64, nodeName string) (*AgentNodeConfig, error)`
- `UpsertNodeConfig(userID int64, nodeName, model, provider string) error`
- `DeleteNodeConfig(userID int64, nodeName string) error`

### 4. Backend Service
**File:** `internal/services/agent_config_service.go`

Methods:
- `GetAgentConfig(userID int64) (*AgentConfigResponse, error)` - returns all 6 nodes with current/default
- `UpdateNodeConfig(userID int64, nodeName, model string) error` - validate and save
- `ResetNodeConfig(userID int64, nodeName string) error` - delete override
- `GetAvailableModels() map[string][]string`

### 5. Backend Controller
**File:** `internal/controllers/agent_config_controller.go`

Endpoints (all require developer role):
- `GET /agent/config` - get all node configs for user
- `PUT /agent/config/{node_name}` - update node model
- `DELETE /agent/config/{node_name}` - reset to default
- `GET /agent/config/models` - get available models

### 6. Routes
**Modify:** `internal/routes/router.go`

- Add `AgentConfig *controllers.AgentConfigController` to Controllers struct
- Add route group `/agent/config` with the 4 endpoints

### 7. Agent Config Cache
**File:** `internal/agent/config_cache.go`

```go
type AgentConfigCache struct {
    configs map[int64]map[string]string // userID -> nodeName -> model
    mu      sync.RWMutex
    service *services.AgentConfigService
}
```

Methods:
- `GetModel(userID int64, nodeName string) string` - returns model or default
- `RefreshUser(userID int64) error` - reload from DB
- `Invalidate(userID int64)` - clear cache for user

### 8. Orchestrator Integration
**Modify:** `internal/agent/orchestrator.go`

Key changes:
- Add `configCache *AgentConfigCache` field
- Modify `Run` and `RunWithStreaming` to:
  1. Get userID from request
  2. Fetch models from cache for each node
  3. Create tenant-specific workers with correct models
- Cache workers per-user to avoid recreation every message

### 9. Evaluator Integration
**Modify:** `internal/agent/evaluator.go`

- Change `NewEvaluator` to accept model parameter
- Remove hardcoded `anthropic.ModelClaudeSonnet4_5`
- Update `runEvaluation` in orchestrator to pass user's evaluator model

### 10. Title Generator Integration
**Modify:** `internal/agent/orchestrator.go` (callAnthropicAPI function)

- Accept model as parameter instead of hardcoded `claude-haiku-4-5-20251001`
- Update `generateConversationTitle` to pass user's title generator model

### 11. Main.go Wiring
**Modify:** `cmd/agent/main.go`

- Initialize AgentConfigRepository
- Initialize AgentConfigService
- Initialize AgentConfigController
- Pass config service to orchestrator
- Add controller to routes

### 12. Frontend API Client
**Modify:** `modules/client/api/index.js`

```javascript
export const getAgentConfig = () => apiGet("/agent/config");
export const getAvailableModels = () => apiGet("/agent/config/models");
export const updateAgentNodeConfig = (nodeName, model) =>
  apiPut(`/agent/config/${nodeName}`, { model });
export const resetAgentNodeConfig = (nodeName) =>
  apiDelete(`/agent/config/${nodeName}`);
```

### 13. Frontend Component
**File:** `modules/client/components/AgentConfigMenu/AgentConfigMenu.jsx`

Features:
- Gear icon trigger (ActionIcon with Settings icon)
- Popover/Dropdown showing all 6 nodes
- Each node: name, current model, Select dropdown
- Save changes on selection
- Reset button per node
- Only renders if `has_developer_access` is true
- Fetches config on open (not on page load for non-developers)

### 14. Chat Page Integration
**Modify:** `modules/client/components/Chat/Chat.jsx`

- Import AgentConfigMenu component
- Add to header area near existing tools dropdown (evaluator toggle)
- Position: gear icon in the control bar

---

## Files to Create
1. `migrations/061_agent_node_config.sql`
2. `internal/models/AgentNodeConfig.go`
3. `internal/repositories/agent_config_repository.go`
4. `internal/services/agent_config_service.go`
5. `internal/controllers/agent_config_controller.go`
6. `internal/agent/config_cache.go`
7. `modules/client/components/AgentConfigMenu/AgentConfigMenu.jsx`

## Files to Modify
1. `internal/routes/router.go` - add routes
2. `internal/agent/orchestrator.go` - config cache, dynamic workers
3. `internal/agent/evaluator.go` - parameterize model
4. `cmd/agent/main.go` - wire components
5. `modules/client/api/index.js` - add API functions
6. `modules/client/components/Chat/Chat.jsx` - add config menu

---

## Runtime Flow

```
1. Developer opens /chat page
2. Chat component loads
3. AgentConfigMenu checks developer access
4. If developer: fetches config from GET /agent/config
5. User changes model for a node
6. PUT /agent/config/{node_name} called
7. Backend updates DB, invalidates cache
8. Next chat message uses new model for that node
```

## Performance Considerations
- Config cached in memory, not fetched per-message
- Cache invalidated only on config CRUD operations
- Workers created per-user on first message, cached for session
- No additional latency on chat messages
