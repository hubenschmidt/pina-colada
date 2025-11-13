# JIRA TICKET: Job Search Chat Enhancements

## Ticket Type

**Story / Enhancement**

## Summary

Enhance job search chat with export functionality, dedicated job leads panel, and streamlined application tracking

## Priority

Medium-High

## Components

- Frontend (React/Next.js)
- Backend Agent (Python/FastAPI)
- Database (Supabase/Postgres)

---

## Description

Based on user feedback from real chat sessions, we need to enhance the job search chat experience with four key improvements:

1. **Chat Export Functionality**: Allow users to export entire chat conversations to .txt file
2. **Dedicated Job Leads Panel**: Display job leads in a separate, persistent window (not the chat interface)
3. **Quick Action Buttons**: Add UI controls to quickly mark jobs as "Applied", "Remove", or "Do Not Apply"
4. **Agent Database Actions**: Enable the chatbot to perform the same database actions via natural language

### Current State

- Chat UI exists at `modules/client/components/Chat/Chat.tsx`
- Job tracking exists at `modules/client/components/JobTracker/JobTracker.tsx`
- Agent can search for jobs and check applied jobs via `worker_tools.py`
- Database schema includes Job table with status field (applied, interviewing, rejected, offer, accepted)

### Problem Statement

Currently:

- Users cannot easily save their chat conversations for reference
- Job leads are presented inline in chat, making them hard to track and compare
- Users must manually navigate to JobTracker to record applications
- No way to quickly dismiss irrelevant job leads
- Agent cannot directly update job status when user says "I applied to X"

---

## Acceptance Criteria

### 1. Chat Export Feature

- [ ] Single button in chat interface labeled "Export Chat"
- [ ] Clicking exports entire conversation to `.txt` file
- [ ] Filename format: `pinacolada-chat-{timestamp}.txt`
- [ ] Export includes all messages (user + assistant) with timestamps
- [ ] Export preserves formatting (line breaks, links)

### 2. Job Leads Panel

- [ ] When agent responds with job leads, jobs are automatically added to database with `lead_type = 'Qualifying'`
- [ ] Job Leads Panel opens automatically showing database-backed leads
- [ ] Panel queries jobs where `lead_type IS NOT NULL`
- [ ] Panel displays leads in structured list format (not chat bubbles)
- [ ] Each job shows: Company, Title, URL (as clickable link), Lead Type badge
- [ ] Panel includes filter dropdown to show any combination of lead types:
  - Qualifying (default)
  - Cold
  - Warm
  - Hot
  - All (shows all lead types)
- [ ] Panel persists across page refreshes (reads from database)
- [ ] Panel can be manually opened/closed by user
- [ ] Panel shows count of filtered jobs
- [ ] User can manually change lead type for any job in the panel

### 3. Quick Action Buttons on Job Leads

Each job in the leads panel has three action buttons:

- [ ] **"Applied"** button

  - Updates job in database: `status = 'applied'`, `lead_type = NULL`
  - Sets `date = NOW()` (if not already set)
  - Removes job from leads panel (no longer shows in filtered view)
  - Job now appears in JobTracker with 'applied' status
  - Shows success toast notification

- [ ] **"Remove"** button

  - Deletes job record from database entirely
  - Removes from leads panel immediately
  - No confirmation required
  - Shows success toast notification

- [ ] **"Do Not Apply"** button
  - Updates job in database: `status = 'do_not_apply'`, `lead_type = NULL`
  - Removes job from leads panel
  - Job now appears in JobTracker with 'do_not_apply' status (distinct styling)
  - Agent's `check_applied_jobs` tool filters these out in future searches
  - Shows success toast notification

### 4. Agent Natural Language Database Actions

Agent should respond to commands like:

- [ ] "I applied to Alto Neuroscience" → Adds job with `status = 'applied'`
- [ ] "Mark Ramp as applied" → Same as above
- [ ] "I don't want to apply to CoreWeave" → Adds with `status = 'do_not_apply'`
- [ ] "Remove Tome from my list" → Adds with `status = 'do_not_apply'`

Implementation requirements:

- [ ] Create new tool in `worker_tools.py`: `update_job_status(company, job_title, status, job_url, notes)`
- [ ] Tool should handle fuzzy matching on company name
- [ ] Tool should be available to Job Hunter agent and Worker agent
- [ ] Tool should return confirmation message
- [ ] Agent should acknowledge action in chat

---

## Technical Implementation Details

### Database Schema Changes

**Add two new fields to Job table:**

1. **New status value: `'do_not_apply'`**

```sql
ALTER TYPE job_status ADD VALUE IF NOT EXISTS 'do_not_apply';
```

2. **New column: `lead_type`**

```sql
-- Add lead_type column as ENUM
CREATE TYPE lead_type AS ENUM ('Qualifying', 'Cold', 'Warm', 'Hot');

ALTER TABLE "Job" ADD COLUMN lead_type lead_type;

-- Add index for filtering leads panel
CREATE INDEX idx_lead_type ON "Job"(lead_type) WHERE lead_type IS NOT NULL;

COMMENT ON COLUMN "Job".lead_type IS 'Type of job lead: Qualifying (initial), Cold, Warm, Hot. NULL if not a lead.';
```

**New status value: `'lead'`** (optional, for clarity)

```sql
-- Add 'lead' status to distinguish un-applied leads from applied jobs
ALTER TYPE job_status ADD VALUE IF NOT EXISTS 'lead';
```

**Recommendation**: Use `status='lead'` for new leads, then transition to `'applied'` when user clicks Applied button

### Frontend Changes

#### 1. Chat Export (`modules/client/components/Chat/Chat.tsx`)

Add export button to header:

```tsx
const exportChat = () => {
  const timestamp = new Date().toISOString().replace(/[:.]/g, "-");
  const content = messages
    .map((msg) => `[${msg.timestamp}] ${msg.role}: ${msg.content}\n`)
    .join("\n");

  const blob = new Blob([content], { type: "text/plain" });
  const url = URL.createObjectURL(blob);
  const a = document.createElement("a");
  a.href = url;
  a.download = `pinacolada-chat-${timestamp}.txt`;
  a.click();
};
```

#### 2. Job Leads Panel (`modules/client/components/JobLeadsPanel/JobLeadsPanel.tsx` - NEW)

New component structure:

```tsx
interface JobLead {
  id: string; // UUID from database
  company: string;
  title: string;
  url: string;
  lead_type: "Qualifying" | "Cold" | "Warm" | "Hot";
  created_at: string;
}

interface JobLeadsPanelProps {
  isOpen: boolean;
  onClose: () => void;
  selectedLeadTypes: ("Qualifying" | "Cold" | "Warm" | "Hot")[];
  onFilterChange: (types: string[]) => void;
}
```

Features:

- Sliding panel or modal (suggest sliding panel from right side)
- Sticky header with count, filter dropdown, and close button
- Filter dropdown shows checkboxes for: Qualifying, Cold, Warm, Hot, All
- Queries database for jobs where `lead_type IN (selected types)`
- Real-time updates when jobs are added/modified
- Scrollable list of job cards
- Each card shows company, title, URL, lead type badge, and three action buttons
- Lead type can be changed via dropdown on each card

#### 3. Job Lead Detection and Auto-Save Logic

Parse agent responses to detect job leads and save to database:

- Pattern: Company name + "- " + Job title + "- " + URL
- Example: `Ramp - Senior Software Engineer - https://ramp.com/careers/...`
- Use regex to extract structured data
- Automatically save each lead to database with `status='lead'`, `lead_type='Qualifying'`
- Trigger panel open when new jobs added
- Panel queries database to show all leads

```tsx
const jobLeadPattern = /^(.+?)\s*-\s*(.+?)\s*-\s*(https?:\/\/.+)$/gm;

const extractAndSaveJobLeads = async (messageContent: string) => {
  const matches = [...messageContent.matchAll(jobLeadPattern)];

  for (const match of matches) {
    const lead = {
      company: match[1].trim(),
      job_title: match[2].trim(),
      job_url: match[3].trim(),
      status: "lead" as const,
      lead_type: "Qualifying" as const,
      source: "agent" as const,
    };

    // Save to database via API
    await createJob(lead);
  }

  // Open panel to show newly added leads
  setLeadsPanelOpen(true);
};
```

#### 4. API Integration

Extend existing jobs API to support lead_type:

```typescript
// modules/client/lib/jobs-api.ts

// Fetch all leads (jobs with lead_type not null)
export async function fetchLeads(
  leadTypes?: ("Qualifying" | "Cold" | "Warm" | "Hot")[]
): Promise<AppliedJob[]> {
  const query = supabase.from("Job").select("*").not("lead_type", "is", null);

  if (leadTypes && leadTypes.length > 0) {
    query.in("lead_type", leadTypes);
  }

  return query.order("created_at", { ascending: false });
}

// Update lead type
export async function updateLeadType(
  jobId: string,
  leadType: "Qualifying" | "Cold" | "Warm" | "Hot"
): Promise<void> {
  await updateJob(jobId, { lead_type: leadType });
}

// Convert lead to application
export async function markLeadAsApplied(jobId: string): Promise<void> {
  await updateJob(jobId, {
    status: "applied",
    lead_type: null,
    date: new Date().toISOString(),
  });
}

// Mark lead as do not apply
export async function markLeadAsDoNotApply(jobId: string): Promise<void> {
  await updateJob(jobId, {
    status: "do_not_apply",
    lead_type: null,
  });
}
```

### Backend Changes

#### 1. New Agent Tool (`modules/agent/src/agent/tools/worker_tools.py`)

```python
@tool
def update_job_status(
    company: str,
    job_title: str,
    status: Literal["applied", "interviewing", "rejected", "offer", "accepted", "do_not_apply"],
    job_url: Optional[str] = None,
    notes: Optional[str] = None
) -> str:
    """
    Update the status of a job application. Use this when the user says they applied
    to a company, or wants to mark a job as 'do not apply'.

    Args:
        company: The company name (fuzzy matching supported)
        job_title: The job title
        status: The status to set
        job_url: Optional URL for the job posting
        notes: Optional notes about the application

    Returns:
        Confirmation message
    """
    from agent.services.job_service import update_job_by_company

    # Try to find existing job (including leads)
    job = update_job_by_company(
        company=company,
        job_title=job_title,
        status=status,
        lead_type=None,  # Clear lead_type when status changes
        job_url=job_url,
        notes=notes
    )

    if not job:
        # If not found, create new entry
        from agent.services.job_service import add_job
        job = add_job(
            company=company,
            job_title=job_title,
            status=status,
            job_url=job_url,
            notes=notes,
            source="agent"
        )

    return f"Successfully marked {company} - {job_title} as {status}"
```

#### 2. Update Job Service (`modules/agent/src/agent/services/job_service.py`)

Add support for lead_type and new status values:

```python
def add_job(
    company: str,
    job_title: str,
    status: str = "lead",  # Default to 'lead' for agent-added jobs
    lead_type: Optional[str] = "Qualifying",  # Default lead type
    job_url: Optional[str] = None,
    notes: Optional[str] = None,
    source: str = "agent"
) -> Dict:
    """Add a new job to the database"""
    # Implementation with lead_type support
    pass

def update_job_by_company(
    company: str,
    job_title: str,
    status: Optional[str] = None,
    lead_type: Optional[str] = None,
    job_url: Optional[str] = None,
    notes: Optional[str] = None
) -> Optional[Dict]:
    """
    Find and update a job by company name (fuzzy matching).
    Returns updated job or None if not found.
    """
    # Implementation with fuzzy matching
    pass

def get_all_leads(lead_types: Optional[List[str]] = None) -> List[Dict]:
    """
    Get all jobs where lead_type IS NOT NULL
    Optionally filter by specific lead types
    """
    pass
```

#### 3. Update Job Filtering Logic

Modify `check_applied_jobs` and `job_search` to filter out jobs with `status = 'do_not_apply'` but INCLUDE leads:

```python
# In check_applied_jobs tool
excluded_statuses = ['applied', 'interviewing', 'offer', 'accepted', 'do_not_apply']

# In job_search tool - filter out applied and do_not_apply, but keep leads visible
# This ensures the agent can still search for jobs even if they're in the leads panel
```

**Important**: Jobs with `status='lead'` should NOT be filtered out of search results, allowing users to see them in both the search results and leads panel.

---

## UX Considerations

1. **Job Leads Panel Position**:

   - Recommend side panel (right side) that slides in
   - Should not block chat interface
   - Should be dismissible but reopenable via button
   - Button in header: "Leads (5)" shows count

2. **Lead Type Badges**:

   - Qualifying: Blue/gray (default)
   - Cold: Light blue
   - Warm: Orange/yellow
   - Hot: Red/pink
   - Each job card shows lead type badge with dropdown to change

3. **Button Styling**:

   - "Applied": Green button (positive action)
   - "Remove": Gray button with icon only (neutral action)
   - "Do Not Apply": Red/orange button with X icon (negative action)

4. **Feedback**:

   - Show toast notifications on success
   - Show error messages if database operation fails
   - Update panel immediately (optimistic UI)
   - When job transitions, briefly highlight the row before removing

5. **Persistence**:

   - Panel reads from database, so naturally persists across refreshes
   - Panel open/closed state saved in localStorage
   - Filter selection saved in localStorage

6. **JobTracker Integration**:
   - "Do Not Apply" jobs visible in JobTracker with red/muted styling
   - Filter in JobTracker to hide/show "do_not_apply" status
   - "Lead" status jobs visible in JobTracker with blue styling
   - Clicking job in JobTracker with lead_type opens Leads Panel

---

## Testing Checklist

### Frontend Testing

- [ ] Export button creates valid .txt file with all messages
- [ ] Job leads panel opens automatically when agent sends jobs
- [ ] Agent-sent jobs are saved to database with `status='lead'`, `lead_type='Qualifying'`
- [ ] Panel queries and displays all jobs where `lead_type IS NOT NULL`
- [ ] Filter dropdown shows/hides leads based on selected lead types
- [ ] Lead type can be changed via dropdown on each card
- [ ] "Applied" button updates job: `status='applied'`, `lead_type=NULL`, removes from panel
- [ ] "Remove" button deletes job from database, removes from panel
- [ ] "Do Not Apply" button updates job: `status='do_not_apply'`, `lead_type=NULL`, removes from panel
- [ ] Panel state (open/closed, filters) persists across page refreshes
- [ ] Toast notifications appear on actions
- [ ] Jobs appear in JobTracker with correct status after actions
- [ ] "Do Not Apply" jobs show in JobTracker with distinct styling
- [ ] Mobile responsive design works

### Backend Testing

- [ ] `update_job_status` tool updates jobs correctly
- [ ] `update_job_status` clears `lead_type` when changing status
- [ ] Fuzzy matching works for company names
- [ ] `add_job()` accepts `lead_type` parameter
- [ ] `get_all_leads()` returns only jobs with non-null `lead_type`
- [ ] `check_applied_jobs` filters out 'do_not_apply' and 'applied' jobs
- [ ] `check_applied_jobs` INCLUDES jobs with status='lead'
- [ ] `job_search` excludes 'do_not_apply' jobs from results
- [ ] Agent responds appropriately to "I applied to X" messages
- [ ] Agent responds appropriately to "Don't show me X" messages
- [ ] Database migration runs successfully on both local Postgres and Supabase

### Integration Testing

- [ ] End-to-end: User asks for jobs → agent returns jobs → jobs saved to DB with `status='lead'` → leads panel opens → user sees jobs
- [ ] End-to-end: User filters to "Hot" leads only → panel shows only Hot leads
- [ ] End-to-end: User changes lead from "Qualifying" to "Hot" → updates in database → still shows in panel
- [ ] End-to-end: User clicks "Applied" → job updates in database → removed from leads panel → appears in JobTracker with "applied" status
- [ ] End-to-end: User clicks "Do Not Apply" → updates in database → removed from panel → appears in JobTracker with distinct styling
- [ ] End-to-end: User says "I applied to Alto Neuroscience" → agent updates existing lead → confirms in chat → job removed from leads panel
- [ ] End-to-end: User asks for jobs again → agent filters out "do_not_apply" jobs → doesn't show previously dismissed jobs
- [ ] End-to-end: User refreshes page → leads panel state and filters persist → shows same jobs

---

## Dependencies

- No new npm packages required
- No new Python packages required
- Requires database migration for new status value

---

## Estimated Effort

- Frontend (Chat Export): 2 hours
- Frontend (Job Leads Panel with filtering): 10 hours
- Frontend (Auto-save jobs from agent responses): 3 hours
- Frontend (Lead type management): 3 hours
- Frontend (JobTracker styling for new statuses): 2 hours
- Backend (Update tool and service functions): 4 hours
- Backend (Lead type support): 2 hours
- Backend (Filtering logic updates): 2 hours
- Database Migration: 2 hours
- Testing: 6 hours

**Total**: ~36 hours (4-5 days)

---

## Open Questions

1. ~~Should "Do Not Apply" jobs be visible in JobTracker, or completely hidden?~~

   - **ANSWERED**: Yes, visible with distinct styling and filterable

2. ~~Should job leads panel persist across page refreshes?~~

   - **ANSWERED**: Yes, by storing leads in database with `lead_type` column

3. ~~Should we parse existing chat history to populate leads panel retroactively?~~

   - **ANSWERED**: No, only new messages after feature is deployed

4. What if agent sends job leads in a different format?

   - **Recommendation**: Add format detection logic, fallback to current behavior

5. When agent returns jobs that already exist as leads, should we:
   - Skip duplicates entirely?
   - Update existing leads with new information?
   - **Recommendation**: Check for duplicates by company+title before inserting

---

## Success Metrics

- Reduction in time to track job applications (target: 50% faster)
- Increase in number of jobs tracked per user session
- Reduction in duplicate job searches (from filtering do_not_apply)
- User feedback via NPS or satisfaction survey

---

## Screenshots/Mockups

_To be added by designer_

Suggested layout:

```
+----------------------------------+     +---------------------------+
|  Chat Interface                  |     |  Job Leads Panel          |
|                                  |     |  [X Close]                |
|  User: startups in nyc...        |     |                           |
|  Agent: Here are some...         |     |  Found Jobs (5)           |
|  - Ramp - Senior Software...     |     |                           |
|                                  |     |  Ramp                     |
|  [Export Chat]                   |     |  Senior Software Engineer |
|                                  |     |  https://ramp.com/...     |
|                                  |     |  [Applied] [Remove] [X]   |
|                                  |     |                           |
|                                  |     |  CoreWeave                |
|                                  |     |  Senior Software Eng...   |
|                                  |     |  https://coreweave.com... |
|                                  |     |  [Applied] [Remove] [X]   |
+----------------------------------+     +---------------------------+
```

---

## Related Tickets

- None (new feature)

## Labels

`enhancement`, `job-search`, `frontend`, `backend`, `database`, `ux`
