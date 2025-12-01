# Feature: Polymorphic Comments

## Overview
A polymorphic Comment table that allows users to add comments to any Account, Lead, Deal, or Task record. Comments are displayed at the bottom of entity detail pages (similar to Jira's comment section), providing a chronological activity thread for collaboration and context.

## User Story
As a CRM user, I want to add comments to any record (Account, Lead, Deal, Task) so that I can track discussions, decisions, and context over time.

---

## Scenarios

### Scenario 1: Add a comment to a record

**Given** I am viewing an Account/Lead/Deal/Task detail page
**When** I type a comment in the comment input and click "Add Comment"
**Then** the comment is saved and appears in the comments list with my name and timestamp

### Scenario 2: View comments on a record

**Given** I am viewing an Account/Lead/Deal/Task detail page
**When** the page loads
**Then** I see all comments for that record in chronological order (oldest first)

### Scenario 3: Edit my own comment

**Given** I have previously added a comment to a record
**When** I click the edit button on my comment
**Then** I can modify the text and save the updated comment

### Scenario 4: Delete my own comment

**Given** I have previously added a comment to a record
**When** I click the delete button on my comment and confirm
**Then** the comment is removed from the list

### Scenario 5: Comments persist when creating a new record

**Given** I am creating a new Account/Lead/Deal/Task (unsaved)
**When** I add comments before saving the record
**Then** the comments are stored locally and saved when I submit the form

---

## Verification Checklist

### Functional Requirements
- [ ] Comment table created with polymorphic columns (commentable_type, commentable_id)
- [ ] CRUD API endpoints for comments
- [ ] CommentsSection component displays on Account, Lead, Deal, Task detail pages
- [ ] Comments show author name and timestamp
- [ ] Users can only edit/delete their own comments
- [ ] Comments ordered chronologically (oldest first)
- [ ] Pending comments supported for new (unsaved) records

### Non-Functional Requirements
- [ ] Performance: Comments load in < 200ms for up to 100 comments
- [ ] Security: Tenant isolation enforced (users only see comments within their tenant)
- [ ] Accessibility: Form inputs have proper labels; timestamps are screen-reader friendly

### Edge Cases
- [ ] Empty comment text rejected (validation)
- [ ] Very long comments handled (text truncation or scroll)
- [ ] Deleted parent record cascades to delete comments

---

## Implementation Notes

### Database Schema

```sql
CREATE TABLE IF NOT EXISTS "Comment" (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL REFERENCES "Tenant"(id),
    commentable_type TEXT NOT NULL,  -- 'Account', 'Lead', 'Deal', 'Task'
    commentable_id BIGINT NOT NULL,
    content TEXT NOT NULL,
    created_by BIGINT NOT NULL REFERENCES "User"(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_comment_entity
    ON "Comment"(commentable_type, commentable_id);
CREATE INDEX IF NOT EXISTS idx_comment_tenant
    ON "Comment"(tenant_id);
```

### API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/comments?entity_type={type}&entity_id={id}` | List comments for an entity |
| POST | `/api/comments` | Create a new comment |
| PUT | `/api/comments/{id}` | Update a comment |
| DELETE | `/api/comments/{id}` | Delete a comment |

### Files to Modify

**Backend:**
- `modules/agent/migrations/0XX_comment.sql` - Migration
- `modules/agent/src/models/Comment.py` - SQLAlchemy model
- `modules/agent/src/repositories/comment_repository.py` - Data access
- `modules/agent/src/services/comment_service.py` - Business logic
- `modules/agent/src/api/routes/comments.py` - API routes
- `modules/agent/src/api/routes/__init__.py` - Register routes

**Frontend:**
- `modules/client/api/index.ts` - API client methods
- `modules/client/components/CommentsSection.tsx` - Comments UI component
- `modules/client/components/AccountForm/AccountForm.tsx` - Add CommentsSection
- `modules/client/components/LeadTracker/LeadForm.tsx` - Add CommentsSection
- `modules/client/components/DealForm.tsx` - Add CommentsSection (if exists)
- `modules/client/app/tasks/[id]/page.tsx` - Add CommentsSection

### Dependencies
- Existing User model (for created_by FK)
- Existing Tenant model (for tenant_id FK)
- Existing NotesSection pattern (reference implementation)

### Out of Scope
- @mentions or notifications
- Rich text formatting (markdown)
- File attachments on comments
- Comment threading/replies
