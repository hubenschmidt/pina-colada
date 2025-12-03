"""Service layer for data provenance business logic."""

import logging
from typing import Optional, Any, List

from fastapi import HTTPException

from repositories.data_provenance_repository import (
    find_provenance,
    create_provenance as create_provenance_repo,
    ProvenanceCreate,
)

# Re-export Pydantic models for controllers
__all__ = ["ProvenanceCreate"]

logger = logging.getLogger(__name__)

VALID_ENTITY_TYPES = ("Organization", "Individual")


async def get_provenance(entity_type: str, entity_id: int, field_name: Optional[str] = None) -> List:
    """Get provenance records for an entity."""
    if entity_type not in VALID_ENTITY_TYPES:
        raise HTTPException(status_code=400, detail="Invalid entity_type")
    return await find_provenance(entity_type, entity_id, field_name)


async def create_provenance(
    entity_type: str,
    entity_id: int,
    field_name: str,
    source: str,
    user_id: Optional[int],
    source_url: Optional[str] = None,
    confidence: Optional[float] = None,
    raw_value: Optional[Any] = None,
):
    """Create a new provenance record."""
    if entity_type not in VALID_ENTITY_TYPES:
        raise HTTPException(status_code=400, detail="Invalid entity_type")

    return await create_provenance_repo(
        entity_type=entity_type,
        entity_id=entity_id,
        field_name=field_name,
        source=source,
        source_url=source_url,
        confidence=confidence,
        verified_by=user_id,
        raw_value=raw_value,
    )
