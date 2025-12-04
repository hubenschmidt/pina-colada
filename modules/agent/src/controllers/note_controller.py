"""Controller layer for note routing to services."""

from typing import List

from fastapi import Request

from lib.decorators import handle_http_exceptions
from serializers.note import note_to_response
from schemas.note import NoteCreate, NoteUpdate
from services.note_service import (
    get_notes_by_entity as get_notes_service,
    get_note as get_note_service,
    create_note as create_note_service,
    update_note as update_note_service,
    delete_note as delete_note_service,
)

# Re-export for routes
__all__ = ["NoteCreate", "NoteUpdate"]


@handle_http_exceptions
async def get_notes(request: Request, entity_type: str, entity_id: int) -> List[dict]:
    """Get all notes for a specific entity."""
    tenant_id = request.state.tenant_id
    notes = await get_notes_service(tenant_id, entity_type, entity_id)
    return [note_to_response(n) for n in notes]


@handle_http_exceptions
async def get_note(request: Request, note_id: int) -> dict:
    """Get a single note by ID."""
    tenant_id = request.state.tenant_id
    note = await get_note_service(note_id, tenant_id)
    return note_to_response(note)


@handle_http_exceptions
async def create_note(request: Request, data: NoteCreate) -> dict:
    """Create a new note."""
    tenant_id = request.state.tenant_id
    user_id = request.state.user_id
    note = await create_note_service(tenant_id, data.entity_type, data.entity_id, data.content, user_id)
    return note_to_response(note)


@handle_http_exceptions
async def update_note(request: Request, note_id: int, data: NoteUpdate) -> dict:
    """Update a note."""
    tenant_id = request.state.tenant_id
    user_id = request.state.user_id
    note = await update_note_service(note_id, tenant_id, data.content, user_id)
    return note_to_response(note)


@handle_http_exceptions
async def delete_note(request: Request, note_id: int) -> dict:
    """Delete a note."""
    tenant_id = request.state.tenant_id
    await delete_note_service(note_id, tenant_id)
    return {"status": "deleted"}
