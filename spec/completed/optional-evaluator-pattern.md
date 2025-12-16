# Optional Evaluator Pattern for Agent-Go

## Overview
Add an optional evaluator that runs after agent responses, controlled by a UI checkbox. Evaluator validates response quality, retries on failure (score < 60), and always shows result to user.

## Files to Modify

### Backend (agent-go)
1. `modules/agent-go/internal/controllers/websocket_controller.go` - Add `UseEvaluator` to WSMessage
2. `modules/agent-go/internal/agent/orchestrator.go` - Integrate evaluator into RunWithStreaming
3. `modules/agent-go/internal/agent/evaluator.go` - Fix pass threshold to 60

### Frontend (client)
4. `modules/client/hooks/useWs.jsx` - Add `useEvaluator` state and pass in message
5. `modules/client/components/Chat/Chat.jsx` - Add checkbox UI in Tools dropdown

---

## Implementation Steps

### Step 1: Update WSMessage struct
**File:** `websocket_controller.go`
```go
type WSMessage struct {
    UUID         string `json:"uuid"`
    Message      string `json:"message"`
    Init         bool   `json:"init,omitempty"`
    UserID       int64  `json:"user_id,omitempty"`
    TenantID     int64  `json:"tenant_id,omitempty"`
    Type         string `json:"type,omitempty"`
    UseEvaluator bool   `json:"use_evaluator,omitempty"`  // NEW
}
```

### Step 2: Update RunRequest
**File:** `orchestrator.go`
```go
type RunRequest struct {
    SessionID    string
    UserID       string
    Message      string
    UseEvaluator bool  // NEW
}
```

### Step 3: Add StreamEvent for evaluation results
**File:** `orchestrator.go`
```go
type StreamEvent struct {
    // ... existing fields ...
    EvalResult *EvaluatorResult `json:"eval_result,omitempty"`
}
```

### Step 4: Integrate evaluator into RunWithStreaming
**File:** `orchestrator.go`

After agent execution completes (around line 300), if `req.UseEvaluator`:
1. Track which worker handled the request (via `currentAgent` variable)
2. Map worker to evaluator type:
   - `job_search` -> CareerEvaluator
   - `crm_worker` -> CRMEvaluator
   - `general_worker` -> GeneralEvaluator
3. Run evaluation
4. If `score < 60` AND NOT `user_input_needed`:
   - Retry agent once with feedback
   - Re-evaluate
5. Send evaluation result via StreamEvent
6. Always return response to user (even if failed)

### Step 5: Fix evaluator threshold
**File:** `evaluator.go`

Update `ShouldRetry` to use 60 threshold:
```go
func (e *Evaluator) ShouldRetry(result *EvaluatorResult) bool {
    if result.SuccessCriteriaMet || result.UserInputNeeded {
        return false
    }
    return result.Score < 60 && e.retryCount < e.maxRetries
}
```

### Step 6: Add useEvaluator state to frontend
**File:** `useWs.jsx`
- Add `useEvaluator` state (default: false)
- Pass `use_evaluator: useEvaluator` in sendMessage

### Step 7: Add checkbox UI
**File:** `Chat.jsx`

Add checkbox in Tools dropdown (replace placeholder):
```jsx
<div className={styles.toolsMenu}>
  <label className={styles.evaluatorToggle}>
    <input
      type="checkbox"
      checked={useEvaluator}
      onChange={(e) => setUseEvaluator(e.target.checked)}
    />
    <span>Use evaluator</span>
  </label>
</div>
```

### Step 8: Handle evaluation events in frontend
**File:** `useWs.jsx`

Add handler for `on_eval_result` event to display evaluation feedback.

---

## Worker-to-Evaluator Mapping
| Worker | Evaluator |
|--------|-----------|
| job_search | CareerEvaluator |
| crm_worker | CRMEvaluator |
| general_worker | GeneralEvaluator |

## Evaluation Flow
```
User Message + use_evaluator=true
    |
Agent executes (streaming)
    |
Get final response + detect worker
    |
Run evaluator (Claude Sonnet 4.5)
    |
Score < 60 AND NOT user_input_needed?
    +-- YES: Retry agent once with feedback
    |        Re-evaluate
    |        Return response (pass or fail)
    +-- NO: Return response immediately
    |
Send eval_result event to frontend
    |
Always show response to user
```

## Default Behavior
- Checkbox unchecked by default (evaluator disabled)
- When disabled: agent flow unchanged
- When enabled: evaluator runs post-response
- Score threshold: 60 (pass/fail)
- If fails after retry: show response anyway
- If user input needed: skip retry, show immediately
