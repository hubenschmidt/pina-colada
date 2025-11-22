# Prompt Overrides - Tenant Customization Layer

## Overview

This spec adds a database-driven customization layer that allows tenant-specific prompt overrides without exposing core agent logic. Core prompts remain in code; only customizable portions are stored in the database.

**Prerequisites:** `specialized-evaluators.md` must be implemented first (centralized prompts in `agent/prompts/`).

---

## Design Principle

- **Core agent behavior = Code** (protected IP)
- **Tenant/product customizations = Data** (API-accessible)

This approach:
- Protects core application logic from API exposure
- Allows tenant-specific customizations within guardrails
- Supports runtime configuration without code redeployment

---

## Database Model

```python
# models/PromptOverride.py
from sqlalchemy import Column, Text, DateTime, BigInteger, Boolean, ForeignKey, func
from models import Base


class PromptOverride(Base):
    """PromptOverride SQLAlchemy model for tenant customizations."""

    __tablename__ = "PromptOverride"

    id = Column(BigInteger, primary_key=True, autoincrement=True)
    tenant_id = Column(BigInteger, ForeignKey("Tenant.id", ondelete="CASCADE"), nullable=True)  # null = global default
    node_type = Column(Text, nullable=False)       # evaluator, worker, orchestrator
    node_name = Column(Text, nullable=False)       # career_evaluator, job_hunter, etc.

    # Customizable fields (layered on top of core prompts)
    additional_instructions = Column(Text, nullable=True)  # extra context/instructions
    custom_criteria = Column(Text, nullable=True)          # additional evaluation criteria
    tone = Column(Text, nullable=True)                     # professional, casual, formal

    is_active = Column(Boolean, default=True, nullable=False)

    # Audit fields
    created_at = Column(DateTime(timezone=True), server_default=func.now(), nullable=False)
    updated_at = Column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now(), nullable=False)
    updated_by = Column(Text, nullable=False)


class PromptOverrideHistory(Base):
    """PromptOverrideHistory SQLAlchemy model for audit trail."""

    __tablename__ = "PromptOverrideHistory"

    id = Column(BigInteger, primary_key=True, autoincrement=True)
    override_id = Column(BigInteger, ForeignKey("PromptOverride.id", ondelete="CASCADE"), nullable=False)
    tenant_id = Column(BigInteger, nullable=True)
    node_type = Column(Text, nullable=False)
    node_name = Column(Text, nullable=False)
    additional_instructions = Column(Text, nullable=True)
    custom_criteria = Column(Text, nullable=True)
    tone = Column(Text, nullable=True)
    changed_at = Column(DateTime(timezone=True), server_default=func.now(), nullable=False)
    changed_by = Column(Text, nullable=False)
    change_type = Column(Text, nullable=False)  # create, update, delete
```

**Supported Nodes:**
- **Evaluators**: `career_evaluator`, `scraper_evaluator`, `general_evaluator`
- **Workers**: `job_hunter`, `scraper`, `worker`, `cover_letter_writer`
- **Orchestrator**: `orchestrator`

---

## CRUD API Endpoints

All endpoints require existing app authentication. These expose **only customization overrides**, not core prompts.

```
GET    /prompt-overrides                   # List overrides (filter by ?node_type=evaluator&tenant_id=1)
GET    /prompt-overrides/{id}              # Get single override
POST   /prompt-overrides                   # Create new override
PUT    /prompt-overrides/{id}              # Update override
DELETE /prompt-overrides/{id}              # Soft delete (set is_active=False)
```

**Request/Response Schema:**
```python
class PromptOverrideCreate(BaseModel):
    tenant_id: Optional[int]              # null for global defaults
    node_type: str                        # evaluator, worker, orchestrator
    node_name: str                        # career_evaluator, job_hunter, etc.
    additional_instructions: Optional[str]
    custom_criteria: Optional[str]
    tone: Optional[str]

class PromptOverrideUpdate(BaseModel):
    additional_instructions: Optional[str]
    custom_criteria: Optional[str]
    tone: Optional[str]
    is_active: Optional[bool]

class PromptOverrideResponse(BaseModel):
    id: int
    tenant_id: Optional[int]
    node_type: str
    node_name: str
    additional_instructions: Optional[str]
    custom_criteria: Optional[str]
    tone: Optional[str]
    is_active: bool
    updated_at: datetime
    updated_by: str
```

---

## Override Resolution

When building a prompt, resolve overrides in priority order: Tenant → Global → None

```python
# orchestrator.py or dedicated loader
async def get_prompt_overrides(tenant_id: int, node_type: str, node_name: str) -> dict:
    """Get prompt overrides with tenant > global fallback"""
    from lib.db import async_get_session
    from models.PromptOverride import PromptOverride
    from sqlalchemy import select

    async with async_get_session() as session:
        # Try tenant-specific first
        stmt = select(PromptOverride).where(
            PromptOverride.tenant_id == tenant_id,
            PromptOverride.node_type == node_type,
            PromptOverride.node_name == node_name,
            PromptOverride.is_active == True
        )
        result = await session.execute(stmt)
        override = result.scalar_one_or_none()

        # Fall back to global (tenant_id = null)
        if not override:
            stmt = select(PromptOverride).where(
                PromptOverride.tenant_id.is_(None),
                PromptOverride.node_type == node_type,
                PromptOverride.node_name == node_name,
                PromptOverride.is_active == True
            )
            result = await session.execute(stmt)
            override = result.scalar_one_or_none()

        if not override:
            return {}

        return {
            "id": override.id,  # For Langfuse trace tagging
            "additional_instructions": override.additional_instructions,
            "custom_criteria": override.custom_criteria,
            "tone": override.tone,
        }
```

---

## Integration with Centralized Prompts

Update the prompt builder functions in `agent/prompts/` to accept overrides:

```python
# agent/prompts/evaluator_prompts.py

def build_career_evaluator_prompt(resume_context: str, overrides: Optional[dict] = None) -> str:
    """Build career evaluator system prompt with optional overrides."""
    sections = [CAREER_BASE]

    if resume_context:
        sections.append(...)
        sections.append(CAREER_CRITICAL_DISTINCTION)

    sections.append(CAREER_CHECKS)

    # Apply tenant/product customizations
    if overrides:
        if overrides.get("additional_instructions"):
            sections.append(f"ADDITIONAL INSTRUCTIONS:\n{overrides['additional_instructions']}")
        if overrides.get("custom_criteria"):
            sections.append(f"CUSTOM CRITERIA:\n{overrides['custom_criteria']}")

    return "\n\n".join(sections)
```

---

## Postman Collection

Add to `modules/agent/postman_collection.json` for API testing:

```json
{
  "name": "Prompt Overrides",
  "item": [
    {
      "name": "List all overrides",
      "request": {
        "method": "GET",
        "url": "{{base_url}}/prompt-overrides?node_type=evaluator&tenant_id=1"
      }
    },
    {
      "name": "Get override by ID",
      "request": {
        "method": "GET",
        "url": "{{base_url}}/prompt-overrides/1"
      }
    },
    {
      "name": "Create override",
      "request": {
        "method": "POST",
        "url": "{{base_url}}/prompt-overrides",
        "body": {
          "mode": "raw",
          "raw": "{\"tenant_id\": 1, \"node_type\": \"evaluator\", \"node_name\": \"career_evaluator\", \"additional_instructions\": \"Focus on tech industry roles\", \"custom_criteria\": \"Must include salary range\"}"
        }
      }
    },
    {
      "name": "Update override",
      "request": {
        "method": "PUT",
        "url": "{{base_url}}/prompt-overrides/1",
        "body": {
          "mode": "raw",
          "raw": "{\"additional_instructions\": \"Updated instructions...\"}"
        }
      }
    },
    {
      "name": "Delete override",
      "request": {
        "method": "DELETE",
        "url": "{{base_url}}/prompt-overrides/1"
      }
    }
  ]
}
```

---

## File Structure

```
modules/agent/src/
├── models/
│   └── PromptOverride.py              # PromptOverride + PromptOverrideHistory models
├── repositories/
│   └── prompt_override_repository.py  # Data access layer
├── controllers/
│   └── prompt_override_controller.py  # Controller layer
└── api/routes/
    └── prompt_overrides.py            # CRUD API endpoints
```

---

## Security Considerations

1. **Core prompt protection**: Base system prompts are code, not exposed via API
2. **Authentication**: All CRUD endpoints require existing app auth
3. **Input validation**: Validate override content (max length, no injection patterns)
4. **Soft deletes**: DELETE sets `is_active=False` to preserve audit history
5. **Rate limiting**: Consider rate limits on override update endpoints
6. **Override sanitization**: Sanitize user-provided overrides before storage
7. **Tenant isolation**: Overrides are scoped to tenant_id, enforced at query level

---

## Observability Integration

Tag all Langfuse traces with `override_id` to track which customization produced each output.

```python
# In evaluator when invoking LLM
overrides = await get_prompt_overrides(tenant_id, "evaluator", evaluator_name)

langfuse_handler = get_langfuse_handler()
if langfuse_handler and overrides.get("id"):
    langfuse_handler.trace.update(
        metadata={"prompt_override_id": overrides["id"]}
    )
```

**Benefits:**
- Debug regressions by identifying which override caused issues
- A/B test customizations by comparing traces with different `override_id` values
- Measure tenant-specific prompt performance over time

---

## Benefits

1. **Protected core logic**: Base prompts stay in code, not exposed via API
2. **Tenant customization**: Runtime overrides without code redeployment
3. **Audit trail**: Track all override changes with timestamps and user attribution
4. **Multi-tenant support**: Different tenants can have different customizations
5. **Guardrails**: Only specific fields are customizable, not entire prompts
