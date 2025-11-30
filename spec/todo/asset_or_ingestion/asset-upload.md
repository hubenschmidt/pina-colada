# Asset Upload with Tagging - Technical Specification

## Context

The system needs a way to ingest user assets (documents, files) that can be tagged, linked to projects, and later queried by the agentic AI. This is Phase 1 focusing on client-side upload.

**Goal:** Keep asset retrieval efficient for agent context. Tags and project associations enable precise queries so the agent only loads relevant assets, minimizing token usage.

**Phase 2** (API-driven/ETL ingest) will be documented in `spec/asset-api-ingest.md`.

---

## Architecture Overview

### Storage Strategy: Environment-Based

| Environment | Backend | Path Pattern |
|-------------|---------|--------------|
| Local dev | Filesystem | `./storage/assets/{tenant_id}/{asset_id}/{filename}` |
| Test/Prod | Cloudflare R2 | `{tenant_id}/{asset_id}/{filename}` |

**Why Cloudflare R2:**
- 10GB free storage
- 10M reads/month, 1M writes/month free
- Zero egress fees (unlike S3/GCS)
- S3-compatible API (easy migration)

### Agent Access: On-demand via tool

- Agent calls `get_user_assets(tags, project_ids)` during conversation
- Tool returns asset content for context injection
- Text extraction for PDFs/documents

### Tagging: Normalized junction table

- `Tag` table scoped to tenant
- `AssetTag` junction enables many-to-many
- Efficient "find all assets with tag X" queries via indexed FK lookup

### Project Association: Many-to-Many

- `AssetProject` junction table (follows existing pattern: LeadProject, AccountProject)
- Assets can belong to multiple projects
- Enables project-scoped asset queries

---

## Environment Configuration

```python
# .env
STORAGE_BACKEND=local  # "local" or "r2"

# R2 config (only needed when STORAGE_BACKEND=r2)
R2_ACCOUNT_ID=your_account_id
R2_ACCESS_KEY_ID=your_access_key
R2_SECRET_ACCESS_KEY=your_secret_key
R2_BUCKET_NAME=pina-colada-assets
```

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

-- Asset table (metadata only, no file data)
CREATE TABLE "Asset" (
  id BIGSERIAL PRIMARY KEY,
  tenant_id BIGINT NOT NULL REFERENCES "Tenant"(id) ON DELETE CASCADE,
  user_id BIGINT NOT NULL REFERENCES "User"(id) ON DELETE CASCADE,
  filename TEXT NOT NULL,
  content_type TEXT NOT NULL,
  storage_path TEXT NOT NULL,  -- path in storage backend
  file_size BIGINT NOT NULL,
  description TEXT,
  created_at TIMESTAMPTZ DEFAULT now(),
  updated_at TIMESTAMPTZ DEFAULT now()
);
CREATE INDEX idx_asset_user ON "Asset"(user_id);
CREATE INDEX idx_asset_tenant ON "Asset"(tenant_id);

-- AssetTag junction table
CREATE TABLE "AssetTag" (
  asset_id BIGINT NOT NULL REFERENCES "Asset"(id) ON DELETE CASCADE,
  tag_id BIGINT NOT NULL REFERENCES "Tag"(id) ON DELETE CASCADE,
  PRIMARY KEY (asset_id, tag_id)
);
CREATE INDEX idx_assettag_tag ON "AssetTag"(tag_id);

-- AssetProject junction table
CREATE TABLE "AssetProject" (
  asset_id BIGINT NOT NULL REFERENCES "Asset"(id) ON DELETE CASCADE,
  project_id BIGINT NOT NULL REFERENCES "Project"(id) ON DELETE CASCADE,
  created_at TIMESTAMPTZ DEFAULT now(),
  PRIMARY KEY (asset_id, project_id)
);
CREATE INDEX idx_assetproject_project ON "AssetProject"(project_id);
```

---

## Component Specifications

### 1. Storage Abstraction (`lib/storage.py`)

```python
from abc import ABC, abstractmethod
from pathlib import Path
import aiofiles
import boto3

class StorageBackend(ABC):
    @abstractmethod
    async def upload(self, path: str, data: bytes, content_type: str) -> str:
        """Upload file, return storage path."""

    @abstractmethod
    async def download(self, path: str) -> bytes:
        """Download file content."""

    @abstractmethod
    async def delete(self, path: str) -> None:
        """Delete file."""

    @abstractmethod
    def get_url(self, path: str) -> str:
        """Get download URL (signed for R2, file:// for local)."""


class LocalStorage(StorageBackend):
    """Filesystem storage for local development."""

    def __init__(self, base_path: str = "./storage/assets"):
        self.base_path = Path(base_path)
        self.base_path.mkdir(parents=True, exist_ok=True)

    async def upload(self, path: str, data: bytes, content_type: str) -> str:
        full_path = self.base_path / path
        full_path.parent.mkdir(parents=True, exist_ok=True)
        async with aiofiles.open(full_path, "wb") as f:
            await f.write(data)
        return path

    async def download(self, path: str) -> bytes:
        async with aiofiles.open(self.base_path / path, "rb") as f:
            return await f.read()

    async def delete(self, path: str) -> None:
        (self.base_path / path).unlink(missing_ok=True)

    def get_url(self, path: str) -> str:
        return f"file://{self.base_path / path}"


class R2Storage(StorageBackend):
    """Cloudflare R2 storage for production."""

    def __init__(self, account_id: str, access_key: str, secret_key: str, bucket: str):
        self.bucket = bucket
        self.client = boto3.client(
            "s3",
            endpoint_url=f"https://{account_id}.r2.cloudflarestorage.com",
            aws_access_key_id=access_key,
            aws_secret_access_key=secret_key,
        )

    async def upload(self, path: str, data: bytes, content_type: str) -> str:
        self.client.put_object(
            Bucket=self.bucket,
            Key=path,
            Body=data,
            ContentType=content_type,
        )
        return path

    async def download(self, path: str) -> bytes:
        response = self.client.get_object(Bucket=self.bucket, Key=path)
        return response["Body"].read()

    async def delete(self, path: str) -> None:
        self.client.delete_object(Bucket=self.bucket, Key=path)

    def get_url(self, path: str, expires_in: int = 3600) -> str:
        return self.client.generate_presigned_url(
            "get_object",
            Params={"Bucket": self.bucket, "Key": path},
            ExpiresIn=expires_in,
        )


def get_storage() -> StorageBackend:
    """Factory function based on environment."""
    backend = os.getenv("STORAGE_BACKEND", "local")
    if backend == "r2":
        return R2Storage(
            account_id=os.environ["R2_ACCOUNT_ID"],
            access_key=os.environ["R2_ACCESS_KEY_ID"],
            secret_key=os.environ["R2_SECRET_ACCESS_KEY"],
            bucket=os.environ["R2_BUCKET_NAME"],
        )
    return LocalStorage()
```

---

### 2. SQLAlchemy Models

#### `models/Tag.py`

```python
class Tag(Base):
    __tablename__ = "Tag"

    id = Column(BigInteger, primary_key=True)
    tenant_id = Column(BigInteger, ForeignKey("Tenant.id", ondelete="CASCADE"), nullable=False)
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
    tenant_id = Column(BigInteger, ForeignKey("Tenant.id", ondelete="CASCADE"), nullable=False)
    user_id = Column(BigInteger, ForeignKey("User.id", ondelete="CASCADE"), nullable=False)
    filename = Column(Text, nullable=False)
    content_type = Column(Text, nullable=False)
    storage_path = Column(Text, nullable=False)
    file_size = Column(BigInteger, nullable=False)
    description = Column(Text)
    created_at = Column(DateTime(timezone=True), server_default=func.now())
    updated_at = Column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now())

    tenant = relationship("Tenant")
    user = relationship("User")
    tags = relationship("Tag", secondary="AssetTag", back_populates="assets")
    projects = relationship("Project", secondary="AssetProject", back_populates="assets")
```

#### `models/AssetTag.py`

```python
class AssetTag(Base):
    __tablename__ = "AssetTag"

    asset_id = Column(BigInteger, ForeignKey("Asset.id", ondelete="CASCADE"), primary_key=True)
    tag_id = Column(BigInteger, ForeignKey("Tag.id", ondelete="CASCADE"), primary_key=True)
```

#### `models/AssetProject.py`

```python
class AssetProject(Base):
    __tablename__ = "AssetProject"

    asset_id = Column(BigInteger, ForeignKey("Asset.id", ondelete="CASCADE"), primary_key=True)
    project_id = Column(BigInteger, ForeignKey("Project.id", ondelete="CASCADE"), primary_key=True)
    created_at = Column(DateTime(timezone=True), server_default=func.now())
```

---

### 3. API Routes (`api/routes/assets.py`)

#### Upload Asset

```python
@router.post("/assets")
@require_auth
async def upload_asset(
    file: UploadFile,
    tags: str = Form(""),  # comma-separated
    project_ids: str = Form(""),  # comma-separated
    description: str = Form(""),
    session: AsyncSession = Depends(async_get_session),
    user: User = Depends(get_current_user),
    storage: StorageBackend = Depends(get_storage),
):
    """
    Upload file with tags and project associations.

    - Validates file size (<10MB)
    - Uploads to storage backend
    - Creates/reuses tags within tenant
    - Links to projects via junction table
    """
    # Validate size
    content = await file.read()
    if len(content) > 10 * 1024 * 1024:
        raise HTTPException(400, "File exceeds 10MB limit")

    # Generate storage path
    asset_id = generate_snowflake_id()
    storage_path = f"{user.tenant_id}/{asset_id}/{file.filename}"

    # Upload to storage
    await storage.upload(storage_path, content, file.content_type)

    # Create asset record
    asset = Asset(
        id=asset_id,
        tenant_id=user.tenant_id,
        user_id=user.id,
        filename=file.filename,
        content_type=file.content_type,
        storage_path=storage_path,
        file_size=len(content),
        description=description,
    )
    session.add(asset)

    # Handle tags and projects...
```

#### List Assets

```python
@router.get("/assets")
@require_auth
async def list_assets(
    tags: str = Query(""),  # comma-separated filter
    project_id: int = Query(None),  # filter by project
    session: AsyncSession = Depends(async_get_session),
    user: User = Depends(get_current_user),
):
    """
    List user's assets with optional tag/project filter.

    Returns metadata only (no file content).
    """
```

#### Get Asset (Download)

```python
@router.get("/assets/{asset_id}/download")
@require_auth
async def download_asset(
    asset_id: int,
    session: AsyncSession = Depends(async_get_session),
    user: User = Depends(get_current_user),
    storage: StorageBackend = Depends(get_storage),
):
    """Return presigned URL (R2) or stream file (local)."""
```

#### Delete Asset

```python
@router.delete("/assets/{asset_id}")
@require_auth
async def delete_asset(
    asset_id: int,
    session: AsyncSession = Depends(async_get_session),
    user: User = Depends(get_current_user),
    storage: StorageBackend = Depends(get_storage),
):
    """Delete asset from storage and database."""
```

#### Update Asset Projects

```python
@router.patch("/assets/{asset_id}/projects")
@require_auth
async def update_asset_projects(
    asset_id: int,
    project_ids: list[int],
    session: AsyncSession = Depends(async_get_session),
    user: User = Depends(get_current_user),
):
    """Replace asset's project associations."""
```

---

### 4. Agent Tool

#### `tools/asset_tools.py`

```python
@tool
async def get_user_assets(
    tags: list[str] = None,
    project_ids: list[int] = None,
    user_id: int,
    tenant_id: int,
) -> str:
    """
    Query user assets by tags and/or projects for agent context.

    Args:
        tags: List of tag names to filter by (AND logic)
        project_ids: List of project IDs to filter by (OR logic)
        user_id: Current user ID
        tenant_id: Current tenant ID

    Returns:
        Concatenated asset content (text extracted from files)
    """
    # Query assets matching filters
    # Download from storage
    # Extract text from PDFs/docs
    # Return formatted content for agent context
```

**Text Extraction:**
- PDF: Use `pypdf` or similar
- Plain text: Direct decode
- Other: Return filename + metadata

---

### 5. Client Components

#### `components/assets/AssetUpload.tsx`

```typescript
interface AssetUploadProps {
  onUploadComplete?: (asset: Asset) => void;
  defaultProjectIds?: number[];
}

// Features:
// - Drag-and-drop file input
// - Tag selector with autocomplete
// - Project selector (multi-select)
// - Upload progress indicator
// - File size validation (client-side)
```

#### `components/assets/AssetList.tsx`

```typescript
interface AssetListProps {
  filterTags?: string[];
  filterProjectId?: number;
}

// Features:
// - Table/grid view of assets
// - Tag filter chips
// - Project filter dropdown
// - Download/delete actions
// - Tag/project editing inline
```

---

## Data Flow

### Upload Flow

```
Client (file + tags + project_ids)
  → POST /assets (multipart)
  → Validate size < 10MB
  → Upload to storage backend (local or R2)
  → Create Asset record with storage_path
  → Get/create tags in tenant, link via AssetTag
  → Link to projects via AssetProject
  → Return asset metadata
```

### Agent Query Flow

```
User: "Use my resume to write a cover letter"
  → Agent detects need for resume
  → Calls get_user_assets(tags=["resume"])
  → Tool queries Asset + AssetTag + Tag
  → Downloads file from storage backend
  → Extracts text from file
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
│   ├── lib/
│   │   └── storage.py
│   ├── models/
│   │   ├── Asset.py
│   │   ├── Tag.py
│   │   ├── AssetTag.py
│   │   └── AssetProject.py
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

storage/           # gitignored, local dev only
└── assets/
    └── {tenant_id}/
        └── {asset_id}/
            └── {filename}
```

---

## Constraints

- **File size:** 10MB max
- **Supported types:** PDF, TXT, MD, DOC/DOCX (Phase 1)
- **Tag names:** Lowercase, alphanumeric + hyphens
- **Tags per asset:** No hard limit, recommend < 10
- **Storage cleanup:** Delete from storage when Asset record deleted

---

## Future Enhancements (Phase 2+)

- API-driven bulk ingest (`spec/asset-api-ingest.md`)
- Full-text search within assets
- Asset versioning
- Automatic tag suggestions via LLM (see `tagging-strategy.md`)
