"""Provenance schemas for API validation."""

from typing import Any, Optional

from pydantic import BaseModel


class ProvenanceCreate(BaseModel):
    entity_type: str
    entity_id: int
    field_name: str
    source: str
    source_url: Optional[str] = None
    confidence: Optional[float] = None
    raw_value: Optional[Any] = None
