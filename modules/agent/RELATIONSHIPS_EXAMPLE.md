# SQLAlchemy Relationships - Usage Examples

SQLAlchemy is excellent for relationships, similar to Sequelize in TypeScript. Here are examples:

## Basic Relationship Usage

```python
from sqlalchemy.orm import Session, joinedload
from agent.models.Job import Job
from agent.models.LeadStatus import LeadStatus

# Get a job with its lead status (eager loading)
job = session.query(Job).options(joinedload(Job.lead_status)).filter(Job.id == job_id).first()
print(job.lead_status.name)  # Access related LeadStatus

# Get all jobs for a specific lead status
lead_status = session.query(LeadStatus).filter(LeadStatus.name == "Hot").first()
hot_jobs = lead_status.jobs  # Access related Jobs via relationship

# Create a job with a lead status
hot_lead = session.query(LeadStatus).filter(LeadStatus.name == "Hot").first()
new_job = Job(
    company="Acme Corp",
    job_title="Senior Engineer",
    lead_status_id=hot_lead.id
)
session.add(new_job)
session.commit()

# Or set via relationship
new_job.lead_status = hot_lead
session.commit()
```

## Query Examples

```python
# Find all jobs with "Hot" lead status
hot_jobs = session.query(Job).join(LeadStatus).filter(LeadStatus.name == "Hot").all()

# Count jobs by lead status
from sqlalchemy import func
results = session.query(
    LeadStatus.name,
    func.count(Job.id).label('job_count')
).join(Job).group_by(LeadStatus.name).all()

# Get lead statuses that have no jobs
empty_lead_statuses = session.query(LeadStatus).outerjoin(Job).filter(Job.id == None).all()
```

## Adding a New Lead Status

```python
# Just insert into the table - no schema changes needed!
new_status = LeadStatus(
    name="Premium",
    description="Premium leads with high conversion potential"
)
session.add(new_status)
session.commit()
```

## Comparison to Sequelize

SQLAlchemy relationships work very similarly to Sequelize:

**Sequelize (TypeScript):**
```typescript
const job = await Job.findByPk(id, { include: [LeadStatus] });
const hotJobs = await leadStatus.getJobs();
```

**SQLAlchemy (Python):**
```python
job = session.query(Job).options(joinedload(Job.lead_status)).filter(Job.id == id).first()
hot_jobs = lead_status.jobs
```

Both support:
- Eager loading (joinedload)
- Lazy loading (default)
- Back references (back_populates)
- Foreign keys
- Cascading operations

