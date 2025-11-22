# Asset Upload with Tagging - Technical Specification

## Context

The system needs a way to ingest user assets (documents, files) that can be tagged and later queried by the agentic AI. This is Phase 1 focusing on client-side upload.

**Goal:** Keep asset retrieval efficient for agent context. Tags enable precise queries so the agent only loads relevant assets, minimizing token usage.

**Phase 2** (API-driven/ETL ingest) will be documented in `spec/asset-api-ingest.md`.

---

## Architecture Overview

### Storage Strategy: PostgreSQL BYTEA

- Files stored directly in database as binary
- ~10MB practical limit per file
- Simple deployment, no external storage dependencies
- Easy backup/restore with database

### Agent Access: On-demand via tool

- Agent calls `get_user_assets(tags)` during conversation
- Tool returns asset content for context injection
- Text extraction for PDFs/documents

### Tagging: Normalized junction table

- `Tag` table scoped to tenant
- `AssetTag` junction enables many-to-many
- Efficient "find all assets with tag X" queries via indexed FK lookup
- More scalable than JSONB for relational queries

---

## Database Schema

### Migration

```sql
-- Tag table (tenant-scoped)
CREATE TABLE "Tag" (
  id BIGSERIAL PRIMARY KEY,
  tenant_id BIGINT NOT NULL REFERENCES "Tenant"(id) ON DELETE CASCADE,
  name TEXT NOT NULL,
  created_at TIMESTAMPTZ DEFAULT now()
);
CREATE UNIQUE INDEX idx_tag_tenant_name ON "Tag"(tenant_id, LOWER(name));

-- Asset table
CREATE TABLE "Asset" (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL REFERENCES "User"(id) ON DELETE CASCADE,
  filename TEXT NOT NULL,
  content_type TEXT NOT NULL,
  file_data BYTEA NOT NULL,
  file_size BIGINT NOT NULL,
  description TEXT,
  created_at TIMESTAMPTZ DEFAULT now(),
  updated_at TIMESTAMPTZ DEFAULT now()
);
CREATE INDEX idx_asset_user ON "Asset"(user_id);

-- Junction table
CREATE TABLE "AssetTag" (
  asset_id BIGINT NOT NULL REFERENCES "Asset"(id) ON DELETE CASCADE,
  tag_id BIGINT NOT NULL REFERENCES "Tag"(id) ON DELETE CASCADE,
  PRIMARY KEY (asset_id, tag_id)
);
CREATE INDEX idx_assettag_tag ON "AssetTag"(tag_id);
```

---

## Component Specifications

### 1. SQLAlchemy Models

#### `models/Tag.py`

```python
class Tag(Base):
    __tablename__ = "Tag"

    id = Column(BigInteger, primary_key=True)
    tenant_id = Column(BigInteger, ForeignKey("Tenant.id"), nullable=False)
    name = Column(Text, nullable=False)
    created_at = Column(DateTime(timezone=True), server_default=func.now())

    tenant = relationship("Tenant")
    assets = relationship("Asset", secondary="AssetTag", back_populates="tags")
```

#### `models/Asset.py`

```python
class Asset(Base):
    __tablename__ = "Asset"

    id = Column(BigInteger, primary_key=True)
    user_id = Column(BigInteger, ForeignKey("User.id"), nullable=False)
    filename = Column(Text, nullable=False)
    content_type = Column(Text, nullable=False)
    file_data = Column(LargeBinary, nullable=False)
    file_size = Column(BigInteger, nullable=False)
    description = Column(Text)
    created_at = Column(DateTime(timezone=True), server_default=func.now())
    updated_at = Column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now())

    user = relationship("User")
    tags = relationship("Tag", secondary="AssetTag", back_populates="assets")
```

#### `models/AssetTag.py`

```python
class AssetTag(Base):
    __tablename__ = "AssetTag"

    asset_id = Column(BigInteger, ForeignKey("Asset.id"), primary_key=True)
    tag_id = Column(BigInteger, ForeignKey("Tag.id"), primary_key=True)
```

---

### 2. API Routes (`api/routes/assets.py`)

#### Upload Asset

```python
@router.post("/assets")
@require_auth
async def upload_asset(
    file: UploadFile,
    tags: str = Form(""),  # comma-separated
    description: str = Form(""),
    session: AsyncSession = Depends(async_get_session),
    user: User = Depends(get_current_user),
):
    """
    Upload file with tags.

    - Validates file size (<10MB)
    - Creates/reuses tags within tenant
    - Stores file in BYTEA
    """
```

#### List Assets

```python
@router.get("/assets")
@require_auth
async def list_assets(
    tags: str = Query(""),  # comma-separated filter
    session: AsyncSession = Depends(async_get_session),
    user: User = Depends(get_current_user),
):
    """
    List user's assets with optional tag filter.

    Returns metadata only (no file_data).
    """
```

#### Get Asset

```python
@router.get("/assets/{asset_id}")
@require_auth
async def get_asset(
    asset_id: int,
    session: AsyncSession = Depends(async_get_session),
    user: User = Depends(get_current_user),
):
    """Download asset file."""
```

#### Delete Asset

```python
@router.delete("/assets/{asset_id}")
@require_auth
async def delete_asset(
    asset_id: int,
    session: AsyncSession = Depends(async_get_session),
    user: User = Depends(get_current_user),
):
    """Delete asset and orphaned tags."""
```

#### Update Tags

```python
@router.patch("/assets/{asset_id}/tags")
@require_auth
async def update_asset_tags(
    asset_id: int,
    tags: list[str],
    session: AsyncSession = Depends(async_get_session),
    user: User = Depends(get_current_user),
):
    """Replace asset's tags."""
```

#### List Tags

```python
@router.get("/tags")
@require_auth
async def list_tags(
    session: AsyncSession = Depends(async_get_session),
    user: User = Depends(get_current_user),
):
    """List all tags in tenant with usage counts."""
```

---

### 3. Agent Tool

#### `tools/asset_tools.py`

```python
@tool
async def get_user_assets(
    tags: list[str],
    user_id: int,
) -> str:
    """
    Query user assets by tags for agent context.

    Args:
        tags: List of tag names to filter by (AND logic)
        user_id: Current user ID

    Returns:
        Concatenated asset content (text extracted from files)
    """
    # Query assets matching all tags
    # Extract text from PDFs/docs
    # Return formatted content for agent context
```

**Text Extraction:**
- PDF: Use `pypdf` or similar
- Plain text: Direct decode
- Other: Return filename + metadata

---

### 4. Client Components

#### `components/assets/AssetUpload.tsx`

```typescript
interface AssetUploadProps {
  onUploadComplete?: (asset: Asset) => void;
}

// Features:
// - Drag-and-drop file input
// - Tag selector with autocomplete
// - Upload progress indicator
// - File size validation (client-side)
```

#### `components/assets/AssetList.tsx`

```typescript
interface AssetListProps {
  filterTags?: string[];
}

// Features:
// - Table/grid view of assets
// - Tag filter chips
// - Download/delete actions
// - Tag editing inline
```

#### `components/assets/TagInput.tsx`

```typescript
interface TagInputProps {
  selectedTags: string[];
  onChange: (tags: string[]) => void;
  availableTags?: string[];
}

// Features:
// - Autocomplete from existing tags
// - Create new tags inline
// - Chip display for selected
```

---

## Data Flow

### Upload Flow

```
Client (file + tags)
  → POST /assets (multipart)
  → Validate size < 10MB
  → Get/create tags in tenant
  → Store file_data in BYTEA
  → Link via AssetTag junction
  → Return asset metadata
```

### Agent Query Flow

```
User: "Use my resume to write a cover letter"
  → Agent detects need for resume
  → Calls get_user_assets(["resume"])
  → Tool queries Asset + AssetTag + Tag
  → Extracts text from file_data
  → Returns content to agent context
  → Agent uses content for response
```

---

## File Structure

```
modules/agent/
├── migrations/
│   └── versions/
│       └── xxxx_add_asset_tables.py
├── src/
│   ├── models/
│   │   ├── Asset.py
│   │   ├── Tag.py
│   │   └── AssetTag.py
│   ├── api/routes/
│   │   └── assets.py
│   └── agent/tools/
│       └── asset_tools.py

modules/client/
├── components/assets/
│   ├── AssetUpload.tsx
│   ├── AssetList.tsx
│   └── TagInput.tsx
├── app/assets/
│   └── page.tsx
└── lib/api/
    └── assets.ts
```

---

## Constraints

- **File size:** 10MB max (BYTEA practical limit)
- **Supported types:** PDF, TXT, MD, DOC/DOCX (Phase 1)
- **Tag names:** Lowercase, alphanumeric + hyphens
- **Tags per asset:** No hard limit, recommend < 10

---

## Future Enhancements (Phase 2+)

- API-driven bulk ingest (`spec/asset-api-ingest.md`)
- S3/Azure Blob storage for larger files
- Full-text search within assets
- Asset versioning
- Automatic tag suggestions via LLM
