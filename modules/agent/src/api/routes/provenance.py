"""Routes for data provenance API endpoints."""

from typing import Optional, Any
from fastapi import APIRouter, Request, Query, HTTPException
from pydantic import BaseModel
from lib.auth import require_auth
from lib.error_logging import log_errors
from repositories.data_provenance_repository import (
    find_provenance,
    create_provenance,
)

router = APIRouter(prefix="/provenance", tags=["provenance"])


def _provenance_to_dict(p):
    return {
        "id": p.id,
        "entity_type": p.entity_type,
        "entity_id": p.entity_id,
        "field_name": p.field_name,
        "source": p.source,
        "source_url": p.source_url,
        "confidence": float(p.confidence) if p.confidence else None,
        "verified_at": p.verified_at.isoformat() if p.verified_at else None,
        "verified_by": p.verified_by,
        "raw_value": p.raw_value,
        "created_at": p.created_at.isoformat() if p.created_at else None,
    }


class ProvenanceCreate(BaseModel):
    entity_type: str
    entity_id: int
    field_name: str
    source: str
    source_url: Optional[str] = None
    confidence: Optional[float] = None
    raw_value: Optional[Any] = None


@router.get("/{entity_type}/{entity_id}")
@log_errors
@require_auth
async def get_provenance(
    request: Request,
    entity_type: str,
    entity_id: int,
    field_name: Optional[str] = Query(None)
):
    """Get provenance records for an entity."""
    if entity_type not in ("Organization", "Individual"):
        raise HTTPException(status_code=400, detail="Invalid entity_type")

    records = await find_provenance(entity_type, entity_id, field_name)
    return {"provenance": [_provenance_to_dict(r) for r in records]}


@router.post("")
@log_errors
@require_auth
async def create_provenance_route(request: Request, data: ProvenanceCreate):
    """Create a new provenance record."""
    if data.entity_type not in ("Organization", "Individual"):
        raise HTTPException(status_code=400, detail="Invalid entity_type")

    # Get current user ID if available
    user_id = getattr(request.state, "user_id", None)

    record = await create_provenance(
        entity_type=data.entity_type,
        entity_id=data.entity_id,
        field_name=data.field_name,
        source=data.source,
        source_url=data.source_url,
        confidence=data.confidence,
        verified_by=user_id,
        raw_value=data.raw_value
    )
    return {"provenance": _provenance_to_dict(record)}
