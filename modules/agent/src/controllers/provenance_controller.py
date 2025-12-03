"""Controller layer for provenance routing to services."""

import logging
from typing import Optional, Any

from fastapi import Request

from lib.decorators import handle_http_exceptions
from repositories.data_provenance_repository import ProvenanceCreate
from services.provenance_service import (
    get_provenance as get_provenance_service,
    create_provenance as create_provenance_service,
)

# Re-export for routes
__all__ = ["ProvenanceCreate"]

logger = logging.getLogger(__name__)


def _provenance_to_response(p) -> dict:
    """Convert provenance model to response dict."""
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
        "updated_at": p.updated_at.isoformat() if p.updated_at else None,
    }


@handle_http_exceptions
async def get_provenance(entity_type: str, entity_id: int, field_name: Optional[str] = None) -> dict:
    """Get provenance records for an entity."""
    records = await get_provenance_service(entity_type, entity_id, field_name)
    return {"provenance": [_provenance_to_response(r) for r in records]}


@handle_http_exceptions
async def create_provenance(request: Request, data: ProvenanceCreate) -> dict:
    """Create a new provenance record."""
    user_id = request.state.user_id
    record = await create_provenance_service(
        data.entity_type, data.entity_id, data.field_name, data.source,
        user_id, data.source_url, data.confidence, data.raw_value
    )
    return {"provenance": _provenance_to_response(record)}
