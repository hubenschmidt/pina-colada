# Feature: Comment Notification System

## Overview

Reddit-style notification system for comment activity. A bell icon in the header displays unread notification count. Users receive notifications for direct replies to their comments AND activity on threads they've participated in. Notifications auto-mark as read when viewed and auto-dismiss after 30 days.

## User Stories

**As a user**, I want to be notified when someone replies to my comment so that I can respond promptly.

**As a user**, I want to be notified of new activity on threads I've participated in so that I can stay engaged with discussions.

**As a user**, I want to click a notification to navigate directly to the relevant comment so that I can quickly find and respond.

---

## Scenarios

### Scenario 1: Receive Direct Reply Notification

**Given** I have made a comment on a Lead
**When** Another user replies directly to my comment
**Then** A notification is created with type `direct_reply`
**And** The bell icon shows an incremented unread count

### Scenario 2: Receive Thread Activity Notification

**Given** I have participated in comments on a Deal
**When** Another user adds a new comment or reply on that Deal
**Then** A notification is created with type `thread_activity`
**And** The bell icon shows an incremented unread count

### Scenario 3: View Notification Dropdown

**Given** I have unread notifications
**When** I click the bell icon in the header
**Then** I see a dropdown list of notifications
**And** Each notification shows commenter name, preview text, entity type, and relative time
**And** Notifications are sorted by most recent first

### Scenario 4: Navigate to Comment

**Given** I am viewing the notification dropdown
**When** I click on a notification
**Then** I am navigated to the entity page (Lead, Deal, Account, Task)
**And** The page scrolls to or highlights the relevant comment
**And** The notification is marked as read

### Scenario 5: Auto-Mark as Read

**Given** I have an unread notification for a Lead's comments
**When** I navigate to that Lead's detail page
**Then** All notifications for that entity are marked as read
**And** The unread count decreases accordingly

### Scenario 6: Auto-Dismiss Old Notifications

**Given** A notification is older than 30 days
**When** The cleanup job runs
**Then** The notification is deleted from the system

---

## Architecture

### Database Schema

#### CommentNotification Table

```sql
CREATE TABLE "CommentNotification" (
    id SERIAL PRIMARY KEY,
    tenant_id INTEGER NOT NULL REFERENCES "Tenant"(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES "User"(id) ON DELETE CASCADE,
    comment_id INTEGER NOT NULL REFERENCES "Comment"(id) ON DELETE CASCADE,
    notification_type VARCHAR(20) NOT NULL, -- 'direct_reply' | 'thread_activity'
    is_read BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    read_at TIMESTAMP WITH TIME ZONE,

    CONSTRAINT unique_user_comment UNIQUE (user_id, comment_id)
);

CREATE INDEX idx_notification_user_unread ON "CommentNotification" (user_id, is_read) WHERE is_read = FALSE;
CREATE INDEX idx_notification_created ON "CommentNotification" (created_at);
```

#### CommentSubscription Table (implicit via participation)

Rather than explicit subscriptions, users are automatically subscribed to threads where they've commented. Query participation from Comment table directly.

### Notification Types

| Type | Trigger | Recipients |
|------|---------|------------|
| `direct_reply` | Comment created with `parent_comment_id` set | Author of the parent comment |
| `thread_activity` | Any comment created on an entity | All users who have commented on that entity (except comment author) |

---

## Component Specifications

### 1. Backend: Notification Model

#### `models/CommentNotification.py`

```python
class CommentNotification(Base):
    __tablename__ = "CommentNotification"

    id = Column(Integer, primary_key=True)
    tenant_id = Column(Integer, ForeignKey("Tenant.id", ondelete="CASCADE"), nullable=False)
    user_id = Column(Integer, ForeignKey("User.id", ondelete="CASCADE"), nullable=False)
    comment_id = Column(Integer, ForeignKey("Comment.id", ondelete="CASCADE"), nullable=False)
    notification_type = Column(String(20), nullable=False)  # direct_reply, thread_activity
    is_read = Column(Boolean, default=False)
    created_at = Column(DateTime(timezone=True), default=datetime.utcnow)
    read_at = Column(DateTime(timezone=True), nullable=True)

    # Relationships
    user = relationship("User")
    comment = relationship("Comment")
```

### 2. Backend: Notification Service

#### `services/notification_service.py`

```python
async def create_comment_notifications(comment: Comment, session) -> None:
    """Create notifications when a comment is created."""
    notifications = []

    # Direct reply notification
    if comment.parent_comment_id:
        parent = await get_comment_by_id(comment.parent_comment_id)
        if parent and parent.created_by != comment.created_by:
            notifications.append(CommentNotification(
                tenant_id=comment.tenant_id,
                user_id=parent.created_by,
                comment_id=comment.id,
                notification_type="direct_reply",
            ))

    # Thread activity notifications
    participants = await get_thread_participants(
        comment.commentable_type,
        comment.commentable_id,
        exclude_user=comment.created_by
    )
    for user_id in participants:
        # Skip if already getting direct_reply notification
        if not any(n.user_id == user_id for n in notifications):
            notifications.append(CommentNotification(
                tenant_id=comment.tenant_id,
                user_id=user_id,
                comment_id=comment.id,
                notification_type="thread_activity",
            ))

    session.add_all(notifications)

async def get_unread_count(user_id: int, tenant_id: int) -> int:
    """Get count of unread notifications."""
    ...

async def get_notifications(user_id: int, tenant_id: int, limit: int = 20) -> list:
    """Get notifications with comment and entity details."""
    ...

async def mark_as_read(notification_ids: list[int], user_id: int) -> None:
    """Mark specific notifications as read."""
    ...

async def mark_entity_as_read(user_id: int, entity_type: str, entity_id: int) -> None:
    """Mark all notifications for an entity as read (when viewing entity)."""
    ...

async def cleanup_old_notifications(days: int = 30) -> int:
    """Delete notifications older than specified days."""
    ...
```

### 3. Backend: API Routes

#### `api/routes/notifications.py`

```python
router = APIRouter(prefix="/notifications", tags=["notifications"])

@router.get("/count")
async def get_unread_count(request: Request):
    """Get unread notification count for badge."""
    count = await get_unread_count(request.state.user_id, request.state.tenant_id)
    return {"unread_count": count}

@router.get("")
async def list_notifications(request: Request, limit: int = 20):
    """Get notifications with comment previews."""
    return await get_notifications(
        request.state.user_id,
        request.state.tenant_id,
        limit
    )

@router.post("/mark-read")
async def mark_notifications_read(request: Request, notification_ids: list[int]):
    """Mark specific notifications as read."""
    await mark_as_read(notification_ids, request.state.user_id)
    return {"status": "ok"}

@router.post("/mark-entity-read")
async def mark_entity_notifications_read(
    request: Request,
    entity_type: str,
    entity_id: int
):
    """Mark all notifications for an entity as read."""
    await mark_entity_as_read(
        request.state.user_id,
        entity_type,
        entity_id
    )
    return {"status": "ok"}
```

**Response Shape for `GET /notifications`:**

```json
{
  "notifications": [
    {
      "id": 123,
      "type": "direct_reply",
      "is_read": false,
      "created_at": "2025-11-30T10:00:00Z",
      "comment": {
        "id": 456,
        "content": "Great point! I think we should...",
        "created_by_name": "Jennifer Lev",
        "created_at": "2025-11-30T10:00:00Z"
      },
      "entity": {
        "type": "Lead",
        "id": 789,
        "display_name": "Acme Corp Opportunity",
        "url": "/leads/789"
      }
    }
  ],
  "unread_count": 5
}
```

### 4. Frontend: Notification Bell Component

#### `components/NotificationBell.tsx`

```typescript
interface Notification {
  id: number;
  type: "direct_reply" | "thread_activity";
  is_read: boolean;
  created_at: string;
  comment: {
    id: number;
    content: string;
    created_by_name: string;
    created_at: string;
  };
  entity: {
    type: string;
    id: number;
    display_name: string;
    url: string;
  };
}

const NotificationBell = () => {
  const [unreadCount, setUnreadCount] = useState(0);
  const [notifications, setNotifications] = useState<Notification[]>([]);
  const [isOpen, setIsOpen] = useState(false);

  // Poll for unread count every 30 seconds
  useEffect(() => {
    const fetchCount = async () => {
      const { unread_count } = await getUnreadCount();
      setUnreadCount(unread_count);
    };
    fetchCount();
    const interval = setInterval(fetchCount, 30000);
    return () => clearInterval(interval);
  }, []);

  const handleOpen = async () => {
    setIsOpen(true);
    const data = await getNotifications();
    setNotifications(data.notifications);
  };

  const handleNotificationClick = async (notification: Notification) => {
    await markAsRead([notification.id]);
    router.push(notification.entity.url);
    setIsOpen(false);
  };

  return (
    <div className="relative">
      <button onClick={handleOpen}>
        <Bell size={20} />
        {unreadCount > 0 && (
          <span className="absolute -top-1 -right-1 bg-red-500 text-white text-xs rounded-full w-5 h-5 flex items-center justify-center">
            {unreadCount > 99 ? "99+" : unreadCount}
          </span>
        )}
      </button>

      {isOpen && (
        <NotificationDropdown
          notifications={notifications}
          onNotificationClick={handleNotificationClick}
          onClose={() => setIsOpen(false)}
        />
      )}
    </div>
  );
};
```

### 5. Frontend: Integration with Comment Creation

Modify `comments.py` to trigger notification creation:

```python
@router.post("")
async def create_comment_route(request: Request, data: CommentCreate):
    comment = await create_comment(comment_data)

    # Create notifications for relevant users
    await create_comment_notifications(comment, session)

    return _comment_to_dict(comment)
```

### 6. Frontend: Auto-Mark Read on Entity View

Entity detail pages should mark notifications as read:

```typescript
// In entity detail page (e.g., LeadForm, AccountForm)
useEffect(() => {
  if (entityId) {
    markEntityNotificationsRead(entityType, entityId);
  }
}, [entityId, entityType]);
```

---

## File Structure

```
modules/agent/
├── migrations/
│   └── 039_comment_notifications.sql       # NEW: Create notification table
├── src/
│   ├── models/
│   │   └── CommentNotification.py          # NEW: Notification model
│   ├── repositories/
│   │   └── notification_repository.py      # NEW: Data access
│   ├── services/
│   │   └── notification_service.py         # NEW: Business logic
│   └── api/routes/
│       ├── notifications.py                # NEW: Notification endpoints
│       └── comments.py                     # MODIFY: Trigger notifications

modules/client/
├── components/
│   ├── NotificationBell.tsx                # NEW: Bell icon component
│   ├── NotificationDropdown.tsx            # NEW: Dropdown list
│   └── HeaderAuthed.tsx                    # MODIFY: Add NotificationBell
├── api/
│   └── index.ts                            # MODIFY: Add notification API calls
```

---

## Verification Checklist

### Functional Requirements
- [ ] Bell icon displays in header when authenticated
- [ ] Red badge shows unread count (0 = hidden)
- [ ] Clicking bell opens dropdown with notification list
- [ ] Each notification shows: commenter name, preview (50 chars), entity name, relative time
- [ ] Clicking notification navigates to entity and marks as read
- [ ] Direct reply notifications created when someone replies to user's comment
- [ ] Thread activity notifications created when someone comments on entity user has commented on
- [ ] Viewing entity page marks all notifications for that entity as read
- [ ] Notifications older than 30 days are auto-deleted

### Non-Functional Requirements
- [ ] Performance: Unread count API responds < 100ms
- [ ] Performance: Notification list loads < 300ms
- [ ] Security: Users can only see their own notifications
- [ ] Security: Notifications filtered by tenant_id

### Edge Cases
- [ ] User comments on their own thread (no self-notification)
- [ ] User deletes their comment (related notifications should cascade delete)
- [ ] Entity is deleted (notifications for that entity cascade delete)
- [ ] User has 100+ unread notifications (pagination or limit display)
- [ ] Multiple rapid comments don't create duplicate notifications

---

## Implementation Notes

### Estimate of Scope
Medium - 2-3 days of implementation

### Dependencies
- Existing Comment model with `created_by` field
- Existing User model for recipient tracking
- HeaderAuthed component for bell placement

### Out of Scope
- Email/push notifications (future enhancement)
- Notification preferences/settings (future enhancement)
- Real-time WebSocket updates (using polling instead)
- Notification grouping (e.g., "3 replies to your comment")
- @mentions in comments
