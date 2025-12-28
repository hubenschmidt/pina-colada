# AI Agent Audit Tracking

## Problem
When internal AI agents update database tables, there's no auth context, so `created_by`/`updated_by` audit columns remain null.

## Solution
Create a dedicated system user for AI agent operations.

### Implementation
1. Create a User record for the AI agent (e.g., `name: "AI Agent"`, `email: "ai-agent@system.local"`)
2. Store the AI user ID in config/environment (e.g., `AI_AGENT_USER_ID`)
3. Before AI agent operations, call `set_current_user_id(AI_AGENT_USER_ID)` from `lib/audit_context.py`
4. After operations complete, optionally clear with `clear_current_user_id()`

### Usage Pattern
```python
from lib.audit_context import set_current_user_id, clear_current_user_id

AI_AGENT_USER_ID = int(os.getenv("AI_AGENT_USER_ID", 0))

async def ai_agent_task():
    set_current_user_id(AI_AGENT_USER_ID)
    try:
        # ... AI agent database operations ...
        pass
    finally:
        clear_current_user_id()
```

### Future Considerations
- Multiple AI agents with different user IDs for granular tracking
- Add `actor_type` column to distinguish user vs AI actions
- Audit log of AI agent actions for compliance/debugging

## Status
Draft - to be expanded
