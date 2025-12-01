# Document Upload with Tagging - Technical Specification

## Context

The system needs a way to ingest user documents (PDFs, text files, etc.) that can be tagged, linked to entities (Projects, Leads, Accounts), and later queried by the agentic AI. This is Phase 1 focusing on client-side upload.

**Note:** We use "Document" (not "Asset") since the existing `Asset` table is for polymorphic text content. Future phases may add `Image` or other asset types.

**Goal:** Keep document retrieval efficient for agent context. Tags and entity associations enable precise queries so the agent only loads relevant documents, minimizing token usage.

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

- Agent calls `get_user_documents(tags, entity_type, entity_ids)` during conversation
- Tool returns document content for context injection
- Text extraction for PDFs/documents

### Tagging: Reuse existing EntityTag

- Existing `Tag` and `EntityTag` tables are polymorphic
- Documents tagged via `EntityTag(tag_id, 'Document', document_id)`
- Consistent with how other entities are tagged

### Entity Association: Polymorphic

- `EntityDocument` junction table (follows existing `EntityTag` pattern)
- Documents can be linked to any entity type (Project, Lead, Account, etc.)
- Single table enables flexible associations without creating N junction tables

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

### Migration (041_asset_refactor.sql)

```sql
-- Refactor Asset to be base table for joined table inheritance
-- Drop old polymorphic columns, add common asset fields

-- 1. Drop old Asset structure (backup data if needed)
DROP TABLE IF EXISTS "Asset" CASCADE;

-- 2. Create new Asset base table
CREATE TABLE "Asset" (
  id BIGSERIAL PRIMARY KEY,
  asset_type TEXT NOT NULL,  -- 'document', 'image' (discriminator)
  tenant_id BIGINT NOT NULL REFERENCES "Tenant"(id) ON DELETE CASCADE,
  user_id BIGINT NOT NULL REFERENCES "User"(id) ON DELETE CASCADE,
  filename TEXT NOT NULL,
  content_type TEXT NOT NULL,  -- MIME type
  description TEXT,
  created_at TIMESTAMPTZ DEFAULT now(),
  updated_at TIMESTAMPTZ DEFAULT now()
);
CREATE INDEX idx_asset_tenant ON "Asset"(tenant_id);
CREATE INDEX idx_asset_user ON "Asset"(user_id);
CREATE INDEX idx_asset_type ON "Asset"(asset_type);

-- 3. Create Document extension table (joined inheritance)
CREATE TABLE "Document" (
  id BIGINT PRIMARY KEY REFERENCES "Asset"(id) ON DELETE CASCADE,
  storage_path TEXT NOT NULL,  -- path in storage backend
  file_size BIGINT NOT NULL
);

-- 4. EntityAsset polymorphic junction (links assets to any entity)
CREATE TABLE "EntityAsset" (
  asset_id BIGINT NOT NULL REFERENCES "Asset"(id) ON DELETE CASCADE,
  entity_type TEXT NOT NULL,  -- 'Project', 'Lead', 'Account', etc.
  entity_id BIGINT NOT NULL,
  created_at TIMESTAMPTZ DEFAULT now(),
  PRIMARY KEY (asset_id, entity_type, entity_id)
);
CREATE INDEX idx_entityasset_entity ON "EntityAsset"(entity_type, entity_id);
CREATE INDEX idx_entityasset_asset ON "EntityAsset"(asset_id);

-- Tagging uses existing EntityTag table:
-- INSERT INTO "EntityTag"(tag_id, entity_type, entity_id) VALUES (?, 'Asset', asset_id);
```

**Future Image table would be:**
```sql
CREATE TABLE "Image" (
  id BIGINT PRIMARY KEY REFERENCES "Asset"(id) ON DELETE CASCADE,
  storage_path TEXT NOT NULL,
  file_size BIGINT NOT NULL,
  width INT,
  height INT,
  thumbnail_path TEXT
);
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

### 2. SQLAlchemy Models (Joined Table Inheritance)

#### `models/Asset.py` (Base class)

```python
class Asset(Base):
    """Base asset class using joined table inheritance."""
    __tablename__ = "Asset"

    id = Column(BigInteger, primary_key=True)
    asset_type = Column(Text, nullable=False)  # discriminator
    tenant_id = Column(BigInteger, ForeignKey("Tenant.id", ondelete="CASCADE"), nullable=False)
    user_id = Column(BigInteger, ForeignKey("User.id", ondelete="CASCADE"), nullable=False)
    filename = Column(Text, nullable=False)
    content_type = Column(Text, nullable=False)
    description = Column(Text)
    created_at = Column(DateTime(timezone=True), server_default=func.now())
    updated_at = Column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now())

    tenant = relationship("Tenant")
    user = relationship("User")

    __mapper_args__ = {
        "polymorphic_on": asset_type,
        "polymorphic_identity": "asset",
    }
```

#### `models/Document.py` (Extends Asset)

```python
class Document(Asset):
    """Document asset - files stored in external storage."""
    __tablename__ = "Document"

    id = Column(BigInteger, ForeignKey("Asset.id", ondelete="CASCADE"), primary_key=True)
    storage_path = Column(Text, nullable=False)
    file_size = Column(BigInteger, nullable=False)

    __mapper_args__ = {
        "polymorphic_identity": "document",
    }
```

#### `models/EntityAsset.py`

```python
# Polymorphic join table for linking any asset to any entity
EntityAsset = Table(
    "EntityAsset",
    Base.metadata,
    Column("asset_id", BigInteger, ForeignKey("Asset.id", ondelete="CASCADE"), primary_key=True),
    Column("entity_type", Text, primary_key=True),  # 'Project', 'Lead', 'Account'
    Column("entity_id", BigInteger, primary_key=True),
    Column("created_at", DateTime(timezone=True), server_default=func.now(), nullable=False),
)
```

---

### 3. API Routes (`api/routes/documents.py`)

#### Upload Document

```python
@router.post("/documents")
@require_auth
async def upload_document(
    file: UploadFile,
    tags: str = Form(""),  # comma-separated
    entity_type: str = Form(None),  # 'Project', 'Lead', etc.
    entity_id: int = Form(None),
    description: str = Form(""),
    session: AsyncSession = Depends(async_get_session),
    user: User = Depends(get_current_user),
    storage: StorageBackend = Depends(get_storage),
):
    """
    Upload file with tags and optional entity association.

    - Validates file size (<10MB)
    - Uploads to storage backend
    - Creates/reuses tags within tenant (via EntityTag)
    - Links to entity via polymorphic EntityDocument table
    """
    content = await file.read()
    if len(content) > 10 * 1024 * 1024:
        raise HTTPException(400, "File exceeds 10MB limit")

    document_id = generate_snowflake_id()
    storage_path = f"{user.tenant_id}/{document_id}/{file.filename}"

    await storage.upload(storage_path, content, file.content_type)

    document = Document(
        id=document_id,
        tenant_id=user.tenant_id,
        user_id=user.id,
        filename=file.filename,
        content_type=file.content_type,
        storage_path=storage_path,
        file_size=len(content),
        description=description,
    )
    session.add(document)
    # Handle tags via EntityTag, entity link via EntityDocument...
```

#### List Documents

```python
@router.get("/documents")
@require_auth
async def list_documents(
    tags: str = Query(""),  # comma-separated filter
    entity_type: str = Query(None),  # filter by entity type
    entity_id: int = Query(None),  # filter by entity id
    session: AsyncSession = Depends(async_get_session),
    user: User = Depends(get_current_user),
):
    """
    List user's documents with optional tag/entity filter.

    Returns metadata only (no file content).
    """
```

#### Download Document

```python
@router.get("/documents/{document_id}/download")
@require_auth
async def download_document(
    document_id: int,
    session: AsyncSession = Depends(async_get_session),
    user: User = Depends(get_current_user),
    storage: StorageBackend = Depends(get_storage),
):
    """Return presigned URL (R2) or stream file (local)."""
```

#### Delete Document

```python
@router.delete("/documents/{document_id}")
@require_auth
async def delete_document(
    document_id: int,
    session: AsyncSession = Depends(async_get_session),
    user: User = Depends(get_current_user),
    storage: StorageBackend = Depends(get_storage),
):
    """Delete document from storage and database."""
```

#### Link/Unlink Document

```python
@router.post("/documents/{document_id}/link")
@require_auth
async def link_document_to_entity(
    document_id: int,
    entity_type: str,
    entity_id: int,
    session: AsyncSession = Depends(async_get_session),
    user: User = Depends(get_current_user),
):
    """Link document to an entity (Project, Lead, Account, etc.)."""

@router.delete("/documents/{document_id}/link")
@require_auth
async def unlink_document_from_entity(
    document_id: int,
    entity_type: str,
    entity_id: int,
    session: AsyncSession = Depends(async_get_session),
    user: User = Depends(get_current_user),
):
    """Remove document link from an entity."""
```

---

### 4. Agent Tool

#### `tools/document_tools.py`

```python
@tool
async def get_user_documents(
    tags: list[str] = None,
    entity_type: str = None,  # 'Project', 'Lead', 'Account'
    entity_ids: list[int] = None,
    user_id: int,
    tenant_id: int,
) -> str:
    """
    Query user documents by tags and/or entity associations for agent context.

    Args:
        tags: List of tag names to filter by (AND logic, via EntityTag)
        entity_type: Entity type to filter by (e.g., 'Project')
        entity_ids: List of entity IDs to filter by (OR logic within type)
        user_id: Current user ID
        tenant_id: Current tenant ID

    Returns:
        Concatenated document content (text extracted from files)
    """
    # Query documents matching filters via EntityDocument/EntityTag joins
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
  entityType?: string;  // 'Project', 'Lead', etc.
  entityId?: number;
}

// Features:
// - Drag-and-drop file input
// - Tag selector with autocomplete
// - Auto-links to entity if provided
// - Upload progress indicator
// - File size validation (client-side)
```

#### `components/assets/AssetList.tsx`

```typescript
interface AssetListProps {
  filterTags?: string[];
  entityType?: string;
  entityId?: number;
}

// Features:
// - Table/grid view of assets
// - Tag filter chips
// - Download/delete actions
// - Link/unlink to entities
```

---

### 6. Frontend Integration Points

#### Entity Forms
- **LeadForm**: Add "Attachments" section to upload/view documents linked to the Lead
- **AccountForm**: Add "Attachments" section to upload/view documents linked to the Account

#### Sidebar Navigation
```
Assets (new top-level tab)
└── Documents (child tab - list all documents, filter by tags/entities)
└── Images (future - child tab for image assets)
```

#### Component Placement
- `app/assets/page.tsx` - Main assets page with Documents tab
- `app/assets/documents/page.tsx` - Documents list view
- Reusable `<DocumentAttachments entityType="Lead" entityId={id} />` component for forms

---

## Data Flow

### Upload Flow

```
Client (file + tags + entity_type + entity_id)
  → POST /documents (multipart)
  → Validate size < 10MB
  → Upload to storage backend (local or R2)
  → Create Document record with storage_path
  → Get/create tags, link via EntityTag('Document', document_id)
  → Link to entity via EntityDocument (if provided)
  → Return document metadata
```

### Link Document to Multiple Entities

```
Document uploaded to Project "Q4 Campaign"
  → POST /documents/123/link {entity_type: "Lead", entity_id: 456}
  → Now document visible on both Project and Lead views
  → POST /documents/123/link {entity_type: "Account", entity_id: 789}
  → Same document now linked to Project, Lead, and Account
```

### Agent Query Flow

```
User: "Use my resume to write a cover letter"
  → Agent detects need for resume
  → Calls get_user_documents(tags=["resume"])
  → Tool queries Document + EntityTag
  → Downloads file from storage backend
  → Extracts text from file
  → Returns content to agent context
  → Agent uses content for response

User: "What documents do we have for Acme Corp?"
  → Calls get_user_documents(entity_type="Account", entity_ids=[123])
  → Returns all documents linked to that Account
```

---

## File Structure

```
modules/agent/
├── migrations/
│   └── 041_document_storage.sql
├── src/
│   ├── lib/
│   │   └── storage.py
│   ├── models/
│   │   ├── Document.py      # renamed from Asset to avoid confusion
│   │   ├── EntityDocument.py  # polymorphic junction (like EntityTag)
│   │   └── ... (existing Tag.py with EntityTag already exists)
│   ├── api/routes/
│   │   └── documents.py
│   └── agent/tools/
│       └── document_tools.py

modules/client/
├── components/documents/
│   ├── DocumentUpload.tsx
│   ├── DocumentList.tsx
│   └── TagInput.tsx
├── app/documents/
│   └── page.tsx
└── lib/api/
    └── documents.ts

storage/           # gitignored, local dev only
└── documents/
    └── {tenant_id}/
        └── {document_id}/
            └── {filename}
```

---

## Constraints

- **File size:** 10MB max
- **Supported types:** PDF, TXT, MD, DOC/DOCX (Phase 1)
- **Tag names:** Lowercase, alphanumeric + hyphens
- **Tags per document:** No hard limit, recommend < 10
- **Storage cleanup:** Delete from storage when Document record deleted

---

## Future Enhancements (Phase 2+)

- API-driven bulk ingest (`spec/asset-api-ingest.md`)
- Full-text search within assets
- Asset versioning
- Automatic tag suggestions via LLM (see `tagging-strategy.md`)
