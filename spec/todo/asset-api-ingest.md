# Asset API Ingest - Technical Specification

## Context

Phase 2 of asset management focuses on programmatic/bulk asset ingestion. This enables:
- ETL pipelines pushing scraped data
- Third-party integrations
- Bulk imports from external systems
- Automated data collection workflows

**Depends on:** Phase 1 (`spec/asset-upload.md`) for core Asset/Tag schema.

**Goal:** Efficient bulk ingestion with minimal overhead, maintaining tag-based queryability for agent context.

---

## Architecture Overview

### Ingestion Patterns

1. **Single Asset API** - REST endpoint for one asset at a time
2. **Bulk Import API** - Batch endpoint for multiple assets
3. **Webhook Receiver** - Accept pushes from external systems
4. **Background Jobs** - Queue-based processing for large imports

### Authentication

- API keys for service-to-service auth
- Scoped to tenant with configurable permissions
- Rate limiting per key

---

## API Specifications

### 1. Single Asset Ingest

```python
@router.post("/api/v1/assets")
async def create_asset(
    request: AssetCreateRequest,
    api_key: str = Depends(verify_api_key),
):
    """
    Create single asset via API.

    Request body (JSON):
    {
      "filename": "job-posting-123.txt",
      "content_type": "text/plain",
      "content": "base64-encoded-content",
      "tags": ["job-posting", "software-engineer"],
      "description": "Scraped from LinkedIn",
      "metadata": {
        "source_url": "https://linkedin.com/jobs/123",
        "scraped_at": "2024-01-15T10:30:00Z"
      }
    }
    """
```

**Response:**
```json
{
  "id": 456,
  "filename": "job-posting-123.txt",
  "tags": ["job-posting", "software-engineer"],
  "created_at": "2024-01-15T10:31:00Z"
}
```

---

### 2. Bulk Import API

```python
@router.post("/api/v1/assets/bulk")
async def bulk_create_assets(
    request: BulkAssetCreateRequest,
    api_key: str = Depends(verify_api_key),
):
    """
    Create multiple assets in single request.

    Request body:
    {
      "assets": [
        {
          "filename": "job-1.txt",
          "content": "base64...",
          "tags": ["job-posting"]
        },
        {
          "filename": "job-2.txt",
          "content": "base64...",
          "tags": ["job-posting"]
        }
      ],
      "default_tags": ["imported", "2024-01"],
      "on_duplicate": "skip" | "replace" | "error"
    }
    """
```

**Response:**
```json
{
  "created": 45,
  "skipped": 5,
  "errors": [
    {"index": 12, "error": "File too large"}
  ]
}
```

**Constraints:**
- Max 100 assets per request
- Max 50MB total payload
- Processed in transaction (all or nothing, or partial with errors)

---

### 3. Webhook Receiver

```python
@router.post("/api/v1/webhooks/assets")
async def receive_asset_webhook(
    request: Request,
    x_webhook_secret: str = Header(...),
):
    """
    Receive asset data from external systems.

    Validates webhook signature, queues for processing.
    """
```

**Supported webhook sources:**
- Custom integrations (generic JSON payload)
- Scrapy/web scraping pipelines
- Zapier/Make.com

---

### 4. Import Job API

For very large imports (1000+ assets):

```python
@router.post("/api/v1/import-jobs")
async def create_import_job(
    file: UploadFile,  # CSV or JSON Lines
    api_key: str = Depends(verify_api_key),
):
    """
    Create background import job.

    Upload CSV/JSONL file, returns job ID for status polling.
    """

@router.get("/api/v1/import-jobs/{job_id}")
async def get_import_job_status(
    job_id: str,
    api_key: str = Depends(verify_api_key),
):
    """
    Poll import job status.

    Returns: pending, processing, completed, failed
    """
```

---

## Database Additions

### API Key Table

```sql
CREATE TABLE "ApiKey" (
  id BIGSERIAL PRIMARY KEY,
  tenant_id BIGINT NOT NULL REFERENCES "Tenant"(id) ON DELETE CASCADE,
  key_hash TEXT NOT NULL,  -- bcrypt hash
  name TEXT NOT NULL,
  permissions JSONB DEFAULT '["assets:write"]',
  rate_limit_per_hour INT DEFAULT 1000,
  last_used_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ DEFAULT now(),
  expires_at TIMESTAMPTZ
);
CREATE UNIQUE INDEX idx_apikey_hash ON "ApiKey"(key_hash);
```

### Import Job Table

```sql
CREATE TABLE "ImportJob" (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id BIGINT NOT NULL REFERENCES "Tenant"(id),
  status TEXT NOT NULL DEFAULT 'pending',
  total_count INT,
  processed_count INT DEFAULT 0,
  error_count INT DEFAULT 0,
  errors JSONB DEFAULT '[]',
  created_at TIMESTAMPTZ DEFAULT now(),
  completed_at TIMESTAMPTZ
);
```

### Asset Metadata Extension

Add to Asset table:
```sql
ALTER TABLE "Asset" ADD COLUMN metadata JSONB DEFAULT '{}';
ALTER TABLE "Asset" ADD COLUMN source TEXT;  -- 'upload', 'api', 'webhook', 'import'
```

---

## Request/Response Models

### Pydantic Models

```python
class AssetCreateRequest(BaseModel):
    filename: str
    content_type: str
    content: str  # base64 encoded
    tags: list[str] = []
    description: str | None = None
    metadata: dict = {}

class BulkAssetCreateRequest(BaseModel):
    assets: list[AssetCreateRequest]
    default_tags: list[str] = []
    on_duplicate: Literal["skip", "replace", "error"] = "skip"

class AssetCreateResponse(BaseModel):
    id: int
    filename: str
    tags: list[str]
    created_at: datetime

class BulkCreateResponse(BaseModel):
    created: int
    skipped: int
    errors: list[dict]

class ImportJobStatus(BaseModel):
    id: str
    status: Literal["pending", "processing", "completed", "failed"]
    total_count: int | None
    processed_count: int
    error_count: int
    created_at: datetime
    completed_at: datetime | None
```

---

## Authentication & Security

### API Key Generation

```python
def generate_api_key() -> tuple[str, str]:
    """
    Generate API key and hash.

    Returns: (plaintext_key, hash_for_storage)
    Key format: pico_live_xxxxxxxxxxxxx
    """
    raw = secrets.token_urlsafe(32)
    key = f"pico_live_{raw}"
    key_hash = bcrypt.hashpw(key.encode(), bcrypt.gensalt())
    return key, key_hash.decode()
```

### Key Verification

```python
async def verify_api_key(
    x_api_key: str = Header(...),
    session: AsyncSession = Depends(async_get_session),
) -> ApiKey:
    """
    Verify API key and return tenant context.

    - Check key exists and not expired
    - Update last_used_at
    - Check rate limit
    """
```

### Rate Limiting

- Default: 1000 requests/hour per key
- Bulk endpoints count as N requests (N = asset count)
- 429 response when exceeded

---

## ETL Pipeline Integration

### Scrapy Pipeline Example

```python
# In Scrapy project
class PinaColadaPipeline:
    def __init__(self):
        self.api_url = "https://api.pinacolada.co/api/v1/assets"
        self.api_key = os.environ["PINA_COLADA_API_KEY"]
        self.batch = []
        self.batch_size = 50

    def process_item(self, item, spider):
        self.batch.append({
            "filename": f"{item['id']}.json",
            "content_type": "application/json",
            "content": base64.b64encode(json.dumps(item).encode()).decode(),
            "tags": ["scraped", spider.name],
            "metadata": {
                "source_url": item.get("url"),
                "scraped_at": datetime.utcnow().isoformat()
            }
        })

        if len(self.batch) >= self.batch_size:
            self._flush()

        return item

    def _flush(self):
        if not self.batch:
            return
        requests.post(
            f"{self.api_url}/bulk",
            json={"assets": self.batch},
            headers={"X-API-Key": self.api_key}
        )
        self.batch = []
```

---

## Data Flow

### API Ingest Flow

```
External System
  → POST /api/v1/assets (with API key)
  → Verify key & rate limit
  → Decode base64 content
  → Validate size & type
  → Get/create tags
  → Store in BYTEA
  → Return asset ID
```

### Bulk Import Flow

```
ETL Pipeline
  → POST /api/v1/assets/bulk
  → Verify key
  → Begin transaction
  → Process each asset
    → Decode, validate, store
    → Track errors
  → Commit or rollback
  → Return summary
```

### Background Job Flow

```
Client
  → POST /api/v1/import-jobs (upload CSV)
  → Create ImportJob record
  → Queue for background processing
  → Return job ID

Background Worker
  → Pick up job from queue
  → Stream-process CSV/JSONL
  → Update progress periodically
  → Mark completed/failed

Client
  → Poll GET /api/v1/import-jobs/{id}
  → Get status & results
```

---

## File Structure

```
modules/agent/
├── src/
│   ├── models/
│   │   ├── ApiKey.py
│   │   └── ImportJob.py
│   ├── api/routes/
│   │   ├── assets_api.py      # API ingest endpoints
│   │   ├── api_keys.py        # Key management
│   │   └── import_jobs.py     # Background jobs
│   ├── lib/
│   │   └── api_auth.py        # Key verification
│   └── workers/
│       └── import_worker.py   # Background processor
```

---

## Constraints

- **Single asset:** 10MB max
- **Bulk request:** 100 assets, 50MB total
- **Import job:** 10,000 assets per file
- **Rate limit:** 1000 req/hour default
- **API key expiry:** Optional, recommended 90 days

---

## Future Enhancements

- Streaming uploads for large files
- S3 pre-signed URL workflow
- Deduplication by content hash
- Asset transformation pipeline (resize images, extract text)
- Webhook delivery (notify on asset creation)
- Usage analytics per API key
