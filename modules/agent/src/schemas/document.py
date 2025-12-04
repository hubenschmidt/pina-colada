"""Document schemas for API validation."""

from typing import Optional

from pydantic import BaseModel


class DocumentUpdate(BaseModel):
    description: Optional[str] = None


class EntityLink(BaseModel):
    entity_type: str
    entity_id: int
