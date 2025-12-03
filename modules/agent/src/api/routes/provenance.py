"""Routes for data provenance API endpoints."""

from typing import Optional

from fastapi import APIRouter, Request, Query

from controllers.provenance_controller import (
    get_provenance,
    create_provenance,
    ProvenanceCreate,
)
from lib.auth import require_auth
from lib.error_logging import log_errors


router = APIRouter(prefix="/provenance", tags=["provenance"])


@router.get("/{entity_type}/{entity_id}")
@log_errors
@require_auth
async def get_provenance_route(
    request: Request,
    entity_type: str,
    entity_id: int,
    field_name: Optional[str] = Query(None),
):
    """Get provenance records for an entity."""
    return await get_provenance(entity_type, entity_id, field_name)


@router.post("")
@log_errors
@require_auth
async def create_provenance_route(request: Request, data: ProvenanceCreate):
    """Create a new provenance record."""
    return await create_provenance(request, data)
