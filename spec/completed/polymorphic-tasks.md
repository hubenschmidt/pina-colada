# Feature: Polymorphic Tasks with Tasks Tab

## Overview

Enable Tasks to be assigned to any entity (Project, Account, Lead, Deal) using a polymorphic association pattern. Add a Tasks tab to the Sidebar with table view supporting project-scoped and global views.

## User Story

As a user, I want to create tasks linked to any entity (Project, Account, Lead, Deal) and view all my tasks in a centralized Tasks tab, so that I can track work across my entire workflow.

---

## Scenarios

### Scenario 1: Create Task on a Deal

**Given** I am viewing a Deal detail page
**When** I click "Add Task" and fill in the task form
**Then** The task is created with `entity_type='Deal'` and `entity_id={deal_id}`

### Scenario 2: View Tasks Scoped to Project

**Given** I have a project selected in the sidebar
**When** I open the Tasks tab
**Then** I see tasks for that project AND tasks for entities within that project (Deals, Leads)
**And** A badge shows the current project scope
**And** An "Entity" column shows which entity each task belongs to

### Scenario 3: View Tasks Globally

**Given** I toggle to "All Tasks" view
**When** The table loads
**Then** I see all tasks across all entities within my tenant

---

## Architecture

### Polymorphic Pattern

The existing Task model uses `taskable_type` (string) + `taskable_id` (integer). This open pattern accepts any entity type string, allowing flexibility for future entity types without code changes.

**Supported Entity Types:** Project, Account, Lead, Deal (and any future entities)

### Database Schema

No schema changes required - existing Task table has `taskable_type` and `taskable_id`.

Add validation index (optional performance enhancement):

```sql
-- Ensure fast lookups by entity
CREATE INDEX IF NOT EXISTS idx_task_entity
ON "Task"(taskable_type, taskable_id);
```

---

## Component Specifications

### 1. Backend: Task Service

#### `services/task_service.py`

```python
def get_tasks_for_project_scope(project_id: int, session) -> list[Task]:
    """
    Get all tasks within a project's scope:
    - Tasks directly on the Project
    - Tasks on Deals belonging to the Project
    - Tasks on Leads linked to the Project (via LeadProject)
    - Tasks on Accounts linked to the Project (via AccountProject)
    """
    ...

def resolve_entity_display(task: Task, session) -> dict:
    """Return entity display info for a task."""
    # Query entity for name/title based on taskable_type
    return {
        "type": task.taskable_type,
        "id": task.taskable_id,
        "display_name": entity.name or entity.title,
        "url": f"/{task.taskable_type.lower()}s/{task.taskable_id}"
    }
```

### 2. API Routes Enhancement

#### `api/routes/tasks.py`

```python
@router.get("/tasks")
async def list_tasks(
    project_id: int | None = Query(None),
    scope: str = Query("project"),  # "project" or "global"
    page: int = Query(1),
    page_size: int = Query(20),
    session = Depends(get_session),
    user = Depends(get_current_user),
):
    """
    List tasks with scope filtering.

    - scope="project" + project_id: Tasks within project's entity graph
    - scope="global": All tasks in tenant
    """
    ...

@router.post("/tasks")
async def create_task(
    task_data: TaskCreate,
    session = Depends(get_session),
    user = Depends(get_current_user),
):
    """Create task linked to an entity."""
    ...
```

**Response Shape:**

```json
{
  "items": [
    {
      "id": 1,
      "title": "Follow up on proposal",
      "description": "...",
      "due_date": "2025-12-01",
      "priority": { "id": 1, "name": "High", "color": "#ff0000" },
      "status": { "id": 2, "name": "In Progress", "color": "#0000ff" },
      "assigned_to": { "id": 5, "name": "John Doe" },
      "entity": {
        "type": "Deal",
        "id": 42,
        "display_name": "Acme Corp Contract",
        "url": "/deals/42"
      },
      "created_at": "2025-01-15T10:00:00Z"
    }
  ],
  "currentPage": 1,
  "totalPages": 5,
  "total": 100,
  "pageSize": 20,
  "scope": {
    "type": "project",
    "project_id": 3,
    "project_name": "Q1 Sales Initiative"
  }
}
```

### 3. Frontend: Sidebar Tasks Tab

#### `components/Sidebar/Sidebar.tsx` (modification)

Add Tasks section below existing navigation:

```typescript
// New section in sidebar
<NavButton
  href="/tasks"
  icon={<CheckSquare size={16} />}
  label="Tasks"
  isActive={pathname === "/tasks"}
  isCollapsed={isCollapsed}
/>
```

### 4. Frontend: Tasks Page

#### `app/tasks/page.tsx`

```typescript
export default function TasksPage() {
  const { selectedProject } = useProjectContext();
  const [scope, setScope] = useState<"project" | "global">("project");

  return (
    <Stack>
      {/* Scope Badge */}
      {scope === "project" && selectedProject && (
        <Badge variant="light">
          Project: {selectedProject.name}
        </Badge>
      )}

      {/* Scope Toggle */}
      <SegmentedControl
        value={scope}
        onChange={setScope}
        data={[
          { label: "Project Tasks", value: "project" },
          { label: "All Tasks", value: "global" },
        ]}
      />

      {/* Tasks Table */}
      <TasksTable
        projectId={scope === "project" ? selectedProject?.id : undefined}
        scope={scope}
      />
    </Stack>
  );
}
```

### 5. Frontend: Tasks Table Component

#### `components/Tasks/TasksTable.tsx`

```typescript
interface TasksTableProps {
  projectId?: number;
  scope: "project" | "global";
}

const columns = [
  { key: "title", label: "Task" },
  { key: "entity", label: "Entity", render: EntityCell },
  { key: "status", label: "Status", render: StatusBadge },
  { key: "priority", label: "Priority", render: PriorityBadge },
  { key: "assigned_to", label: "Assignee" },
  { key: "due_date", label: "Due Date", render: DateCell },
];

// EntityCell component shows entity type + name with link
function EntityCell({ entity }) {
  return (
    <Group gap="xs">
      <Text size="xs" c="dimmed">{entity.type}</Text>
      <Anchor href={entity.url}>{entity.display_name}</Anchor>
    </Group>
  );
}
```

### 6. Frontend: Task Form (Entity Selection)

#### `components/Tasks/TaskForm.tsx`

When creating a task from the Tasks page (not from an entity detail page):

```typescript
interface TaskFormProps {
  entityType?: string;  // Pre-selected when on entity detail page
  entityId?: number;
  onSubmit: (task: TaskCreate) => void;
}

// If no entity pre-selected, show entity picker
<EntityPicker
  value={{ type: entityType, id: entityId }}
  onChange={setEntity}
  projectId={selectedProject?.id}  // Filter to project scope
/>
```

---

## File Structure

```
modules/agent/
├── src/
│   ├── services/
│   │   └── task_service.py           # NEW: Task business logic
│   ├── controllers/
│   │   └── task_controller.py        # MODIFY: Add scope filtering
│   └── api/routes/
│       └── tasks.py                  # MODIFY: Enhanced endpoints

modules/client/
├── app/
│   └── tasks/
│       └── page.tsx                  # NEW: Tasks page
├── components/
│   ├── Sidebar/
│   │   └── Sidebar.tsx               # MODIFY: Add Tasks nav
│   └── Tasks/
│       ├── TasksTable.tsx            # NEW: Table component
│       ├── TaskForm.tsx              # NEW: Create/edit form
│       └── EntityPicker.tsx          # NEW: Entity selection
└── lib/api/
    └── tasks.ts                      # NEW: API client
```

---

## Verification Checklist

### Functional Requirements
- [ ] Tasks can be created linked to Project, Account, Lead, or Deal
- [ ] Tasks tab visible in sidebar
- [ ] Project-scoped view shows tasks on project + nested entities
- [ ] Global view shows all tenant tasks
- [ ] Entity column displays type and name with link
- [ ] Scope badge displays current project when project-scoped

### Non-Functional Requirements
- [ ] Performance: Task list loads < 500ms for 100 tasks
- [ ] Security: Tasks filtered by tenant_id

### Edge Cases
- [ ] Task on deleted entity displays gracefully (orphan handling)
- [ ] No project selected shows prompt to select or switch to global
- [ ] Empty state for no tasks

---

## Implementation Notes

### Files to Modify
- `modules/agent/src/models/Task.py` - No changes needed
- `modules/agent/src/controllers/task_controller.py` - Add scope filtering
- `modules/client/components/Sidebar/Sidebar.tsx` - Add Tasks nav item

### Files to Create
- `modules/agent/src/services/task_service.py`
- `modules/client/app/tasks/page.tsx`
- `modules/client/components/Tasks/TasksTable.tsx`
- `modules/client/components/Tasks/TaskForm.tsx`
- `modules/client/lib/api/tasks.ts`

### Dependencies
- Existing Task model with polymorphic columns
- Existing Status model for task priority/status
- Project context from `useProjectContext()`

### Out of Scope
- Task notifications/reminders
- Recurring tasks
- Task templates
- Bulk task operations
- Task comments/activity log
