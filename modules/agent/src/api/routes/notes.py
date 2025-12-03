"""Routes for notes API endpoints."""

from fastapi import APIRouter, Request

from controllers.note_controller import (
    get_notes,
    get_note,
    create_note,
    update_note,
    delete_note,
    NoteCreate,
    NoteUpdate,
)
from lib.auth import require_auth
from lib.error_logging import log_errors


router = APIRouter(prefix="/notes", tags=["notes"])


@router.get("")
@require_auth
@log_errors
async def get_notes_route(request: Request, entity_type: str, entity_id: int):
    """Get all notes for a specific entity."""
    return await get_notes(request, entity_type, entity_id)


@router.get("/{note_id}")
@require_auth
@log_errors
async def get_note_route(request: Request, note_id: int):
    """Get a single note by ID."""
    return await get_note(request, note_id)


@router.post("")
@require_auth
@log_errors
async def create_note_route(request: Request, data: NoteCreate):
    """Create a new note."""
    return await create_note(request, data)


@router.put("/{note_id}")
@require_auth
@log_errors
async def update_note_route(request: Request, note_id: int, data: NoteUpdate):
    """Update a note."""
    return await update_note(request, note_id, data)


@router.delete("/{note_id}")
@require_auth
@log_errors
async def delete_note_route(request: Request, note_id: int):
    """Delete a note."""
    return await delete_note(request, note_id)
