# Plan: Lightweight Notification System

## Problem
Currently, `notifications/count` API is called on every route change (`pathname` dependency in useEffect). This creates dozens of unnecessary API calls per session.

## Current Implementation
- `NotificationBell.jsx` fetches count on every `pathname` change
- No context/global state - component-local only
- No real-time updates
- Existing WebSocket is for chat streaming, not notifications

## Options Analysis

| Approach | Pros | Cons |
|----------|------|------|
| **SSE (Server-Sent Events)** | One-way push (perfect for notifications), auto-reconnect, lightweight (~5 bytes/msg), HTTP-friendly | New endpoint needed |
| **Extend Chat WebSocket** | Reuses existing infra | WebSocket is per-conversation, not global |
| **Smart Polling (recommended)** | Simplest, no new infra | Still makes periodic calls |
| **Full WebSocket** | Real-time bidirectional | Overkill for one-way notifications |

### Industry Best Practices (Sources)
- [SSE vs WebSocket comparison](https://medium.com/@asharsaleem4/long-polling-vs-server-sent-events-vs-websockets-a-comprehensive-guide-fb27c8e610d0) - SSE ideal for one-way updates
- [Real-time notification strategies](https://www.readysetcloud.io/blog/allen.helton/which-real-time-notification-method-is-for-you/) - Choose based on bidirectionality needs
- [Why SSE for real-time updates](https://talent500.com/blog/server-sent-events-real-time-updates/) - Lightweight, auto-reconnect

## Recommended Approach: Hybrid - Fetch Once + Smart Polling

The simplest, most pragmatic solution that eliminates route-change spam:

1. **Fetch once on login/app mount** - Store in context
2. **Increment locally** when user creates comments (optimistic)
3. **Decrement locally** when user reads notifications
4. **Refresh on tab focus** - Only when tab becomes visible after being hidden
5. **Optional: Slow poll** - Every 5 minutes if tab is active (configurable)

This matches how Reddit and most apps work - they don't need instant notifications, just reasonably fresh counts.

## Implementation Steps

### Step 1: Create NotificationContext
```jsx
// context/notificationContext.jsx
- Store unreadCount globally
- fetchCount() on initial mount only
- refreshOnFocus() using visibilitychange event
- incrementCount() / decrementCount() for optimistic updates
```

### Step 2: Update NotificationBell
- Remove pathname useEffect
- Consume context instead of local state
- Call context methods for mark-read actions

### Step 3: Add visibility-based refresh
```jsx
useEffect(() => {
  const handleVisibility = () => {
    if (document.visibilityState === 'visible') {
      refreshCount();
    }
  };
  document.addEventListener('visibilitychange', handleVisibility);
  return () => document.removeEventListener('visibilitychange', handleVisibility);
}, []);
```

### Step 4: Optional slow poll (if real-time matters)
- Poll every 5 minutes only if tab is visible
- Clear interval when tab is hidden

## Files to Modify

| File | Changes |
|------|---------|
| `modules/client/context/notificationContext.jsx` | **NEW** - Global notification state |
| `modules/client/app/layout.jsx` | Add NotificationProvider |
| `modules/client/components/NotificationBell/NotificationBell.jsx` | Use context, remove pathname effect |

## Future Enhancement: SSE
If real-time notifications become critical, add SSE endpoint:
- Backend: `/notifications/stream` SSE endpoint
- Frontend: `EventSource` in context, push count updates
- This can be added later without changing the context interface

## Checklist
- [ ] Create notificationContext.jsx
- [ ] Add provider to layout.jsx
- [ ] Refactor NotificationBell to use context
- [ ] Add visibilitychange listener for tab-focus refresh
- [ ] Remove pathname-based polling
- [ ] Test: verify single API call on login, refresh on tab focus
