"""Routes for notes API endpoints."""

from fastapi import APIRouter, Request, HTTPException
from pydantic import BaseModel

from lib.auth import require_auth
from lib.error_logging import log_errors
from repositories.note_repository import (
    find_notes_by_entity,
    find_note_by_id,
    create_note,
    update_note,
    delete_note,
)


# Normalize entity_type to PascalCase for consistency with database
ENTITY_TYPE_MAP = {
    "organization": "Organization",
    "individual": "Individual",
    "contact": "Contact",
    "lead": "Lead",
}


def _normalize_entity_type(entity_type: str) -> str:
    """Normalize entity_type to PascalCase."""
    return ENTITY_TYPE_MAP.get(entity_type.lower(), entity_type)


class NoteCreate(BaseModel):
    entity_type: str
    entity_id: int
    content: str


class NoteUpdate(BaseModel):
    content: str


def _note_to_dict(note) -> dict:
    """Convert Note model to dictionary."""
    creator = note.creator
    created_by_name = None
    created_by_email = None
    individual_id = None
    if creator:
        first = creator.first_name or ""
        last = creator.last_name or ""
        created_by_name = f"{first} {last}".strip() or None
        created_by_email = creator.email
        individual_id = creator.individual_id

    return {
        "id": note.id,
        "tenant_id": note.tenant_id,
        "entity_type": note.entity_type,
        "entity_id": note.entity_id,
        "content": note.content,
        "created_by": note.created_by,
        "created_by_name": created_by_name,
        "created_by_email": created_by_email,
        "individual_id": individual_id,
        "created_at": note.created_at.isoformat() if note.created_at else None,
        "updated_at": note.updated_at.isoformat() if note.updated_at else None,
    }


router = APIRouter(prefix="/notes", tags=["notes"])


@router.get("")
@require_auth
@log_errors
async def get_notes_route(
    request: Request,
    entity_type: str,
    entity_id: int,
):
    """Get all notes for a specific entity."""
    tenant_id = request.state.tenant_id
    normalized_type = _normalize_entity_type(entity_type)
    notes = await find_notes_by_entity(tenant_id, normalized_type, entity_id)
    return [_note_to_dict(n) for n in notes]


@router.get("/{note_id}")
@require_auth
@log_errors
async def get_note_route(request: Request, note_id: int):
    """Get a single note by ID."""
    tenant_id = request.state.tenant_id
    note = await find_note_by_id(note_id, tenant_id)
    if not note:
        raise HTTPException(status_code=404, detail="Note not found")
    return _note_to_dict(note)


@router.post("")
@require_auth
@log_errors
async def create_note_route(request: Request, data: NoteCreate):
    """Create a new note."""
    tenant_id = request.state.tenant_id
    user_id = getattr(request.state, "user_id", None)

    note_data = {
        "tenant_id": tenant_id,
        "entity_type": _normalize_entity_type(data.entity_type),
        "entity_id": data.entity_id,
        "content": data.content,
        "created_by": user_id,
    }

    note = await create_note(note_data)
    return _note_to_dict(note)


@router.put("/{note_id}")
@require_auth
@log_errors
async def update_note_route(request: Request, note_id: int, data: NoteUpdate):
    """Update a note."""
    tenant_id = request.state.tenant_id

    note = await update_note(note_id, tenant_id, {"content": data.content})
    if not note:
        raise HTTPException(status_code=404, detail="Note not found")
    return _note_to_dict(note)


@router.delete("/{note_id}")
@require_auth
@log_errors
async def delete_note_route(request: Request, note_id: int):
    """Delete a note."""
    tenant_id = request.state.tenant_id

    deleted = await delete_note(note_id, tenant_id)
    if not deleted:
        raise HTTPException(status_code=404, detail="Note not found")
    return {"status": "deleted"}
