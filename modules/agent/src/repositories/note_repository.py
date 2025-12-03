"""Repository layer for note data access."""

import logging
from typing import List, Optional, Dict, Any

from pydantic import BaseModel
from sqlalchemy import select, and_
from sqlalchemy.orm import joinedload

from lib.db import async_get_session
from models.Note import Note

logger = logging.getLogger(__name__)


# Pydantic models

class NoteCreate(BaseModel):
    entity_type: str
    entity_id: int
    content: str


class NoteUpdate(BaseModel):
    content: str


async def find_notes_by_entity(
    tenant_id: int, entity_type: str, entity_id: int
) -> List[Note]:
    """Find all notes for a specific entity."""
    async with async_get_session() as session:
        stmt = (
            select(Note)
            .options(joinedload(Note.creator))
            .where(
                and_(
                    Note.tenant_id == tenant_id,
                    Note.entity_type == entity_type,
                    Note.entity_id == entity_id,
                )
            )
            .order_by(Note.created_at.desc())
        )
        result = await session.execute(stmt)
        return list(result.unique().scalars().all())


async def find_note_by_id(note_id: int, tenant_id: int) -> Optional[Note]:
    """Find a note by ID within a tenant."""
    async with async_get_session() as session:
        stmt = (
            select(Note)
            .options(joinedload(Note.creator))
            .where(and_(Note.id == note_id, Note.tenant_id == tenant_id))
        )
        result = await session.execute(stmt)
        return result.unique().scalar_one_or_none()


async def create_note(data: Dict[str, Any]) -> Note:
    """Create a new note."""
    async with async_get_session() as session:
        try:
            note = Note(**data)
            session.add(note)
            await session.commit()
            await session.refresh(note)
            # Eager load the creator relationship
            stmt = (
                select(Note)
                .options(joinedload(Note.creator))
                .where(Note.id == note.id)
            )
            result = await session.execute(stmt)
            return result.unique().scalar_one()
        except Exception as e:
            await session.rollback()
            logger.error(f"Failed to create note: {e}")
            raise


async def update_note(note_id: int, tenant_id: int, data: Dict[str, Any]) -> Optional[Note]:
    """Update a note."""
    async with async_get_session() as session:
        try:
            stmt = select(Note).where(
                and_(Note.id == note_id, Note.tenant_id == tenant_id)
            )
            result = await session.execute(stmt)
            note = result.scalar_one_or_none()
            if not note:
                return None
            for key, value in data.items():
                if hasattr(note, key):
                    setattr(note, key, value)
            await session.commit()
            # Eager load the creator relationship
            stmt = (
                select(Note)
                .options(joinedload(Note.creator))
                .where(Note.id == note_id)
            )
            result = await session.execute(stmt)
            return result.unique().scalar_one()
        except Exception as e:
            await session.rollback()
            logger.error(f"Failed to update note: {e}")
            raise


async def delete_note(note_id: int, tenant_id: int) -> bool:
    """Delete a note."""
    async with async_get_session() as session:
        try:
            stmt = select(Note).where(
                and_(Note.id == note_id, Note.tenant_id == tenant_id)
            )
            result = await session.execute(stmt)
            note = result.scalar_one_or_none()
            if not note:
                return False
            await session.delete(note)
            await session.commit()
            return True
        except Exception as e:
            await session.rollback()
            logger.error(f"Failed to delete note: {e}")
            raise
