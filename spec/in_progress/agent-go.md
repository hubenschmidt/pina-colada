# Agent Service: Python to Go Migration Spec

## Overview

Migrate the `modules/agent` Python service to Go, leveraging Go's concurrency model, lower memory footprint, and single-binary deployment.

## Current Architecture (Python)

```
Orchestrator (OpenAI Agents SDK)
    │
    ├─► Triage Router (Claude Haiku) ──► handoff
    │
    └─► Workers (GPT-5.1)
            ├─► General Worker
            ├─► Job Search Worker
            ├─► Cover Letter Worker
            └─► CRM Worker
```

**Key Files:**
- `src/agent/orchestrator.py` - Main orchestrator
- `src/agent/routers/triage.py` - Triage agent
- `src/agent/workers/` - Specialized workers
- `src/agent/tools/` - Tool implementations
- `src/services/` - 40+ domain services

## Framework Options

Two viable approaches for the Go agent implementation:

### Option A: Google ADK for Go (Recommended)

[Google ADK for Go](https://github.com/google/adk-go) (v0.2.0+) - official framework from Google.

**Pros:**
- Built-in A2A protocol for agent handoffs
- MCP Toolbox with 30+ DB integrations
- Dev UI for debugging/testing
- Active development, Google-backed
- Model-agnostic (works with OpenAI/Anthropic)

**Cons:**
- Gemini-optimized (may need workarounds for other models)
- Framework abstraction adds overhead
- Vendor lock-in to Google patterns
- Less control over token optimization

### Option B: Manual Loop (Maximum Control)

Direct SDK calls with hand-rolled state machine - port of Python approach.

**Pros:**
- Full control over token optimization
- No framework overhead or abstractions
- Direct SDK calls = predictable behavior
- Easier to debug and customize
- Model-agnostic by design

**Cons:**
- Build all orchestration yourself
- No built-in observability UI
- More code to maintain
- Reimplement patterns ADK provides free

### Comparison Matrix

| Aspect | Google ADK | Manual Loop |
|--------|------------|-------------|
| Token optimization | Framework-controlled | Full control |
| Multi-agent orchestration | Built-in A2A | Manual dispatch |
| Tool ecosystem | MCP Toolbox (30+ DBs) | Roll your own |
| Observability | Built-in dev UI | Langfuse integration |
| Model flexibility | Gemini-optimized | Any via SDK |
| Startup time | Framework init | Minimal |
| Maintenance | Google-backed | Self-maintained |
| Learning curve | ADK patterns | Go + LLM SDKs |

## Architecture Mapping

### Python → Go Component Map

| Python Component | Go Equivalent |
|------------------|---------------|
| `orchestrator.py` | `internal/agent/orchestrator.go` |
| `routers/triage.py` | `internal/agent/triage.go` |
| `workers/*.py` | `internal/agent/workers/*.go` |
| `tools/*.py` | `internal/tools/*.go` |
| `services/*.py` | `internal/services/*.go` |
| FastAPI routes | Chi/Echo HTTP router |
| WebSocket handler | gorilla/websocket |
| SQLAlchemy ORM | sqlc or GORM |

### Proposed Go Structure

Maintains the same layered architecture: **models → schemas → serializers → repositories → services → controllers → routes**

```
modules/agent-go/
├── cmd/
│   └── agent/
│       └── main.go                  # Entry point, DI wiring
├── internal/
│   ├── agent/                       # Agent orchestration
│   │   ├── orchestrator.go          # Main orchestrator
│   │   ├── triage.go                # Triage agent
│   │   └── workers/
│   │       ├── general.go
│   │       ├── job_search.go
│   │       ├── cover_letter.go
│   │       └── crm.go
│   ├── tools/                       # Agent tools
│   │   ├── document.go
│   │   ├── job.go
│   │   ├── crm.go
│   │   └── registry.go
│   ├── models/                      # Domain entities (ORM structs)
│   │   ├── conversation.go
│   │   ├── job.go
│   │   ├── contact.go
│   │   ├── company.go
│   │   ├── document.go
│   │   └── user.go
│   ├── schemas/                     # Request/Response DTOs
│   │   ├── conversation_schema.go
│   │   ├── job_schema.go
│   │   ├── contact_schema.go
│   │   └── agent_schema.go
│   ├── serializers/                 # Model ↔ Schema transforms
│   │   ├── conversation_serializer.go
│   │   ├── job_serializer.go
│   │   ├── contact_serializer.go
│   │   └── company_serializer.go
│   ├── repositories/                # Data access layer
│   │   ├── conversation_repo.go
│   │   ├── job_repo.go
│   │   ├── contact_repo.go
│   │   ├── company_repo.go
│   │   ├── document_repo.go
│   │   └── queries/                 # sqlc SQL files
│   │       ├── conversations.sql
│   │       ├── jobs.sql
│   │       └── contacts.sql
│   ├── services/                    # Business logic
│   │   ├── conversation_service.go
│   │   ├── job_service.go
│   │   ├── contact_service.go
│   │   ├── document_service.go
│   │   ├── reasoning_service.go
│   │   └── provider_costs_service.go
│   ├── controllers/                 # Request handlers
│   │   ├── agent_controller.go
│   │   ├── conversation_controller.go
│   │   ├── job_controller.go
│   │   ├── contact_controller.go
│   │   └── document_controller.go
│   ├── routes/                      # HTTP route definitions
│   │   ├── agent_routes.go
│   │   ├── conversation_routes.go
│   │   ├── job_routes.go
│   │   └── router.go                # Main router setup
│   ├── middleware/                  # HTTP middleware
│   │   ├── auth.go                  # JWT/session auth
│   │   ├── tenant.go                # Multi-tenant context
│   │   ├── logging.go
│   │   └── cors.go
│   └── config/
│       └── config.go                # Environment config
├── api/
│   └── websocket/
│       └── handler.go               # WebSocket streaming
├── pkg/                             # Shared utilities
│   ├── db/
│   │   └── postgres.go              # Connection pool
│   ├── s3/
│   │   └── client.go                # S3 operations
│   └── langfuse/
│       └── tracer.go                # Observability
├── go.mod
└── go.sum
```

### Layer Responsibilities

| Layer | Python | Go | Notes |
|-------|--------|----|----|
| **Models** | SQLAlchemy ORM | GORM structs | Domain entities |
| **Schemas** | Pydantic | Go structs + validator | Request/response DTOs |
| **Serializers** | Serializer classes | Transform functions | Model ↔ Schema conversion |
| **Repositories** | Raw SQL / ORM | GORM queries | Data access only |
| **Services** | Business logic | Same pattern | Orchestrates repos |
| **Controllers** | FastAPI deps | Chi handlers | HTTP request handling |
| **Routes** | `@router.get()` | `r.Get("/path", ...)` | Route definitions |

### ORM Choice: GORM (Recommended)

[GORM](https://gorm.io/) is the most mature Go ORM, closest to SQLAlchemy patterns.

**Why GORM over sqlc:**
- Full ORM with migrations, associations, hooks
- Familiar to SQLAlchemy users
- Dynamic queries (filters, pagination)
- Soft deletes, timestamps built-in

**GORM Model Example:**

```go
package models

import "gorm.io/gorm"

type Job struct {
    gorm.Model                          // ID, CreatedAt, UpdatedAt, DeletedAt
    TenantID    int       `gorm:"index"`
    Title       string    `gorm:"size:255"`
    Company     string
    Location    string
    Status      string    `gorm:"default:new"`
    AppliedAt   *time.Time
    UserID      int
    User        User      `gorm:"foreignKey:UserID"`
}

type Contact struct {
    gorm.Model
    TenantID    int       `gorm:"index"`
    FirstName   string
    LastName    string
    Email       string    `gorm:"uniqueIndex"`
    CompanyID   *int
    Company     *Company  `gorm:"foreignKey:CompanyID"`
}
```

**Repository Pattern with GORM:**

```go
package repositories

type JobRepository struct {
    db *gorm.DB
}

func (r *JobRepository) FindByTenant(ctx context.Context, tenantID int, filters JobFilters) ([]models.Job, error) {
    var jobs []models.Job
    query := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID)

    if filters.Status != "" {
        query = query.Where("status = ?", filters.Status)
    }
    if filters.Limit > 0 {
        query = query.Limit(filters.Limit)
    }

    return jobs, query.Find(&jobs).Error
}

func (r *JobRepository) Create(ctx context.Context, job *models.Job) error {
    return r.db.WithContext(ctx).Create(job).Error
}
```

**Alternative: sqlc** (if preferring raw SQL)
- Type-safe SQL with generated Go code
- No runtime reflection
- Better for complex queries
- Requires writing SQL manually

### Serializer Pattern

Serializers handle Model ↔ Schema transformations, keeping conversion logic centralized:

```go
package serializers

type JobSerializer struct{}

// ToSchema converts a model to response DTO
func (s *JobSerializer) ToSchema(job *models.Job) *schemas.JobResponse {
    return &schemas.JobResponse{
        ID:        job.ID,
        Title:     job.Title,
        Company:   job.Company,
        Location:  job.Location,
        Status:    job.Status,
        AppliedAt: job.AppliedAt,
        CreatedAt: job.CreatedAt,
    }
}

// ToSchemaList converts multiple models
func (s *JobSerializer) ToSchemaList(jobs []models.Job) []schemas.JobResponse {
    result := make([]schemas.JobResponse, len(jobs))
    for i, job := range jobs {
        result[i] = *s.ToSchema(&job)
    }
    return result
}

// FromSchema converts request DTO to model (for create/update)
func (s *JobSerializer) FromSchema(req *schemas.CreateJobRequest, tenantID int) *models.Job {
    return &models.Job{
        TenantID: tenantID,
        Title:    req.Title,
        Company:  req.Company,
        Location: req.Location,
        Status:   "new",
    }
}
```

---

## Option A: ADK Integration Pattern

### Agent Definition

```go
package agent

import (
    "github.com/google/adk-go/pkg/agent/llmagent"
    "github.com/google/adk-go/pkg/tool"
)

func NewTriageAgent(model llmagent.Model, workers []llmagent.Agent) (*llmagent.Agent, error) {
    return llmagent.New(llmagent.Config{
        Name:        "triage_agent",
        Model:       model,
        Description: "Routes requests to specialized workers",
        Instructions: triagePrompt,
        SubAgents:   workers, // handoff targets
    })
}

func NewJobSearchWorker(model llmagent.Model, tools []tool.Tool) (*llmagent.Agent, error) {
    return llmagent.New(llmagent.Config{
        Name:        "job_search_worker",
        Model:       model,
        Description: "Searches jobs and manages applications",
        Instructions: jobSearchPrompt,
        Tools:       tools,
    })
}
```

### Tool Definition

```go
package tools

import "github.com/google/adk-go/pkg/tool/functiontool"

var JobSearchTool = functiontool.New(
    "job_search",
    "Search for jobs matching criteria",
    func(ctx context.Context, params JobSearchParams) (*JobSearchResult, error) {
        // Implementation
    },
)

type JobSearchParams struct {
    Query    string   `json:"query"`
    Location string   `json:"location,omitempty"`
    Limit    int      `json:"limit,omitempty"`
}
```

### Orchestrator Loop

```go
func (o *Orchestrator) RunStreaming(ctx context.Context, req RunRequest) (<-chan Event, error) {
    events := make(chan Event)

    go func() {
        defer close(events)

        // Load conversation context
        messages, _ := o.convService.LoadMessages(ctx, req.ThreadID, 6)

        // Run agent with ADK
        result, err := o.triageAgent.Run(ctx, llmagent.RunConfig{
            Input:    req.Message,
            Messages: messages,
            Hooks:    o.hooks,
        })

        // Stream events
        for event := range result.Events {
            events <- translateEvent(event)
        }
    }()

    return events, nil
}
```

---

## Option B: Manual Loop Pattern

Direct OpenAI/Anthropic SDK calls with hand-rolled state machine.

### State Definition

```go
package agent

type AgentState struct {
    Messages        []Message      `json:"messages"`
    RouteTo         string         `json:"route_to"`
    SuccessCriteria string         `json:"success_criteria"`
    Feedback        *string        `json:"feedback"`
    Done            bool           `json:"done"`
    FastIntent      *string        `json:"fast_intent"`
    Tokens          TokenUsage     `json:"tokens"`
    Iteration       int            `json:"iteration"`
}

type RequestContext struct {
    ThreadID      string
    TenantID      int
    SchemaContext string
    Tools         map[string][]Tool
}
```

### Orchestrator Loop

```go
func (o *Orchestrator) RunAgent(ctx context.Context, message string, reqCtx RequestContext) (*AgentState, error) {
    state := o.initializeState(message)

    // Load conversation history
    history, _ := o.convService.LoadMessages(ctx, reqCtx.ThreadID, 6)
    state.Messages = append(history, state.Messages...)

    for !state.Done && state.Iteration < 20 {
        update, nextNode, err := o.dispatchNode(ctx, state.RouteTo, state, reqCtx)
        if err != nil {
            return nil, err
        }
        state = o.mergeState(state, update)
        state.RouteTo = nextNode
        state.Iteration++
    }

    return state, nil
}

func (o *Orchestrator) dispatchNode(ctx context.Context, node string, state *AgentState, reqCtx RequestContext) (*StateUpdate, string, error) {
    switch node {
    case "triage":
        return o.triageRouter.Route(ctx, state, reqCtx)
    case "job_search":
        return o.jobSearchWorker.Execute(ctx, state, reqCtx)
    case "cover_letter":
        return o.coverLetterWorker.Execute(ctx, state, reqCtx)
    case "crm":
        return o.crmWorker.Execute(ctx, state, reqCtx)
    case "general":
        return o.generalWorker.Execute(ctx, state, reqCtx)
    default:
        state.Done = true
        return nil, "", nil
    }
}
```

### Worker Implementation

```go
package workers

type JobSearchWorker struct {
    client  *openai.Client
    tools   []Tool
    tracer  *langfuse.Tracer
}

func (w *JobSearchWorker) Execute(ctx context.Context, state *AgentState, reqCtx RequestContext) (*StateUpdate, string, error) {
    // Build messages with tool schemas
    messages := w.buildMessages(state, reqCtx)

    // Call OpenAI directly
    resp, err := w.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
        Model:    openai.F("gpt-4.1"),
        Messages: openai.F(messages),
        Tools:    openai.F(w.toolSchemas()),
    })
    if err != nil {
        return nil, "", err
    }

    // Track tokens
    w.tracer.LogUsage(resp.Usage)

    // Handle tool calls
    if len(resp.Choices[0].Message.ToolCalls) > 0 {
        results := w.executeTools(ctx, resp.Choices[0].Message.ToolCalls)
        return &StateUpdate{Messages: results}, "job_search", nil // loop back
    }

    // Done - check if needs evaluation
    if w.needsEvaluator(state) {
        return &StateUpdate{Messages: []Message{resp.Choices[0].Message}}, "evaluator", nil
    }

    return &StateUpdate{Messages: []Message{resp.Choices[0].Message}, Done: true}, "", nil
}
```

### Tool Execution

```go
package tools

type ToolRegistry struct {
    tools map[string]ToolFunc
}

func (r *ToolRegistry) Execute(ctx context.Context, name string, args json.RawMessage) (string, error) {
    fn, ok := r.tools[name]
    if !ok {
        return "", fmt.Errorf("unknown tool: %s", name)
    }
    return fn(ctx, args)
}

// Tool definitions
func (r *ToolRegistry) Register() {
    r.tools["job_search"] = func(ctx context.Context, args json.RawMessage) (string, error) {
        var params JobSearchParams
        json.Unmarshal(args, &params)
        return r.jobService.Search(ctx, params)
    }

    r.tools["crm_lookup"] = func(ctx context.Context, args json.RawMessage) (string, error) {
        var params CRMLookupParams
        json.Unmarshal(args, &params)
        return r.crmService.Lookup(ctx, params)
    }
}
```

---

## Dependencies

### Option A (ADK)

```go
require (
    github.com/google/adk-go v0.2.0
    github.com/openai/openai-go v1.0.0
    github.com/anthropics/anthropic-sdk-go v1.0.0
    gorm.io/gorm v1.25.0
    gorm.io/driver/postgres v1.5.0
    github.com/gorilla/websocket v1.5.0
    github.com/go-chi/chi/v5 v5.0.0
    github.com/aws/aws-sdk-go-v2 v1.24.0
    github.com/langfuse/langfuse-go v1.0.0
)
```

### Option B (Manual Loop)

```go
require (
    github.com/openai/openai-go v1.0.0
    github.com/anthropics/anthropic-sdk-go v1.0.0
    gorm.io/gorm v1.25.0
    gorm.io/driver/postgres v1.5.0
    github.com/gorilla/websocket v1.5.0
    github.com/go-chi/chi/v5 v5.0.0
    github.com/aws/aws-sdk-go-v2 v1.24.0
    github.com/langfuse/langfuse-go v1.0.0
)
```

## Migration Strategy

### Phase 1: Foundation
1. Set up Go module structure
2. Implement config loading (env vars)
3. Database layer with sqlc (port key queries)
4. S3 client for document storage

### Phase 2: Tools
1. Port `document_tools.go` (read/list documents)
2. Port `crm_tools.go` (lookup/count/list)
3. Port `job_tools.go` (search/status)
4. Create tool registry

### Phase 3: Agents
1. Implement triage agent with ADK
2. Port workers (general, job_search, cover_letter, crm)
3. Wire up handoff mechanism
4. Implement conversation context loading

### Phase 4: API Layer
1. HTTP routes (Chi router)
2. WebSocket streaming handler
3. Health/metrics endpoints

### Phase 5: Integration
1. Langfuse observability integration
2. Token tracking and cost attribution
3. End-to-end testing
4. Docker containerization

### Phase 6: Cutover
1. Deploy alongside Python service
2. Feature flag traffic routing
3. Monitor and validate
4. Deprecate Python service

## Trade-offs

### Advantages
- **Performance**: Goroutines for parallel tool execution, lower latency
- **Memory**: ~10x lower than Python runtime
- **Deployment**: Single static binary, faster cold starts
- **Type Safety**: Compile-time verification
- **ADK Features**: Built-in A2A, MCP toolbox, dev UI

### Challenges
- **Rewrite Cost**: All workers, tools, services need porting
- **Lost Integrations**: langchain-community tools (Wikipedia, etc.)
- **Library Gaps**: Some Python libs lack Go equivalents
- **Team Familiarity**: May require Go training

## Docker Configuration

### Dockerfile

```dockerfile
# modules/agent-go/Dockerfile
FROM golang:1.23-alpine AS builder
WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git ca-certificates

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -o /agent-go ./cmd/agent

# Runtime image
FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /agent-go /agent-go

EXPOSE 8080
CMD ["/agent-go"]
```

### docker-compose.yml Entry

```yaml
# Add to docker-compose.yml
agent-go:
  build:
    context: ./modules/agent-go
    dockerfile: Dockerfile
  env_file:
    - ./modules/agent-go/.env
  environment:
    - PORT=8080
    - ENV=development
    - DATABASE_URL=postgres://${POSTGRES_USER:-postgres}:${POSTGRES_PASSWORD:-postgres}@app-postgres:5432/${COMPOSE_PROJECT_NAME:-pina_colada}
    - LANGFUSE_HOST=http://langfuse:3000
  ports:
    - "8080:8080"
  depends_on:
    app-postgres:
      condition: service_healthy
  logging:
    driver: "json-file"
    options:
      max-size: "10m"
      max-file: "3"
```

**Port Assignment:**
- Python agent: `8000` (current)
- Go agent: `8080` (new)

---

## Frontend API Switching

Add a dropdown in the frontend settings/dev panel to switch between Python and Go APIs:

### Implementation

```javascript
// Frontend: API base URL switching
const API_BACKENDS = {
  python: 'http://localhost:8000',
  go: 'http://localhost:8080',
};

// Store preference in localStorage
const getBackend = () =>
  localStorage.getItem('api_backend') || 'python';

const setBackend = (backend) =>
  localStorage.setItem('api_backend', backend);

// Use in API client
const apiBaseUrl = API_BACKENDS[getBackend()];
```

### UI Component

Add a simple toggle in the settings or dev toolbar:
- Label: "API Backend"
- Options: "Python (8000)" | "Go (8080)"
- Persists to localStorage
- Only visible in development mode

---

## Authentication: Auth0 Integration

The Go service must implement the same Auth0 JWT verification as Python.

### Required Environment Variables

```env
# modules/agent-go/.env
AUTH0_DOMAIN=your-tenant.auth0.com
AUTH0_AUDIENCE=https://api.pinacolada.co
AUTH0_NAMESPACE=https://pinacolada.co
```

### Auth Middleware Implementation

```go
package middleware

import (
	"context"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/golang-jwt/jwt/v5"
	"github.com/lestrrat-go/jwx/v2/jwk"
)

var (
	jwksCache jwk.Set
	jwksMu    sync.RWMutex
)

type Claims struct {
	jwt.RegisteredClaims
	Email    string `json:"email"`
	TenantID *int64 `json:"https://pinacolada.co/tenant_id"`
}

type AuthContext struct {
	UserID   int64
	TenantID int64
	Email    string
	Auth0Sub string
}

func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "Missing or invalid Authorization header", http.StatusUnauthorized)
			return
		}
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Verify JWT
		claims, err := verifyToken(tokenString)
		if err != nil {
			http.Error(w, "Invalid token: "+err.Error(), http.StatusUnauthorized)
			return
		}

		// Get or create user (same logic as Python)
		user, err := getOrCreateUser(r.Context(), claims.Subject, claims.Email)
		if err != nil {
			http.Error(w, "Failed to resolve user", http.StatusInternalServerError)
			return
		}

		// Resolve tenant ID (header > claim > user default)
		tenantID := resolveTenantID(r, claims, user)

		// Add auth context to request
		authCtx := AuthContext{
			UserID:   user.ID,
			TenantID: tenantID,
			Email:    claims.Email,
			Auth0Sub: claims.Subject,
		}
		ctx := context.WithValue(r.Context(), "auth", authCtx)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func verifyToken(tokenString string) (*Claims, error) {
	domain := os.Getenv("AUTH0_DOMAIN")
	audience := os.Getenv("AUTH0_AUDIENCE")

	// Fetch JWKS (cached)
	keySet, err := getJWKS(domain)
	if err != nil {
		return nil, err
	}

	// Parse and verify token
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, fmt.Errorf("missing kid header")
		}
		key, ok := keySet.LookupKeyID(kid)
		if !ok {
			return nil, fmt.Errorf("key not found")
		}
		var rawKey interface{}
		key.Raw(&rawKey)
		return rawKey, nil
	}, jwt.WithAudience(audience), jwt.WithIssuer("https://"+domain+"/"))

	if err != nil {
		return nil, err
	}

	return token.Claims.(*Claims), nil
}

func resolveTenantID(r *http.Request, claims *Claims, user *models.User) int64 {
	// 1. Check X-Tenant-Id header
	if tid := r.Header.Get("X-Tenant-Id"); tid != "" {
		if id, err := strconv.ParseInt(tid, 10, 64); err == nil {
			return id
		}
	}
	// 2. Check JWT claim
	if claims.TenantID != nil {
		return *claims.TenantID
	}
	// 3. Fall back to user's default tenant
	if user.TenantID != nil {
		return *user.TenantID
	}
	return 0
}
```

### Key Points

1. **Same JWT verification**: RS256 with Auth0 JWKS
2. **Same claim extraction**: `sub`, `email`, namespaced `tenant_id`
3. **Same tenant resolution**: Header → Claim → User default
4. **User get-or-create**: Match Python behavior for first-time logins

### Additional Auth Dependencies

```go
require (
    github.com/golang-jwt/jwt/v5 v5.2.0
    github.com/lestrrat-go/jwx/v2 v2.0.21
)
```

---

## Open Questions

1. **Model Routing**: ADK is Gemini-optimized; need to verify OpenAI/Anthropic integration patterns
2. **Langfuse Go SDK**: Verify feature parity with Python SDK
3. **MCP Toolbox**: Evaluate if it covers our Supabase/Postgres needs
4. **A2A Protocol**: Assess if beneficial for our worker handoff pattern

## References

- [Google ADK Go Repo](https://github.com/google/adk-go)
- [ADK Documentation](https://google.github.io/adk-docs/get-started/go/)
- [ADK Multi-Agent Patterns](https://google.github.io/adk-docs/)
- [OpenAI Go SDK](https://github.com/openai/openai-go)
