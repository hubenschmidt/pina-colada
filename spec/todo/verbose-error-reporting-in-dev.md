# Verbose Error Handling & Logging

## Problem
1. Errors like `'dict' object has no attribute 'id'` are not visible in docker logs
2. Frontend doesn't display detailed errors to users in dev mode
3. No centralized error UI component in frontend
4. No verbose logging flag for development/testing

## Goals
1. Add error banner/toast component to frontend for user-visible errors
2. Add `VERBOSE_ERROR_LOGS=true` env var for detailed error display in dev
3. Fix the underlying contact lookup bug causing the error
4. Improve backend error logging for easier debugging

---

## Part 1: Fix the Immediate Bug

**Bug:** `list_entities("contact")` fails because `search_contacts()` returns `List[Dict]` but code tries to access `.id` attribute.

**File:** `agent/tools/crm_tools.py` (line ~257-264)

```python
# Current (broken):
for contact in results[:20]:
    formatted.append(f"- Contact id={contact.id}")  # contact is a dict!

# Fix:
for contact in results[:20]:
    formatted.append(f"- Contact id={contact['id']}")  # use dict access
```

---

## Part 2: Frontend Error Display

### 2a. Create Error Banner Component
**File:** `modules/client/components/ErrorBanner/ErrorBanner.jsx` (new)

- Mantine Alert component at top of page
- Persistent until dismissed (user must click X)
- Red styling with error icon
- Shows error message + stack trace/details if `NEXT_PUBLIC_VERBOSE_ERRORS=true`
- Similar pattern to existing `DeleteConfirmBanner` component

### 2b. Add Error Context Provider
**File:** `modules/client/context/errorContext.jsx` (new)

- React context to manage error state globally
- `showError(message, details)` - display error toast
- `clearError()` - dismiss error
- Wrap app in this provider

### 2c. Update Chat Component
**File:** `modules/client/components/Chat/Chat.jsx`

- When WebSocket sends error, call `showError()`
- Display backend error messages in verbose mode
- Show connection errors to user

### 2d. Update useWs Hook
**File:** `modules/client/hooks/useWs.jsx`

- On `socket.onerror`, extract error details
- Pass to error context for display
- Log verbose details if enabled

---

## Part 3: Environment Configuration

### Frontend (.env.local.example)
```
# Error display verbosity (development only)
NEXT_PUBLIC_VERBOSE_ERRORS=true
```

### Backend (.env.example)
```
# Verbose error logging (development only)
VERBOSE_ERROR_LOGS=true
```

---

## Part 4: Backend Error Improvements

### 4a. Update graph.py error handler
**File:** `agent/graph.py`

- Include stack trace in error message when `VERBOSE_ERROR_LOGS=true`
- Send structured error to frontend via WebSocket

### 4b. Add error type to WebSocket protocol
Current: `{"on_chat_model_stream": "Sorry, there was an error..."}`
New: `{"type": "error", "message": "...", "details": "...", "stack": "..."}`

---

## Files to Modify/Create

| File | Action |
|------|--------|
| `agent/tools/crm_tools.py` | FIX - dict access for contacts |
| `modules/client/components/ErrorBanner/` | CREATE |
| `modules/client/context/errorContext.jsx` | CREATE |
| `modules/client/components/Chat/Chat.jsx` | MODIFY - use error context |
| `modules/client/hooks/useWs.jsx` | MODIFY - surface errors |
| `modules/client/.env.local.example` | MODIFY - add VERBOSE_ERRORS |
| `agent/graph.py` | MODIFY - structured error messages |
| `.env.example` (backend) | MODIFY - add VERBOSE_ERROR_LOGS |

---

## Implementation Order

1. **Fix the bug** - `crm_tools.py` dict access for contacts
2. **Add env vars** - `NEXT_PUBLIC_VERBOSE_ERRORS` and `VERBOSE_ERROR_LOGS`
3. **Create ErrorBanner component** - Mantine Alert, persistent, dismissible
4. **Create error context** - provider with `showError(msg, details)` and `clearError()`
5. **Update Chat/useWs** - surface errors to context
6. **Update graph.py** - structured error messages with stack trace
7. **Test** - verify errors display in dev mode
