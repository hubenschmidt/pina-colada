"""Service layer for note business logic."""

import logging
from typing import Optional, Dict, Any, List

from fastapi import HTTPException

from repositories.note_repository import (
    find_notes_by_entity,
    find_note_by_id,
    create_note as create_note_repo,
    update_note as update_note_repo,
    delete_note as delete_note_repo,
)

logger = logging.getLogger(__name__)

ENTITY_TYPE_MAP = {
    "organization": "Organization",
    "individual": "Individual",
    "contact": "Contact",
    "lead": "Lead",
}


def _normalize_entity_type(entity_type: str) -> str:
    """Normalize entity_type to PascalCase."""
    return ENTITY_TYPE_MAP.get(entity_type.lower(), entity_type)


async def get_notes_by_entity(tenant_id: int, entity_type: str, entity_id: int) -> List:
    """Get all notes for a specific entity."""
    normalized_type = _normalize_entity_type(entity_type)
    return await find_notes_by_entity(tenant_id, normalized_type, entity_id)


async def get_note(note_id: int, tenant_id: int):
    """Get a single note by ID."""
    note = await find_note_by_id(note_id, tenant_id)
    if not note:
        raise HTTPException(status_code=404, detail="Note not found")
    return note


async def create_note(
    tenant_id: int,
    entity_type: str,
    entity_id: int,
    content: str,
    user_id: Optional[int],
):
    """Create a new note."""
    note_data = {
        "tenant_id": tenant_id,
        "entity_type": _normalize_entity_type(entity_type),
        "entity_id": entity_id,
        "content": content,
        "created_by": user_id,
    }
    return await create_note_repo(note_data)


async def update_note(
    note_id: int,
    tenant_id: int,
    content: str,
    user_id: Optional[int],
):
    """Update a note."""
    note = await update_note_repo(note_id, tenant_id, {"content": content, "updated_by": user_id})
    if not note:
        raise HTTPException(status_code=404, detail="Note not found")
    return note


async def delete_note(note_id: int, tenant_id: int):
    """Delete a note."""
    deleted = await delete_note_repo(note_id, tenant_id)
    if not deleted:
        raise HTTPException(status_code=404, detail="Note not found")
    return True
