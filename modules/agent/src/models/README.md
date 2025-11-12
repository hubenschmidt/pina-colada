# DealTracker Models

SQLAlchemy ORM models for the DealTracker schema with functional TypedDict interfaces.

## Model Architecture

All models follow this pattern:
- **SQLAlchemy ORM class** - For database persistence
- **TypedDict data models** - For functional business logic
  - `{Model}Data` - Complete model data
  - `{Model}CreateData` - Fields for creation
  - `{Model}UpdateData` - Optional fields for updates
- **Conversion functions**:
  - `orm_to_dict()` - Convert ORM to dict
  - `dict_to_orm()` - Convert dict to ORM
  - `update_orm_from_dict()` - Update existing ORM from dict

## Core Models

### Status
Central status/stage definitions for all trackable entities.
- **Categories**: `job`, `lead`, `deal`, `task_status`, `task_priority`
- **Fields**: name, description, category, is_terminal
- **Primary key**: BIGINT (auto-increment)

### Organization
Companies and organizations.
- **Fields**: name, website, phone, industry, employee_count, description, notes
- **Unique constraint**: Lowercase name
- **Relationships**: jobs, opportunities, partnerships, contacts

### Individual
People (separate from their company affiliations).
- **Fields**: first_name, last_name, email, phone, linkedin_url, title, notes
- **Unique constraint**: Lowercase email
- **Relationships**: contacts

### Contact
Links an Individual to an Organization with role information.
- **Fields**: individual_id, organization_id, title, department, role, email, phone, is_primary, notes
- **Relationships**: individual, organization

### Deal
High-level deals that contain multiple leads.
- **Fields**: name, description, owner_user_id, current_status_id, value_amount, value_currency, probability, expected_close_date
- **Relationships**: current_status, leads

### Lead
Base table for Joined Table Inheritance (Job, Opportunity, Partnership extend this).
- **Fields**: deal_id, type (discriminator), title, description, source, current_status_id, owner_user_id
- **Type values**: `'Job'`, `'Opportunity'`, `'Partnership'`
- **Relationships**: deal, current_status, job, opportunity, partnership

## Lead Subtypes (Joined Table Inheritance)

### Job (extends Lead)
Job applications.
- **Primary key**: Foreign key to Lead.id
- **Fields**: organization_id, job_title, job_url, notes, resume_date, salary_range
- **Relationships**: lead (parent), organization
- **Special notes**: When querying/creating, need to handle both Lead and Job tables

### Opportunity (extends Lead)
Business opportunities.
- **Primary key**: Foreign key to Lead.id
- **Fields**: organization_id, opportunity_name, estimated_value, probability, expected_close_date, notes
- **Relationships**: lead (parent), organization

### Partnership (extends Lead)
Partnership opportunities.
- **Primary key**: Foreign key to Lead.id
- **Fields**: organization_id, partnership_type, partnership_name, start_date, end_date, notes
- **Relationships**: lead (parent), organization

## Migration from Old Schema

### Key Changes

**Old Job Model (UUID-based)**:
```python
id: UUID
company: str
job_title: str
status: str  # 'applied', 'interviewing', etc.
lead_status_id: UUID -> LeadStatus
```

**New Job Model (BIGINT, extends Lead)**:
```python
id: BIGINT (FK to Lead.id)
organization_id: BIGINT (FK to Organization.id)
job_title: str
# Status now through Lead.current_status_id -> Status
# Company now through organization relationship
```

### Important Notes for Job Creation

When creating a Job in the new schema, you must:

1. **Create/get Organization** first
2. **Create Lead** record with:
   - `deal_id` (required)
   - `type='Job'` (discriminator)
   - `title` (e.g., "Company - Job Title")
   - `current_status_id` (FK to Status, not the old status string)
3. **Create Job** record with:
   - `id` = the Lead.id you just created
   - `organization_id`
   - `job_title`
   - Other job-specific fields

### Status Mapping

**Old** → **New**:
- `'applied'` → Status with name='Applied', category='job'
- `'interviewing'` → Status with name='Interviewing', category='job'
- `'rejected'` → Status with name='Rejected', category='job'
- `'offer'` → Status with name='Offer', category='job'
- `'accepted'` → Status with name='Accepted', category='job'

## Usage Examples

### Creating a Job (Atomic)

```python
from models import Job, Lead, Organization, Status

# 1. Get or create organization
org = session.query(Organization).filter_by(name="Acme Corp").first()

# 2. Get status
status = session.query(Status).filter_by(name="Applied", category="job").first()

# 3. Create Lead first
lead = Lead(
    deal_id=default_deal_id,
    type="Job",
    title="Acme Corp - Senior Engineer",
    source="manual",
    current_status_id=status.id
)
session.add(lead)
session.flush()  # Get the lead.id

# 4. Create Job with lead.id
job = Job(
    id=lead.id,
    organization_id=org.id,
    job_title="Senior Engineer",
    job_url="https://...",
    salary_range="150-200K"
)
session.add(job)
session.commit()
```

### Querying Jobs with Lead Data

```python
# Join with Lead to get status information
jobs = session.query(Job).join(Lead).join(Status).all()

for job in jobs:
    data = orm_to_dict(job)  # Includes flattened lead fields
    print(f"{data['title']} - {data['current_status']['name']}")
```

## Database Schema Compatibility

These models are compatible with the schema defined in:
- `modules/agent/migrations/002_dealtracker.sql`

Ensure the database migration is run before using these models.
