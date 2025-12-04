"""Controller layer for provenance routing to services."""

import logging
from typing import Optional, Any

from fastapi import Request

from lib.decorators import handle_http_exceptions
from serializers.provenance import provenance_to_response
from schemas.provenance import ProvenanceCreate
from services.provenance_service import (
    get_provenance as get_provenance_service,
    create_provenance as create_provenance_service,
)

# Re-export for routes
__all__ = ["ProvenanceCreate"]


@handle_http_exceptions
async def get_provenance(entity_type: str, entity_id: int, field_name: Optional[str] = None) -> dict:
    """Get provenance records for an entity."""
    records = await get_provenance_service(entity_type, entity_id, field_name)
    return {"provenance": [provenance_to_response(r) for r in records]}


@handle_http_exceptions
async def create_provenance(request: Request, data: ProvenanceCreate) -> dict:
    """Create a new provenance record."""
    user_id = request.state.user_id
    record = await create_provenance_service(
        data.entity_type, data.entity_id, data.field_name, data.source,
        user_id, data.source_url, data.confidence, data.raw_value
    )
    return {"provenance": provenance_to_response(record)}
