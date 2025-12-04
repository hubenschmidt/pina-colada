"""Note schemas for API validation."""

from pydantic import BaseModel


class NoteCreate(BaseModel):
    entity_type: str
    entity_id: int
    content: str


class NoteUpdate(BaseModel):
    content: str
