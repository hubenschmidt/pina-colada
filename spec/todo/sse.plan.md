# Generic SSE (Server-Sent Events) Implementation

## Goal
Create a reusable SSE infrastructure for real-time updates across the application.
- **First use case**: Crawler run history table (real-time status updates)
- **Future use cases**: Notifications, job status, any polling-based data

## Coding Rules
- **Go**: Guard clauses, early returns, no nested conditionals. CSR pattern. KISS/YAGNI.
- **JS**: Functional programming, guard clauses, no else/switch. ECMAScript 2023.

## Architecture Decision
**Hook-based** (not HOC/wrapper) - matches existing `useWs.jsx` pattern and provides more flexibility.

---

## Backend Files

### 1. SSE Writer Package: `internal/sse/writer.go` (NEW)

Generic SSE helper with:
- `NewWriter(w http.ResponseWriter) *Writer` - sets headers, returns flusher
- `Send(event Event) error` - writes typed event with JSON data
- `SendKeepAlive()` - sends comment to keep connection alive
- `StreamWithKeepAlive(ctx, writer, interval, eventCh)` - main loop with keep-alive

```go
type Event struct {
    ID   string      `json:"id,omitempty"`
    Type string      `json:"event,omitempty"`
    Data interface{} `json:"data"`
}
```

### 2. Pub/Sub Manager: `internal/sse/pubsub.go` (NEW)

In-memory pub/sub for SSE subscriptions:
- `Subscribe(topic string) <-chan Event`
- `Unsubscribe(topic string, ch <-chan Event)`
- `Publish(topic string, event Event)`

Topic format: `crawler:{configID}` for crawler-specific events.

### 3. Controller: `internal/controllers/automation_controller.go`

Add `StreamCrawlerRuns` method:
```go
// GET /automation/crawlers/{id}/runs/stream
func (c *AutomationController) StreamCrawlerRuns(w http.ResponseWriter, r *http.Request)
```

- Creates SSE writer
- Sends initial `PagedResponse` as `init` event
- Subscribes to `crawler:{id}` topic
- Streams updates until client disconnects

### 4. Service: `internal/services/automation_service.go`

Add notification calls in run lifecycle:
- `StartRun()` → publish `run_started` event
- `CompleteRun()` → publish `run_completed` event

### 5. Routes: `internal/routes/router.go`

Add route: `r.Get("/runs/stream", c.Automation.StreamCrawlerRuns)`

---

## Frontend Files

### 1. SSE Hook: `hooks/useSSE.jsx` (NEW)

Generic SSE hook using fetch + ReadableStream (supports auth headers):

```jsx
export const useSSE = (url, { enabled, onEvent, reconnectDelay }) => {
  // Returns: { data, error, isConnecting, isConnected, reconnect }
}
```

Features:
- Auth token via `fetchBearerToken()`
- Auto-reconnect on disconnect
- Parse typed events (init, update, error)
- Clean up on unmount

### 2. Integration: `app/automation/page.jsx`

Update `RunHistoryDataTable` to optionally use SSE:
- When expanded, connect to SSE endpoint
- On `update` events, merge new/updated runs into data
- Show connection status indicator

---

## SSE Event Format

```
event: <type>
data: <JSON>

: keepalive
```

| Event Type | Payload | When |
|------------|---------|------|
| `init` | `PagedResponse` | On connect |
| `run_started` | `AutomationRunLogResponse` | Run begins |
| `run_completed` | `AutomationRunLogResponse` | Run finishes |
| `error` | `{error: string}` | On error |

---

## Implementation Order

1. `internal/sse/writer.go` - SSE writer utilities
2. `internal/sse/pubsub.go` - In-memory pub/sub manager
3. `internal/services/automation_service.go` - Add publish calls
4. `internal/controllers/automation_controller.go` - Add StreamCrawlerRuns
5. `internal/routes/router.go` - Register route
6. `hooks/useSSE.jsx` - Generic frontend hook
7. `app/automation/page.jsx` - Integrate with crawler history

---

## Critical Files
- `modules/agent/internal/sse/writer.go` (new)
- `modules/agent/internal/sse/pubsub.go` (new)
- `modules/agent/internal/controllers/automation_controller.go` (modify)
- `modules/agent/internal/services/automation_service.go` (modify)
- `modules/agent/internal/routes/router.go` (modify)
- `modules/client/hooks/useSSE.jsx` (new)
- `modules/client/app/automation/page.jsx` (modify)
