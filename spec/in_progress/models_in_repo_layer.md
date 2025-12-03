# Feature: Pydantic BaseModels in Repository Layer

## Overview
Relocate all Pydantic BaseModel definitions from route files to their corresponding repository files, ensuring data validation schemas live alongside data access logic.

## User Story
As a developer, I want Pydantic models defined in the repository layer so that data validation schemas are colocated with data access logic.

---

## Verification Checklist

### Functional Requirements
- [x] All BaseModels removed from route files
- [x] All BaseModels added to corresponding repository files
- [x] Routes import BaseModels from repositories
- [x] All files compile without errors
- [ ] API endpoints function correctly (manual testing required)

### Non-Functional Requirements
- [x] No circular import issues
- [x] Consistent import patterns across all routes

---

## Implementation Notes

### Repositories Updated (18 files):
- `individual_repository.py` - IndividualCreate, IndividualUpdate, IndContactCreate, IndContactUpdate
- `organization_repository.py` - OrganizationCreate, OrganizationUpdate, OrgContactCreate, OrgContactUpdate, OrgTechnologyCreate, FundingRoundCreate, SignalCreate
- `contact_repository.py` - ContactCreate, ContactUpdate
- `task_repository.py` - TaskCreate, TaskUpdate
- `project_repository.py` (new) - ProjectCreate, ProjectUpdate
- `note_repository.py` - NoteCreate, NoteUpdate
- `document_repository.py` - DocumentUpdate, EntityLink
- `comment_repository.py` - CommentCreate, CommentUpdate
- `opportunity_repository.py` - OpportunityCreate, OpportunityUpdate
- `notification_repository.py` - MarkReadRequest, MarkEntityReadRequest
- `data_provenance_repository.py` - ProvenanceCreate
- `industry_repository.py` - IndustryCreate
- `technology_repository.py` - TechnologyCreate
- `tenant_repository.py` - TenantCreate
- `job_repository.py` - JobCreate, JobUpdate
- `partnership_repository.py` - PartnershipCreate, PartnershipUpdate
- `preferences_repository.py` (new) - UpdateUserPreferencesRequest, UpdateTenantPreferencesRequest
- `saved_report_repository.py` - ReportFilter, Aggregation, ReportQueryRequest, SavedReportCreate, SavedReportUpdate

### Routes Updated (19 files):
- `individuals.py`, `organizations.py`, `contacts.py`, `tasks.py`, `projects.py`
- `notes.py`, `documents.py`, `comments.py`, `opportunities.py`, `notifications.py`
- `provenance.py`, `industries.py`, `technologies.py`
- `auth.py`, `jobs.py`, `partnerships.py`, `preferences.py`, `reports.py`

### Out of Scope
- `api/routes/mocks/k401_rollover.py` - Mock endpoint for demo, isolated code
