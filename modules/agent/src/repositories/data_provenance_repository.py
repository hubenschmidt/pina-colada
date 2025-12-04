"""Repository layer for data provenance data access."""

from typing import Any, List, Optional
from sqlalchemy import select
from lib.db import async_get_session
from models.DataProvenance import DataProvenance
from schemas.provenance import ProvenanceCreate

__all__ = ["ProvenanceCreate"]


async def find_provenance(
    entity_type: str,
    entity_id: int,
    field_name: Optional[str] = None
) -> List[DataProvenance]:
    """Find provenance records for an entity, optionally filtered by field."""
    async with async_get_session() as session:
        stmt = (
            select(DataProvenance)
            .where(
                DataProvenance.entity_type == entity_type,
                DataProvenance.entity_id == entity_id
            )
            .order_by(DataProvenance.verified_at.desc())
        )
        if field_name:
            stmt = stmt.where(DataProvenance.field_name == field_name)
        result = await session.execute(stmt)
        return list(result.scalars().all())


async def find_provenance_by_id(provenance_id: int) -> Optional[DataProvenance]:
    """Find provenance by ID."""
    async with async_get_session() as session:
        return await session.get(DataProvenance, provenance_id)


async def create_provenance(
    entity_type: str,
    entity_id: int,
    field_name: str,
    source: str,
    source_url: Optional[str] = None,
    confidence: Optional[float] = None,
    verified_by: Optional[int] = None,
    raw_value: Optional[Any] = None
) -> DataProvenance:
    """Create a new provenance record."""
    async with async_get_session() as session:
        provenance = DataProvenance(
            entity_type=entity_type,
            entity_id=entity_id,
            field_name=field_name,
            source=source,
            source_url=source_url,
            confidence=confidence,
            verified_by=verified_by,
            raw_value=raw_value
        )
        session.add(provenance)
        await session.commit()
        await session.refresh(provenance)
        return provenance
