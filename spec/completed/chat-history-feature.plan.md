# Chat History & Persistence Feature

## Overview

Implement chat history persistence similar to ChatGPT/Claude, allowing users to:
- View a list of past conversations
- Resume any previous conversation
- Create new conversations
- Archive/delete old conversations

## Current State

### Frontend
- **Chat.jsx** - Messages in React state, lost on refresh
- **useWs.jsx** - UUID generated per session, no persistence
- **Sidebar.jsx** - Projects section at top (lines 73-104), uses `FolderKanban` icon

### Backend
- **server.py** - WebSocket with UUID-based thread_id
- **orchestrator.py** - Uses `MemorySaver()` (in-memory only)
- **Database** - No Conversation/Message tables exist

---

## Implementation Plan

### Phase 1: Database Schema

**New Migration:** `056_conversation_tables.sql`

```sql
-- Conversation metadata
CREATE TABLE "Conversation" (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL REFERENCES "Tenant"(id),
    user_id BIGINT NOT NULL REFERENCES "User"(id),
    thread_id UUID NOT NULL UNIQUE,
    title TEXT,  -- Auto-generated from first message or user-set
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    archived_at TIMESTAMPTZ  -- Soft delete
);

CREATE INDEX idx_conversation_user ON "Conversation"(user_id, archived_at);
CREATE INDEX idx_conversation_thread ON "Conversation"(thread_id);

-- Individual messages
CREATE TABLE "ConversationMessage" (
    id BIGSERIAL PRIMARY KEY,
    conversation_id BIGINT NOT NULL REFERENCES "Conversation"(id) ON DELETE CASCADE,
    role TEXT NOT NULL CHECK (role IN ('user', 'assistant')),
    content TEXT NOT NULL,
    token_usage JSONB,  -- {input, output, total}
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_message_conversation ON "ConversationMessage"(conversation_id, created_at);
```

---

### Phase 2: Backend API

**New Files:**
- `repositories/conversation_repository.py`
- `services/conversation_service.py`
- `routers/conversation_router.py`

**Endpoints:**
| Method | Path | Description |
|--------|------|-------------|
| GET | `/conversations` | List user's conversations (paginated) |
| GET | `/conversations/:id` | Get conversation with messages |
| POST | `/conversations` | Create new conversation |
| PATCH | `/conversations/:id` | Update title |
| DELETE | `/conversations/:id` | Archive (soft delete) |
| DELETE | `/conversations/:id/permanent` | Hard delete (must be archived first) |
| POST | `/conversations/:id/messages` | Add message (internal use) |

**Repository Functions:**
```python
async def list_conversations(user_id: int, limit: int = 50) -> List[Conversation]
async def get_conversation(thread_id: str) -> Optional[Conversation]
async def get_conversation_with_messages(thread_id: str) -> ConversationWithMessages
async def create_conversation(user_id: int, tenant_id: int, thread_id: str, title: str) -> Conversation
async def update_conversation_title(thread_id: str, title: str) -> Conversation
async def archive_conversation(thread_id: str) -> None
async def delete_conversation_permanent(thread_id: str) -> None
async def add_message(conversation_id: int, role: str, content: str, token_usage: dict) -> Message
```

---

### Phase 3: WebSocket Message Persistence

**Modify:** `agent/graph.py`

On each message exchange:
1. Check if conversation exists for thread_id
2. If not, create conversation with auto-title from first user message
3. Save user message to ConversationMessage
4. After assistant response, save assistant message with token_usage

**Modify:** `agent/orchestrator.py`

On `run_streaming()`:
- After graph completion, persist final assistant response
- Include token_usage in message metadata

---

### Phase 4: Frontend - Chat History Panel

**New Component:** `components/ChatHistory/ChatHistory.jsx`

Location: Sidebar, ABOVE Projects section

```jsx
// Sidebar.jsx modification (insert before Projects)
<ChatHistory
  conversations={conversations}
  activeId={activeConversationId}
  onSelect={handleSelectConversation}
  onNew={handleNewConversation}
  onArchive={handleArchiveConversation}
/>
```

**UI Elements:**
- "New Chat" button (+ icon)
- Scrollable list of conversations
  - Title (truncated)
  - Relative timestamp ("2h ago", "Yesterday")
  - Active state highlight
- Click to load/resume conversation
- Hover: Archive/delete option

**Styling:** Match existing Sidebar nav items (lime-500 accent)

---

### Phase 5: Frontend State Management

**New Context:** `context/conversationContext.jsx`

```javascript
const ConversationContext = createContext({
  conversations: [],
  activeConversation: null,
  loadConversations: async () => {},
  selectConversation: async (threadId) => {},
  createConversation: async () => {},
  archiveConversation: async (threadId) => {},
});
```

**Modify:** `hooks/useWs.jsx`

- Accept optional `threadId` prop to resume conversation
- On connection, if threadId provided, load existing messages
- On new message, sync with backend persistence

**Modify:** `components/Chat/Chat.jsx`

- Integrate with conversationContext
- Load messages when activeConversation changes
- Update title after first message (auto-generate)

---

### Phase 6: Auto-Title Generation (LLM)

Use Claude Haiku to generate 3-5 word summary title:
- Trigger after first assistant response completes
- Background task, non-blocking
- Pass first user message + assistant response snippet
- Update conversation title via API
- Fallback: first 50 chars of user message if LLM fails

---

### Phase 7: Archive & Delete

**Soft delete (archive):**
- Set `archived_at` timestamp
- Hidden from list, but recoverable

**Hard delete:**
- `DELETE /conversations/:id/permanent` endpoint
- Cascades to delete all ConversationMessage rows
- Only available for already-archived conversations
- UI: Show in "Archived" section with permanent delete option

---

## Files to Create/Modify

| File | Action |
|------|--------|
| `migrations/056_conversation_tables.sql` | CREATE |
| `repositories/conversation_repository.py` | CREATE |
| `services/conversation_service.py` | CREATE |
| `routers/conversation_router.py` | CREATE |
| `agent/graph.py` | MODIFY - add persistence |
| `agent/orchestrator.py` | MODIFY - save messages |
| `components/ChatHistory/ChatHistory.jsx` | CREATE |
| `components/ChatHistory/ChatHistory.module.css` | CREATE |
| `components/Sidebar/Sidebar.jsx` | MODIFY - add ChatHistory |
| `context/conversationContext.jsx` | CREATE |
| `hooks/useWs.jsx` | MODIFY - thread resumption |
| `components/Chat/Chat.jsx` | MODIFY - context integration |
| `api/index.js` | MODIFY - add conversation endpoints |

---

## API Client Functions

```javascript
// api/index.js additions
export const getConversations = () => api.get('/conversations');
export const getConversation = (id) => api.get(`/conversations/${id}`);
export const createConversation = () => api.post('/conversations');
export const updateConversationTitle = (id, title) => api.patch(`/conversations/${id}`, { title });
export const archiveConversation = (id) => api.delete(`/conversations/${id}`);
export const deleteConversationPermanent = (id) => api.delete(`/conversations/${id}/permanent`);
```

---

## Testing

1. **Unit Tests:**
   - conversation_repository CRUD operations
   - Message persistence on WebSocket exchange

2. **Integration Tests:**
   - Create conversation -> send messages -> reload -> verify messages
   - Archive conversation -> verify not in list
   - Resume conversation -> verify history loads
   - Hard delete -> verify cascade deletion

3. **E2E Tests:**
   - Full flow: New chat -> send message -> refresh -> resume
