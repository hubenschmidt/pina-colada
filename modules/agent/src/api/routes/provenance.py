"""Routes for data provenance API endpoints."""

from typing import Optional

from fastapi import APIRouter, Request, Query

from controllers.provenance_controller import (
    create_provenance,
    get_provenance,
)
from schemas.provenance import ProvenanceCreate


router = APIRouter(prefix="/provenance", tags=["provenance"])


@router.get("/{entity_type}/{entity_id}")
async def get_provenance_route(
    request: Request,
    entity_type: str,
    entity_id: int,
    field_name: Optional[str] = Query(None),
):
    """Get provenance records for an entity."""
    return await get_provenance(entity_type, entity_id, field_name)


@router.post("")
async def create_provenance_route(request: Request, data: ProvenanceCreate):
    """Create a new provenance record."""
    return await create_provenance(request, data)
