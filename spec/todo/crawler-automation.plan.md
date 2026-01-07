# Automated Job Lead Sourcing Bot

## Goal
Create a scheduled automation bot that searches for job leads every 30 minutes, creates proposals for approval, and sends a daily digest email at 6am EST.

## Coding Rules
- **Go**: Guard clauses, early returns, no nested conditionals. CSR pattern. KISS/YAGNI.
- **JS**: Functional programming, guard clauses, no else/switch. ECMAScript 2023.
- **CSR**: Routes→Controllers→Services→Repositories. Services never import models.

## Key Features
- Per-user automation configuration
- Configurable: model, interval, leads per run, compilation target, system prompt
- Target entity (Individual) + source documents (resume) selection
- Search keywords + resume-driven queries
- Daily digest email to multiple recipients
- Run history tracking

---

## Database Schema

### Migration: `xxx_automation_config.up.sql`

```sql
CREATE TABLE "Automation_Config" (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL REFERENCES "Tenant"(id),
    user_id BIGINT NOT NULL REFERENCES "User"(id),
    enabled BOOLEAN NOT NULL DEFAULT false,

    -- Scheduling
    interval_minutes INT NOT NULL DEFAULT 30,
    last_run_at TIMESTAMPTZ,
    next_run_at TIMESTAMPTZ,
    run_count INT NOT NULL DEFAULT 0,

    -- Model: Uses user's existing Agent_Node_Config (configured via /chat settings)
    -- No separate model/provider columns - shares chat config

    -- Search Configuration
    leads_per_run INT NOT NULL DEFAULT 10,
    compilation_target INT NOT NULL DEFAULT 100,
    system_prompt TEXT,
    search_keywords JSONB,
    ats_mode BOOLEAN NOT NULL DEFAULT true,
    time_filter VARCHAR(20) DEFAULT 'week',

    -- Target Entity
    target_individual_id BIGINT REFERENCES "Individual"(id),
    source_document_ids JSONB,

    -- Digest Configuration
    digest_enabled BOOLEAN NOT NULL DEFAULT true,
    digest_emails TEXT,
    last_digest_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE(tenant_id, user_id)
);

CREATE TABLE "Automation_Run_Log" (
    id BIGSERIAL PRIMARY KEY,
    automation_config_id BIGINT NOT NULL REFERENCES "Automation_Config"(id) ON DELETE CASCADE,
    started_at TIMESTAMPTZ NOT NULL,
    completed_at TIMESTAMPTZ,
    status VARCHAR(20) NOT NULL DEFAULT 'running',
    leads_found INT NOT NULL DEFAULT 0,
    proposals_created INT NOT NULL DEFAULT 0,
    error_message TEXT,
    search_query TEXT
);

CREATE INDEX idx_automation_config_enabled ON "Automation_Config"(enabled, next_run_at);
CREATE INDEX idx_automation_run_log_config ON "Automation_Run_Log"(automation_config_id, started_at DESC);
```

---

## Backend Files

### 1. Model: `internal/models/automation_config.go`
- `AutomationConfig` struct with GORM tags
- `AutomationRunLog` struct

### 2. Repository: `internal/repositories/automation_repository.go`
Key methods:
- `GetUserConfig(userID) → *AutomationConfigDTO`
- `UpsertConfig(userID, input) → error`
- `GetDueAutomations(now) → []AutomationConfigDTO`
- `UpdateLastRun(configID, nextRun)`
- `GetDigestDueConfigs(now) → []AutomationConfigDTO`
- `CreateRunLog() / CompleteRunLog()`

### 3. Service: `internal/services/automation_service.go`
- `GetConfig(userID)`
- `UpdateConfig(userID, input)`
- `GetRunHistory(userID, limit)`

### 4. Controller: `internal/controllers/automation_controller.go`
Endpoints:
| Method | Path | Description |
|--------|------|-------------|
| GET | `/automation/config` | Get user config |
| PUT | `/automation/config` | Update config |
| GET | `/automation/config/models` | List models |
| POST | `/automation/config/toggle` | Enable/disable |
| GET | `/automation/runs` | Run history |
| POST | `/automation/test-run` | Manual test |

### 5. Scheduler: `internal/scheduler/scheduler.go`
- Uses `github.com/robfig/cron/v3`
- Cron entries:
  - `* * * * *` → Check for due automations
  - `0 11 * * *` → Daily digest (11:00 UTC = 6am EST)

### 6. Worker: `internal/scheduler/automation_worker.go`
`ExecuteAutomation(ctx, config)`:
1. Load target individual's documents
2. Build search query (keywords + resume)
3. Execute `SerperTools.JobSearchCtx`
4. Create proposals via `CRMTools.ProposeBatchLeadCreateCtx`
5. Mark proposals with `source: "automation"`
6. Update run log

### 7. Digest: `internal/scheduler/digest_service.go`
- Query proposals from last 24h where source="automation"
- Group by status (pending/approved/rejected)
- Send formatted email to configured addresses

### 8. Routes: `internal/routes/router.go`
Add `/automation` route group

### 9. Main: `cmd/agent/main.go`
Initialize and start scheduler after services

---

## Frontend Files

### 1. Sidebar: `components/Sidebar/Sidebar.jsx`
Add after Data, before Reports:
```jsx
<button onClick={() => router.push("/automation")}>
  <Bot className="h-4 w-4 text-lime-600" />
  Automation
</button>
```

### 2. Page: `app/automation/page.jsx`
Sections:
- Enable/Disable toggle with status
- **Model Config**: Import and reuse `AgentConfigMenu` component from `/components/Chat/AgentConfigMenu.jsx`
  - This shares the same model/provider selection, cost tiers, and presets
  - Automation will use the user's existing agent config
- Interval, leads per run, compilation target inputs
- Target Individual selector (searchable)
- Document selector (multi-select)
- Keywords input (tags)
- System prompt textarea (automation-specific, separate from agent config)
- Digest toggle + emails input
- Run history table
- Test Run button

### 3. API: `api/index.js`
```javascript
getAutomationConfig()
updateAutomationConfig(config)
toggleAutomation(enabled)
getAutomationModels()
getAutomationRuns(limit)
triggerTestRun()
```

---

## Implementation Order

1. Database migration
2. Go models
3. Repository
4. Service
5. Controller + routes
6. Scheduler + worker + digest
7. Main.go integration
8. Frontend API functions
9. Automation page
10. Sidebar nav item

---

## Key Integration Points

- **Proposal source**: Add `source: "automation"` to proposal payload
- **Scheduler startup**: Init in main.go, defer Stop()
- **Dependency**: `go get github.com/robfig/cron/v3`
