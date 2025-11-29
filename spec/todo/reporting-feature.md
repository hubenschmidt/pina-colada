# Feature: Reports

## Overview
A reporting module that provides both pre-built (canned) reports for common business insights and a custom report builder for ad-hoc analysis. Accessible via a new "Reports" section in the sidebar with sub-tabs for Canned and Custom reports.

## User Stories

**Canned Reports**
As a sales manager, I want pre-built reports on leads, accounts, and contacts so that I can make mission-critical decisions quickly without building queries from scratch.

**Custom Reports**
As a power user, I want to build custom reports by selecting entities, columns, and filters so that I can analyze data in ways not covered by canned reports, save my configurations, and export to Excel.

---

## Scenarios

### Scenario 1: View Lead Pipeline Report (Canned)

**Given** User navigates to Reports > Canned > Lead Pipeline
**When** The report loads
**Then** User sees:
- Lead counts by status (bar chart optional, table required)
- Conversion rate metrics (leads progressed vs total)
- Lead source breakdown (which sources generate leads)
- Date range filter (default: last 30 days)

### Scenario 2: View Account Overview Report (Canned)

**Given** User navigates to Reports > Canned > Account Overview
**When** The report loads
**Then** User sees:
- Total Organizations and Individuals counts
- Industry breakdown (count per industry)
- Geographic distribution (by headquarters_country/state)
- Company type distribution
- Revenue range distribution

### Scenario 3: View Contact Coverage Report (Canned)

**Given** User navigates to Reports > Canned > Contact Coverage
**When** The report loads
**Then** User sees:
- Average contacts per organization
- Organizations with zero contacts (coverage gaps)
- Decision-maker ratio (contacts marked is_decision_maker / total)
- Contact density heatmap or table by org

### Scenario 4: Build a Custom Report

**Given** User navigates to Reports > Custom
**When** User clicks "New Report"
**Then** User sees a report builder with:
1. **Entity Selector**: Primary entity (Organizations, Individuals, Contacts, Leads)
2. **Column Picker**: Available fields from selected entity + joinable entities
3. **Joins Panel**: Optional joins (e.g., Orgs with their Contacts, Individuals with their Organizations)
4. **Filters Panel**: Field-based filters with operators (equals, contains, >, <, between, is null)
5. **Aggregations**: Group by + aggregate functions (COUNT, SUM, AVG, MIN, MAX)
6. **Preview**: Live preview of results (limited rows)
7. **Actions**: Run full query, Export to Excel, Save Report

### Scenario 5: Save and Load Custom Report

**Given** User has built a custom report
**When** User clicks "Save Report" and provides a name
**Then** The report definition is saved and appears in the Custom reports list

**Given** User navigates to Reports > Custom
**When** User selects a previously saved report
**Then** The builder loads with the saved configuration and user can run/modify/export

### Scenario 6: Export Report to Excel

**Given** User is viewing any report (canned or custom)
**When** User clicks "Export to Excel"
**Then** An .xlsx file downloads containing the current report data

---

## API Design

### Endpoints

```
# Canned Reports
GET  /api/reports/canned/lead-pipeline?date_from=&date_to=
GET  /api/reports/canned/account-overview
GET  /api/reports/canned/contact-coverage

# Custom Reports - Execution
POST /api/reports/custom/preview    # Returns limited rows for preview
POST /api/reports/custom/run        # Returns full result set
POST /api/reports/custom/export     # Returns Excel file

# Saved Reports - CRUD
GET    /api/reports/saved           # List all saved reports for tenant
POST   /api/reports/saved           # Create new saved report
GET    /api/reports/saved/{id}      # Get saved report definition
PUT    /api/reports/saved/{id}      # Update saved report
DELETE /api/reports/saved/{id}      # Delete saved report
```

### Request/Response Models

```python
# Custom Report Query Request
class ReportQueryRequest(BaseModel):
    primary_entity: Literal["organizations", "individuals", "contacts", "leads"]
    columns: list[str]  # e.g., ["name", "email", "organization.name"]
    joins: list[str] | None  # e.g., ["contacts", "organizations"]
    filters: list[ReportFilter] | None
    group_by: list[str] | None
    aggregations: list[Aggregation] | None
    limit: int | None = 100

class ReportFilter(BaseModel):
    field: str
    operator: Literal["eq", "neq", "gt", "gte", "lt", "lte", "contains", "starts_with", "is_null", "is_not_null", "between", "in"]
    value: Any

class Aggregation(BaseModel):
    function: Literal["count", "sum", "avg", "min", "max"]
    field: str
    alias: str

# Saved Report Model
class SavedReport(BaseModel):
    id: int
    name: str
    description: str | None
    query: ReportQueryRequest
    created_at: datetime
    updated_at: datetime
```

### Database Model

```sql
-- New table for saved reports
CREATE TABLE saved_reports (
    id SERIAL PRIMARY KEY,
    tenant_id INTEGER NOT NULL REFERENCES tenants(id),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    query_definition JSONB NOT NULL,  -- Stores ReportQueryRequest as JSON
    created_by INTEGER REFERENCES individuals(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_saved_reports_tenant ON saved_reports(tenant_id);
```

---

## Frontend Design

### Routing Structure

```
/reports
├── /canned
│   ├── /lead-pipeline
│   ├── /account-overview
│   └── /contact-coverage
└── /custom
    ├── /new              # New report builder
    └── /[id]             # Edit saved report
```

### Sidebar Update

Add to `Sidebar.tsx`:
```
Reports (collapsible)
├── Canned
│   ├── Lead Pipeline
│   ├── Account Overview
│   └── Contact Coverage
└── Custom
```

### Components

| Component | Purpose |
|-----------|---------|
| `ReportLayout` | Shared layout for all report pages |
| `CannedReportPage` | Template for canned reports with filters + data table |
| `ReportBuilder` | Interactive builder for custom reports |
| `EntitySelector` | Dropdown for primary entity selection |
| `ColumnPicker` | Multi-select for available columns |
| `JoinPanel` | Toggle joins to related entities |
| `FilterBuilder` | Add/remove filter conditions |
| `AggregationPanel` | Configure group by and aggregations |
| `ReportPreview` | Live preview data table |
| `SaveReportModal` | Modal for naming/saving report |
| `ExportButton` | Triggers Excel download |

---

## Verification Checklist

### Functional Requirements
- [ ] Sidebar shows Reports section with Canned and Custom sub-tabs
- [ ] Lead Pipeline report displays lead counts by status, conversion rates, source breakdown
- [ ] Account Overview report shows org/individual counts, industry/geo distribution
- [ ] Contact Coverage report shows contacts per org, decision-maker ratio, coverage gaps
- [ ] Custom report builder allows entity selection
- [ ] Custom report builder allows column selection with join fields
- [ ] Custom report builder supports filter conditions with multiple operators
- [ ] Custom report builder supports group by and aggregations
- [ ] Reports can be saved with a name
- [ ] Saved reports can be loaded and modified
- [ ] All reports export to Excel (.xlsx)

### Non-Functional Requirements
- [ ] Performance: Canned reports load in < 2s for typical data volumes
- [ ] Performance: Custom report preview returns in < 3s (limited to 100 rows)
- [ ] Security: Reports scoped to tenant_id (no cross-tenant data leakage)
- [ ] UX: Builder provides immediate feedback on invalid configurations

### Edge Cases
- [ ] Empty result sets display appropriate "No data" message
- [ ] Large exports (>10k rows) show progress indicator
- [ ] Invalid saved report definitions fail gracefully with error message
- [ ] Deleted entities referenced in saved reports handled gracefully

---

## Implementation Notes

### Estimate of Scope
Medium-Large feature. Recommend phased implementation:
1. Phase 1: Sidebar + Canned reports (3 endpoints, 3 pages)
2. Phase 2: Custom report builder + preview/run
3. Phase 3: Save/load + Excel export

### Files to Modify

**Backend**
- `modules/agent/src/api/routes/__init__.py` - Register new routes
- `modules/agent/src/api/routes/reports.py` - New file for all report endpoints
- `modules/agent/src/models/SavedReport.py` - New model
- `modules/agent/src/models/__init__.py` - Export new model
- `modules/agent/src/repositories/saved_report_repository.py` - New repository
- `modules/agent/src/services/report_builder.py` - Dynamic query builder service
- `modules/agent/migrations/0XX_saved_reports.sql` - New migration

**Frontend**
- `modules/client/components/Sidebar/Sidebar.tsx` - Add Reports section
- `modules/client/app/reports/layout.tsx` - New layout
- `modules/client/app/reports/canned/lead-pipeline/page.tsx` - New page
- `modules/client/app/reports/canned/account-overview/page.tsx` - New page
- `modules/client/app/reports/canned/contact-coverage/page.tsx` - New page
- `modules/client/app/reports/custom/page.tsx` - Saved reports list
- `modules/client/app/reports/custom/new/page.tsx` - Builder page
- `modules/client/app/reports/custom/[id]/page.tsx` - Edit saved report
- `modules/client/components/Reports/` - New component directory
- `modules/client/api/index.ts` - Add report API methods

### Dependencies
- Excel export: `openpyxl` (Python) for server-side .xlsx generation
- No new frontend dependencies expected (Mantine covers UI needs)

### Out of Scope
- Scheduled/automated report generation
- Charts/visualizations (tables only for initial release)
- Report sharing between users
- Email delivery of reports
