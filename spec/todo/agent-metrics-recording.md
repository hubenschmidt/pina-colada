# Plan: Agent Metrics Recording & Visualization

## Goal
Build a "record mode" for developers to capture agent runtime metrics and visualize cost/performance comparisons across different model configurations.

## Requirements Summary
- **Record button** next to gear icon on /chat (developer only)
- **Manual start/stop** recording sessions
- **Capture metrics**: tokens (in/out), time, cost estimate, model, presets, all runtime parameters
- **New /metrics page** (developer only, in sidebar) with line charts
- **Two comparison modes**: same prompt with different models + historical trends

---

## Coding Rules (MUST FOLLOW)

### Go (`~/commands/go-code-rules.md`)
- **Guard clauses**: Early returns, no nested if/else, no switch-case, no break/continue
- **CSR pattern**: Routes → Controllers → Services → Repositories
- **Imports**: grouped (stdlib, external, internal with blank lines)
- **GORM**: TableName() returns PascalCase (e.g., `"Agent_Recording_Session"`)
- **JSON tags**: snake_case

### CSR Pattern (`~/commands/csr-code-rules.md`)
- **Services never import models** - use repository DTOs/input structs
- **Repositories define**: Input structs (`SessionCreateInput`), DTOs (`SessionWithMetricsDTO`)
- **Controllers**: Parse request, call service, write response
- **Sentinel errors**: Define in services (e.g., `ErrSessionNotFound`)

### JavaScript (`~/commands/js-code-rules.md`)
- **Guard clauses**: No nested conditionals, else, switch/case
- **Functional programming**: Avoid OOP, prefer pure functions
- **React**: Avoid useRef
- **ECMAScript 2023**

---

## Part 1: Database Schema

### Migration: `067_agent_metrics_recording.sql`

```sql
-- Recording sessions (manual start/stop)
CREATE TABLE "Agent_Recording_Session" (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES "User"(id) ON DELETE CASCADE,
    tenant_id BIGINT NOT NULL REFERENCES "Tenant"(id) ON DELETE CASCADE,
    name VARCHAR(255),
    started_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    ended_at TIMESTAMP WITH TIME ZONE,
    config_snapshot JSONB,  -- cost tier, preset, node configs at start
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Individual metric data points (one per agent turn)
CREATE TABLE "Agent_Metric" (
    id BIGSERIAL PRIMARY KEY,
    session_id BIGINT NOT NULL REFERENCES "Agent_Recording_Session"(id) ON DELETE CASCADE,
    conversation_id BIGINT REFERENCES "Conversation"(id) ON DELETE SET NULL,
    thread_id VARCHAR(255),

    -- Timing
    started_at TIMESTAMP WITH TIME ZONE NOT NULL,
    ended_at TIMESTAMP WITH TIME ZONE NOT NULL,
    duration_ms INT NOT NULL,

    -- Tokens
    input_tokens INT NOT NULL,
    output_tokens INT NOT NULL,
    total_tokens INT NOT NULL,

    -- Cost (calculated from model pricing)
    estimated_cost_usd DECIMAL(10, 6),

    -- Model/Config used
    model VARCHAR(100) NOT NULL,
    provider VARCHAR(50) NOT NULL,
    node_name VARCHAR(100),  -- which worker/agent

    -- Full config snapshot for this turn
    config_snapshot JSONB,  -- temperature, max_tokens, etc.

    -- User input (for "same prompt different model" comparisons)
    user_message TEXT,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_agent_metric_session ON "Agent_Metric"(session_id);
CREATE INDEX idx_agent_metric_started ON "Agent_Metric"(started_at);
```

---

## Part 2: Backend - Cost Calculation

### File: `internal/agent/utils/cost.go` (NEW)

Model pricing table (per 1M tokens):

| Model | Input $/1M | Output $/1M |
|-------|------------|-------------|
| gpt-5-nano | $0.10 | $0.40 |
| gpt-5-mini | $0.15 | $0.60 |
| gpt-5 | $2.50 | $10.00 |
| gpt-5.1 | $3.00 | $12.00 |
| gpt-5.2 | $5.00 | $20.00 |
| claude-haiku-4-5 | $0.25 | $1.25 |
| claude-sonnet-4-5 | $3.00 | $15.00 |
| claude-opus-4-5 | $15.00 | $75.00 |

```go
func CalculateCost(model string, inputTokens, outputTokens int) float64
```

---

## Part 3: Backend - Recording API

### Repository: `internal/repositories/metric_repository.go` (NEW)

```go
type MetricRepository struct { db *gorm.DB }

func (r *MetricRepository) CreateSession(userID, tenantID int64, configSnapshot json.RawMessage) (*RecordingSession, error)
func (r *MetricRepository) EndSession(sessionID int64) error
func (r *MetricRepository) GetActiveSession(userID int64) (*RecordingSession, error)
func (r *MetricRepository) RecordMetric(metric AgentMetric) error
func (r *MetricRepository) GetSessionMetrics(sessionID int64) ([]AgentMetric, error)
func (r *MetricRepository) ListSessions(userID int64, limit int) ([]RecordingSession, error)
```

### Service: `internal/services/metric_service.go` (NEW)

```go
type MetricService struct { repo *MetricRepository; configService *AgentConfigService }

func (s *MetricService) StartRecording(userID, tenantID int64) (*RecordingSession, error)
func (s *MetricService) StopRecording(userID int64) (*RecordingSession, error)
func (s *MetricService) GetActiveSession(userID int64) (*RecordingSession, error)
func (s *MetricService) RecordTurn(sessionID int64, metric TurnMetric) error
func (s *MetricService) GetSessionData(sessionID int64) (*SessionWithMetrics, error)
func (s *MetricService) ListSessions(userID int64, limit int) ([]RecordingSession, error)
func (s *MetricService) GetComparisonData(sessionIDs []int64) (*ComparisonData, error)
```

### Controller: `internal/controllers/metric_controller.go` (NEW)

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/metrics/recording/start` | Start recording session |
| POST | `/metrics/recording/stop` | Stop active session |
| GET | `/metrics/recording/active` | Get active session (if any) |
| GET | `/metrics/sessions` | List recording sessions |
| GET | `/metrics/sessions/{id}` | Get session with all metrics |
| GET | `/metrics/compare` | Compare multiple sessions |

### Routes: Add to `internal/routes/router.go`

```go
r.Route("/metrics", func(r chi.Router) {
    r.Use(middleware.RequireRole("developer"))
    r.Post("/recording/start", c.Metric.StartRecording)
    r.Post("/recording/stop", c.Metric.StopRecording)
    r.Get("/recording/active", c.Metric.GetActiveSession)
    r.Get("/sessions", c.Metric.ListSessions)
    r.Get("/sessions/{id}", c.Metric.GetSession)
    r.Get("/compare", c.Metric.Compare)
})
```

---

## Part 4: Backend - Orchestrator Integration

### Modify: `internal/agent/orchestrator.go`

1. Add `metricService *services.MetricService` to Orchestrator
2. In `RunWithStreaming`, after completion:
   - Check if user has active recording session
   - If yes, call `metricService.RecordTurn()` with all metrics

```go
// After agent completes, record metrics if session active
if activeSession := o.getActiveRecordingSession(userID); activeSession != nil {
    o.recordMetric(activeSession.ID, TurnMetric{
        ThreadID:     req.SessionID,
        StartedAt:    startTime,
        EndedAt:      time.Now(),
        DurationMs:   elapsed,
        InputTokens:  turnTokens.Input,
        OutputTokens: turnTokens.Output,
        Model:        triageModel,
        Provider:     triageProvider,
        NodeName:     currentAgent,
        ConfigSnapshot: configSnapshot,
        UserMessage:  req.Message,
    })
}
```

---

## Part 5: Frontend - Record Button

### Modify: `components/Chat/AgentConfigMenu.jsx`

Add record button next to gear icon (inside DeveloperFeature wrapper):

```jsx
<DeveloperFeature>
  <button
    className={`${styles.recordButton} ${isRecording ? styles.recording : ''}`}
    onClick={toggleRecording}
    title={isRecording ? "Stop Recording" : "Start Recording"}
  >
    {isRecording ? <Square size={16} /> : <Circle size={16} />}
  </button>
  <button className={styles.settingsButton} onClick={() => setIsOpen(!isOpen)}>
    <Settings size={20} />
  </button>
</DeveloperFeature>
```

### Modify: `components/Chat/AgentConfigMenu.module.css`

```css
.recordButton {
  /* Red circle/square icon */
  color: #dc3545;
}
.recordButton.recording {
  animation: pulse 1.5s infinite;
}
```

### New API functions in `api/index.js`:

```javascript
export const startRecording = () => apiPost("/metrics/recording/start");
export const stopRecording = () => apiPost("/metrics/recording/stop");
export const getActiveRecording = () => apiGet("/metrics/recording/active");
export const getRecordingSessions = () => apiGet("/metrics/sessions");
export const getRecordingSession = (id) => apiGet(`/metrics/sessions/${id}`);
export const compareRecordingSessions = (ids) => apiGet(`/metrics/compare?ids=${ids.join(",")}`);
```

---

## Part 6: Frontend - Metrics Page

### New Page: `app/metrics/page.jsx`

- Developer-only (redirect if not developer role)
- Session list with date, name, duration, total cost
- Click session to view charts

### New Component: `components/Metrics/MetricsCharts.jsx`

Use a charting library (recharts or chart.js):

**Charts to display:**
1. **Token Usage Over Time** - Line chart showing input/output tokens per turn
2. **Cost Breakdown** - Bar chart by model/agent
3. **Latency Distribution** - Line chart of response times
4. **Model Comparison** - Side-by-side when comparing sessions

### Modify: Sidebar Navigation

Add "Metrics" link (developer only) in sidebar/navigation component.

---

## Files to Create/Modify

| Action | File |
|--------|------|
| CREATE | `modules/agent/migrations/067_agent_metrics_recording.sql` |
| CREATE | `modules/agent-go/internal/agent/utils/cost.go` |
| CREATE | `modules/agent-go/internal/repositories/metric_repository.go` |
| CREATE | `modules/agent-go/internal/services/metric_service.go` |
| CREATE | `modules/agent-go/internal/controllers/metric_controller.go` |
| MODIFY | `modules/agent-go/internal/routes/router.go` |
| MODIFY | `modules/agent-go/internal/agent/orchestrator.go` |
| MODIFY | `modules/agent-go/cmd/agent/main.go` |
| MODIFY | `modules/client/api/index.js` |
| MODIFY | `modules/client/components/Chat/AgentConfigMenu.jsx` |
| MODIFY | `modules/client/components/Chat/AgentConfigMenu.module.css` |
| CREATE | `modules/client/app/metrics/page.jsx` |
| CREATE | `modules/client/components/Metrics/MetricsCharts.jsx` |
| CREATE | `modules/client/components/Metrics/MetricsCharts.module.css` |
| MODIFY | Sidebar component (add Metrics link) |

---

## Implementation Order

### Phase 1: Backend Foundation
1. Create migration 067
2. Create cost.go with pricing table
3. Create metric_repository.go
4. Create metric_service.go
5. Create metric_controller.go
6. Add routes
7. Wire up in main.go

### Phase 2: Orchestrator Integration
8. Modify orchestrator to check for active session
9. Record metrics on turn completion

### Phase 3: Frontend - Recording
10. Add API functions
11. Add record button to AgentConfigMenu
12. Style the record button

### Phase 4: Frontend - Visualization
13. Create /metrics page
14. Create MetricsCharts component
15. Add sidebar link
16. Install charting library (recharts)

### Phase 5: Polish
17. Session naming/management
18. Export functionality (CSV/JSON)
19. Historical trend charts
