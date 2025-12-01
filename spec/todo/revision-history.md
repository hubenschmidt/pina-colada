# Feature: Revision History

## Overview
Track revisions to database entities (starting with Job, Lead, Status) to enable AI Agent analysis of historical changes. Uses a hybrid storage approach: metadata in Postgres, full snapshots in blob storage (MinIO locally, Cloudflare R2 in test/prod).

## User Story
As an AI Agent, I want access to revision history of Jobs and their status changes so that I can analyze patterns, detect anomalies, and provide insights to users.

---

## Architecture

### Storage Strategy (Per Environment)

| Environment | Metadata | Snapshots | Bucket/Endpoint |
|-------------|----------|-----------|-----------------|
| Local dev | Postgres (local) | MinIO (existing docker-compose) | `minio:9000/revisions` |
| Test | Postgres (Supabase) | Cloudflare R2 | `revisions-test` bucket |
| Prod | Postgres (Supabase) | Cloudflare R2 | `revisions-prod` bucket |

### Schema: `Revision` table (Postgres)

```sql
CREATE TABLE revision (
    id              BIGSERIAL PRIMARY KEY,
    entity_type     TEXT NOT NULL,          -- 'Job', 'Lead', 'Status'
    entity_id       BIGINT NOT NULL,        -- FK to the entity
    revision_number INTEGER NOT NULL,       -- Sequential per entity
    blob_key        TEXT NOT NULL,          -- S3 key: '{entity_type}/{entity_id}/{revision_number}.json'
    changed_fields  TEXT[],                 -- Which fields triggered this revision
    actor_id        BIGINT,                 -- FK to Individual (nullable for system)
    source          TEXT,                   -- 'manual', 'agent', 'api', 'system'
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_revision_entity ON revision(entity_type, entity_id, revision_number);
CREATE INDEX idx_revision_created ON revision(entity_type, created_at);
```

### Blob Storage

- **Key format**: `{entity_type}/{entity_id}/{revision_number}.json`
- **Content**: Full JSON snapshot of the entity at that revision
- **API**: S3-compatible (boto3) — same code for MinIO and R2

### Storage Abstraction

```python
class RevisionStorage(Protocol):
    def put(self, key: str, data: dict) -> None: ...
    def get(self, key: str) -> dict: ...

# Configured via env vars:
# BLOB_ENDPOINT, BLOB_ACCESS_KEY, BLOB_SECRET_KEY, BLOB_BUCKET
```

---

## Scenarios

### Scenario 1: Job Status Change Creates Revision

**Given** a Job with `current_status_id` pointing to "Applied"
**When** the status is changed to "Interviewing" via the API
**Then** a new revision record is created with:
- `entity_type`: "Lead" (since status is on Lead)
- `entity_id`: the Lead's ID
- `changed_fields`: ["current_status_id"]
- `source`: "api"
- Full snapshot uploaded to blob storage

### Scenario 2: Query Revision History for a Job

**Given** a Job with 5 revisions
**When** the AI Agent queries revisions for that Job
**Then** it receives metadata for all 5 revisions from Postgres
**And** can fetch full snapshots from blob storage as needed

### Scenario 3: AI Agent Point-in-Time Comparison

**Given** a Job with 5 revisions
**When** the AI Agent compares snapshot #1 vs #5
**Then** it can identify what changed between those two points without loading intermediate snapshots

### Scenario 4: AI Agent Full Replay Analysis

**Given** a Job with 5 revisions
**When** the AI Agent replays snapshots #1 → #2 → #3 → #4 → #5 chronologically
**Then** it can analyze mutation patterns like:
- State transitions (Applied → Interview → Offer vs Applied → Rejected → Re-applied)
- Time-in-state metrics ("sat in Interview for 45 days")
- Yo-yo detection (status bouncing back and forth)
- Field evolution (how `salary_range` changed across negotiations)
- Actor patterns (manual changes vs agent changes)

---

## Analysis Architecture

### Core Analysis Layer (Deterministic)

Python functions that compute structured results from revision data:

| Analysis Type | Approach |
|---------------|----------|
| State transitions | Sequence/graph traversal of status changes |
| Time-in-state metrics | Timestamp arithmetic between revisions |
| Yo-yo detection | Pattern matching on status sequences |
| Field evolution | JSON diff comparison across snapshots |
| Actor patterns | Aggregation queries on `actor_id`/`source` |

These functions are fast, cheap, testable, and produce deterministic outputs with no hallucination risk.

### Presentation Layer (LLM)

The AI Agent consumes structured analysis results and:
- Summarizes insights in natural language for users
- Interprets patterns contextually (e.g., "this looks like a negotiation pattern")
- Handles open-ended queries (e.g., "what's unusual about this job hunt?")

The LLM does not perform the core computations—it explains them.

---

## Configuration

### Cascading Behavior

**No cascading** — each entity tracks only its own field changes:
- Job status change (`Lead.current_status_id`) → revision on Lead only
- Status definition change (`Status.name`) → revision on Status only
- AI Agent reconstructs relationships by joining revision data across entities

### Tracked Fields (Per Entity)

Define which field changes trigger a revision:

```python
TRACKED_FIELDS = {
    "Lead": ["current_status_id", "title", "description", "source"],
    "Job": ["job_title", "job_url", "salary_range"],
    "Status": ["name", "is_terminal", "category"],
}
```

### MinIO Console (Local Dev)

- **URL**: http://localhost:9091
- **Username**: `minio`
- **Password**: `miniosecret`

### Environment Variables

```bash
# Local dev (MinIO)
BLOB_ENDPOINT=http://minio:9000
BLOB_ACCESS_KEY=minio
BLOB_SECRET_KEY=miniosecret
BLOB_BUCKET=revisions

# Test (R2)
BLOB_ENDPOINT=https://<account>.r2.cloudflarestorage.com
BLOB_ACCESS_KEY=<r2_access_key>
BLOB_SECRET_KEY=<r2_secret_key>
BLOB_BUCKET=revisions-test

# Prod (R2)
BLOB_BUCKET=revisions-prod
```

---

## Verification Checklist

### Functional Requirements
- [ ] Revision created when tracked field changes
- [ ] Revision NOT created when untracked field changes
- [ ] Full snapshot stored in blob storage
- [ ] Metadata stored in Postgres with correct indexes
- [ ] Actor ID and source captured correctly
- [ ] Revision numbers are sequential per entity

### Non-Functional Requirements
- [ ] Blob upload does not block the main transaction
- [ ] Graceful handling if blob storage is unavailable
- [ ] Works with existing MinIO in docker-compose (local)
- [ ] Works with R2 (test/prod)

### Edge Cases
- [ ] First revision for an entity (revision_number = 1)
- [ ] Multiple fields change in single update
- [ ] Entity deleted — revisions remain for historical analysis
- [ ] Null actor_id for system-initiated changes

---

## Implementation Notes

### Estimate of Scope
- New `Revision` model and migration
- Storage abstraction (S3-compatible client)
- SQLAlchemy event listener or service layer hook
- Environment-specific configuration

### Files to Modify
- `modules/agent/src/models/` — Add `Revision.py`
- `modules/agent/src/services/` — Add `revision_service.py`
- `modules/agent/src/config/` — Add blob storage config
- `docker-compose.yml` — Update MinIO command to auto-create `revisions` bucket:
  ```yaml
  command: -c 'mkdir -p /data/langfuse /data/revisions && minio server --address ":9000" --console-address ":9001" /data'
  ```

### Dependencies
- `boto3` — S3-compatible client (already likely installed)
- Existing MinIO service (docker-compose)
- Cloudflare R2 account (for test/prod)

### Out of Scope
- UI for viewing revision history
- Retention/cleanup policies (can add later)
- Compression of snapshots
- Real-time streaming of revisions

---

## Go Analytics Microservice (Future)

### Rationale
A dedicated Go service for snapshot analysis enables:
- High-performance diff computation across large JSON snapshots
- CPU-bound analysis without blocking Python async event loop
- Skill development opportunity with production Go codebase

### Scope
**Initial capability**: Snapshot diff computation
- Input: Two blob keys (or raw JSON payloads)
- Output: Structured diff (added/removed/changed fields)

### Architecture

```
modules/analytics/
├── Dockerfile
├── go.mod
├── main.go
├── internal/
│   ├── api/          # HTTP handlers
│   ├── diff/         # JSON diff logic
│   └── storage/      # S3-compatible client (MinIO/R2)
└── config/
```

### Integration
- Python agent calls Go service via HTTP when analysis needed
- Go service reads snapshots directly from MinIO/R2
- Returns structured JSON (not natural language)

### Docker Compose Addition

```yaml
analytics:
  build:
    context: ./modules/analytics
    dockerfile: Dockerfile
  ports:
    - "9002:9002"
  environment:
    - BLOB_ENDPOINT=${BLOB_ENDPOINT:-http://minio:9000}
    - BLOB_ACCESS_KEY=${BLOB_ACCESS_KEY:-minio}
    - BLOB_SECRET_KEY=${BLOB_SECRET_KEY:-miniosecret}
    - BLOB_BUCKET=${BLOB_BUCKET:-revisions}
  depends_on:
    minio:
      condition: service_healthy
  networks:
    - app-network
```

### API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/health` | Health check |
| POST | `/diff` | Compare two snapshots |

### Example Request

```json
POST /diff
{
  "left": "Lead/123/1.json",
  "right": "Lead/123/5.json"
}
```

### Example Response

```json
{
  "added": [],
  "removed": [],
  "changed": [
    {
      "path": "current_status_id",
      "from": 1,
      "to": 3
    },
    {
      "path": "salary_range.max",
      "from": 120000,
      "to": 145000
    }
  ]
}
```

### Dependencies
- Go 1.22+
- `github.com/wI2L/jsondiff` or similar
- `github.com/aws/aws-sdk-go-v2` for S3

### Implementation Phases
1. Scaffold Go service with health endpoint
2. Add S3 client for MinIO/R2
3. Implement `/diff` endpoint
4. Integrate with Python agent (HTTP client)
