"""Repository layer for conversation data access."""

import logging
from datetime import datetime, timezone
from typing import Any, Dict, List, Optional
from uuid import UUID

from sqlalchemy import and_, func, select, update
from sqlalchemy.orm import selectinload
from lib.db import async_get_session
from models.Conversation import Conversation, ConversationMessage

logger = logging.getLogger(__name__)


async def list_conversations(
    user_id: int,
    tenant_id: int,
    limit: int = 50,
    include_archived: bool = False,
) -> List[Conversation]:
    """List conversations for a user, ordered by updated_at desc."""
    async with async_get_session() as session:
        conditions = [
            Conversation.user_id == user_id,
            Conversation.tenant_id == tenant_id,
        ]
        if not include_archived:
            conditions.append(Conversation.archived_at.is_(None))

        stmt = (
            select(Conversation)
            .where(and_(*conditions))
            .order_by(Conversation.updated_at.desc())
            .limit(limit)
        )
        result = await session.execute(stmt)
        return list(result.scalars().all())


async def list_tenant_conversations(
    tenant_id: int,
    search: Optional[str] = None,
    limit: int = 100,
    offset: int = 0,
    include_archived: bool = False,
) -> tuple[List[Conversation], int]:
    """List all conversations for a tenant with search and pagination."""
    async with async_get_session() as session:
        conditions = [Conversation.tenant_id == tenant_id]

        if not include_archived:
            conditions.append(Conversation.archived_at.is_(None))

        if search:
            conditions.append(Conversation.title.ilike(f"%{search}%"))

        # Count total
        count_stmt = select(func.count(Conversation.id)).where(and_(*conditions))
        count_result = await session.execute(count_stmt)
        total = count_result.scalar() or 0

        # Fetch with eager loading
        stmt = (
            select(Conversation)
            .options(selectinload(Conversation.created_by), selectinload(Conversation.updated_by))
            .where(and_(*conditions))
            .order_by(Conversation.updated_at.desc())
            .offset(offset)
            .limit(limit)
        )
        result = await session.execute(stmt)
        conversations = list(result.scalars().all())

        return conversations, total


async def get_conversation(thread_id: UUID) -> Optional[Conversation]:
    """Get conversation by thread_id."""
    async with async_get_session() as session:
        stmt = select(Conversation).where(Conversation.thread_id == thread_id)
        result = await session.execute(stmt)
        return result.scalar_one_or_none()


async def get_conversation_with_messages(thread_id: UUID) -> Optional[Dict[str, Any]]:
    """Get conversation with all messages."""
    async with async_get_session() as session:
        stmt = select(Conversation).where(Conversation.thread_id == thread_id)
        result = await session.execute(stmt)
        conversation = result.scalar_one_or_none()

        if not conversation:
            return None

        msg_stmt = (
            select(ConversationMessage)
            .where(ConversationMessage.conversation_id == conversation.id)
            .order_by(ConversationMessage.created_at.asc())
        )
        msg_result = await session.execute(msg_stmt)
        messages = list(msg_result.scalars().all())

        return {
            "conversation": conversation,
            "messages": messages,
        }


async def create_conversation(
    user_id: int,
    tenant_id: int,
    thread_id: UUID,
    title: Optional[str] = None,
) -> Conversation:
    """Create a new conversation."""
    async with async_get_session() as session:
        conversation = Conversation(
            user_id=user_id,
            tenant_id=tenant_id,
            thread_id=thread_id,
            title=title,
            created_by_id=user_id,
            updated_by_id=user_id,
        )
        session.add(conversation)
        await session.commit()
        await session.refresh(conversation)
        return conversation


async def update_conversation_title(thread_id: UUID, title: str) -> Optional[Conversation]:
    """Update conversation title."""
    async with async_get_session() as session:
        stmt = select(Conversation).where(Conversation.thread_id == thread_id)
        result = await session.execute(stmt)
        conversation = result.scalar_one_or_none()

        if not conversation:
            return None

        conversation.title = title
        conversation.updated_at = datetime.now(timezone.utc)
        await session.commit()
        await session.refresh(conversation)
        return conversation


async def touch_conversation(thread_id: UUID) -> None:
    """Update conversation's updated_at timestamp."""
    async with async_get_session() as session:
        stmt = (
            update(Conversation)
            .where(Conversation.thread_id == thread_id)
            .values(updated_at=datetime.now(timezone.utc))
        )
        await session.execute(stmt)
        await session.commit()


async def archive_conversation(thread_id: UUID) -> bool:
    """Archive a conversation (soft delete)."""
    async with async_get_session() as session:
        stmt = select(Conversation).where(Conversation.thread_id == thread_id)
        result = await session.execute(stmt)
        conversation = result.scalar_one_or_none()

        if not conversation:
            return False

        conversation.archived_at = datetime.now(timezone.utc)
        await session.commit()
        return True


async def unarchive_conversation(thread_id: UUID) -> bool:
    """Unarchive a conversation."""
    async with async_get_session() as session:
        stmt = select(Conversation).where(Conversation.thread_id == thread_id)
        result = await session.execute(stmt)
        conversation = result.scalar_one_or_none()

        if not conversation:
            return False

        conversation.archived_at = None
        await session.commit()
        return True


async def delete_conversation_permanent(thread_id: UUID) -> bool:
    """Permanently delete a conversation (must be archived first)."""
    async with async_get_session() as session:
        stmt = select(Conversation).where(Conversation.thread_id == thread_id)
        result = await session.execute(stmt)
        conversation = result.scalar_one_or_none()

        if not conversation:
            return False

        if not conversation.archived_at:
            logger.warning(f"Cannot permanently delete non-archived conversation: {thread_id}")
            return False

        await session.delete(conversation)
        await session.commit()
        return True


async def add_message(
    conversation_id: int,
    role: str,
    content: str,
    token_usage: Optional[Dict[str, int]] = None,
) -> ConversationMessage:
    """Add a message to a conversation."""
    async with async_get_session() as session:
        message = ConversationMessage(
            conversation_id=conversation_id,
            role=role,
            content=content,
            token_usage=token_usage,
        )
        session.add(message)
        await session.commit()
        await session.refresh(message)
        return message


async def get_or_create_conversation(
    user_id: int,
    tenant_id: int,
    thread_id: UUID,
) -> Conversation:
    """Get existing conversation or create new one."""
    existing = await get_conversation(thread_id)
    if existing:
        return existing
    return await create_conversation(user_id, tenant_id, thread_id)
