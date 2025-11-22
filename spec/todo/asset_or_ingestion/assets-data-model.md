# Assets Data Model Specification

## Overview

A generic, polymorphic content storage system enabling AI agents to build context from user-provided assets (resumes, cover letters, sample answers, etc.) and associate them with any entity in the CRM.

## Schema

### Asset Table

```sql
CREATE TABLE asset (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    entity_type VARCHAR(50) NOT NULL,      -- 'lead', 'job', 'organization', 'user'
    entity_id BIGINT NOT NULL,              -- FK to associated entity
    content TEXT NOT NULL,                  -- embedded text content
    content_type VARCHAR(100) NOT NULL,     -- MIME type: 'text/plain', 'application/pdf'
    filename VARCHAR(255),                  -- original filename
    checksum VARCHAR(64),                   -- SHA-256 for deduplication
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    INDEX idx_entity (entity_type, entity_id),
    INDEX idx_checksum (checksum)
);
```

### AssetTag Table

```sql
CREATE TABLE asset_tag (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    asset_id BIGINT NOT NULL,
    tag VARCHAR(100) NOT NULL,

    FOREIGN KEY (asset_id) REFERENCES asset(id) ON DELETE CASCADE,
    UNIQUE INDEX idx_asset_tag (asset_id, tag),
    INDEX idx_tag (tag)
);
```

## Polymorphic Association Pattern

The `entity_type` + `entity_id` pattern allows assets to attach to any entity without schema changes:

| entity_type | entity_id references |
|-------------|---------------------|
| `user` | User.id (personal assets like resumes) |
| `lead` | Lead.id (base lead context) |
| `job` | Job.id (job-specific assets) |
| `organization` | Organization.id (company research docs) |
| `deal` | Deal.id (deal-level context) |

**Query pattern**:
```python
assets = session.query(Asset).filter(
    Asset.entity_type == 'job',
    Asset.entity_id == job_id
).all()
```

## Tag Taxonomy

Recommended tags for job-search use case:

| Tag | Description |
|-----|-------------|
| `resume` | CV/resume document |
| `cover_letter` | Cover letter variants |
| `sample_answers` | Pre-written Q&A responses |
| `portfolio` | Work samples |
| `summary` | Brief profile summary |
| `research` | Company/industry research |
| `notes` | User notes and preferences |

Tags are flexible - agents can create new tags as needed.

## AI Agent Usage Patterns

### Context Retrieval for Job Applications

```python
def get_user_context_assets(user_id: int, tags: list[str] = None) -> list[Asset]:
    """Retrieve assets for building agent context."""
    query = session.query(Asset).filter(
        Asset.entity_type == 'user',
        Asset.entity_id == user_id
    )
    if tags:
        query = query.join(AssetTag).filter(AssetTag.tag.in_(tags))
    return query.all()
```

### Enriching Lead Context

```python
def get_lead_context(lead_id: int) -> str:
    """Build context string from user assets + lead-specific assets."""
    user_assets = get_user_context_assets(user_id, ['resume', 'summary'])
    lead_assets = session.query(Asset).filter(
        Asset.entity_type == 'lead',
        Asset.entity_id == lead_id
    ).all()

    return "\n---\n".join(m.content for m in user_assets + lead_assets)
```

### Web Search Context

When searching for leads/opportunities, the agent retrieves relevant assets to:
- Understand user qualifications and preferences
- Match opportunities to user profile
- Generate personalized outreach

## Migration: Loading agent/me Files

Initial load of existing files from `modules/agent/me/`:

| File | content_type | tags |
|------|--------------|------|
| `resume.pdf` | application/pdf | resume |
| `coverletter1.pdf` | application/pdf | cover_letter |
| `coverletter2.pdf` | application/pdf | cover_letter |
| `sample_answers.txt` | text/plain | sample_answers |
| `summary.txt` | text/plain | summary |

**Note**: PDF content should be extracted to text before storage for AI consumption.

## SQLAlchemy Models

```python
class Asset(Base):
    __tablename__ = 'asset'

    id = Column(BigInteger, primary_key=True, autoincrement=True)
    entity_type = Column(String(50), nullable=False)
    entity_id = Column(BigInteger, nullable=False)
    content = Column(Text, nullable=False)
    content_type = Column(String(100), nullable=False)
    filename = Column(String(255))
    checksum = Column(String(64))
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    tags = relationship('AssetTag', back_populates='asset', cascade='all, delete-orphan')

class AssetTag(Base):
    __tablename__ = 'asset_tag'

    id = Column(BigInteger, primary_key=True, autoincrement=True)
    asset_id = Column(BigInteger, ForeignKey('asset.id', ondelete='CASCADE'), nullable=False)
    tag = Column(String(100), nullable=False)

    asset = relationship('Asset', back_populates='tags')
```

## Future Considerations

- **Embeddings**: Add vector column for semantic search
- **Versioning**: Track asset versions over time
- **Access control**: Multi-user support with ownership
