"""Repository layer for comment data access."""

import logging
from typing import List, Optional, Dict, Any

from sqlalchemy import select, and_
from sqlalchemy.orm import joinedload

from models.Comment import Comment
from lib.db import async_get_session

logger = logging.getLogger(__name__)


async def find_comments_by_entity(
    tenant_id: int, commentable_type: str, commentable_id: int
) -> List[Comment]:
    """Find all comments for a specific entity."""
    async with async_get_session() as session:
        stmt = (
            select(Comment)
            .options(joinedload(Comment.creator))
            .where(
                and_(
                    Comment.tenant_id == tenant_id,
                    Comment.commentable_type == commentable_type,
                    Comment.commentable_id == commentable_id,
                )
            )
            .order_by(Comment.created_at.asc())
        )
        result = await session.execute(stmt)
        return list(result.unique().scalars().all())


async def find_comment_by_id(comment_id: int, tenant_id: int) -> Optional[Comment]:
    """Find a comment by ID within a tenant."""
    async with async_get_session() as session:
        stmt = (
            select(Comment)
            .options(joinedload(Comment.creator))
            .where(and_(Comment.id == comment_id, Comment.tenant_id == tenant_id))
        )
        result = await session.execute(stmt)
        return result.unique().scalar_one_or_none()


async def create_comment(data: Dict[str, Any]) -> Comment:
    """Create a new comment."""
    async with async_get_session() as session:
        try:
            comment = Comment(**data)
            session.add(comment)
            await session.commit()
            await session.refresh(comment)
            # Eager load the creator relationship
            stmt = (
                select(Comment)
                .options(joinedload(Comment.creator))
                .where(Comment.id == comment.id)
            )
            result = await session.execute(stmt)
            return result.unique().scalar_one()
        except Exception as e:
            await session.rollback()
            logger.error(f"Failed to create comment: {e}")
            raise


async def update_comment(comment_id: int, tenant_id: int, data: Dict[str, Any]) -> Optional[Comment]:
    """Update a comment."""
    async with async_get_session() as session:
        try:
            stmt = select(Comment).where(
                and_(Comment.id == comment_id, Comment.tenant_id == tenant_id)
            )
            result = await session.execute(stmt)
            comment = result.scalar_one_or_none()
            if not comment:
                return None
            for key, value in data.items():
                if hasattr(comment, key):
                    setattr(comment, key, value)
            await session.commit()
            await session.refresh(comment)
            return comment
        except Exception as e:
            await session.rollback()
            logger.error(f"Failed to update comment: {e}")
            raise


async def delete_comment(comment_id: int, tenant_id: int) -> bool:
    """Delete a comment."""
    async with async_get_session() as session:
        try:
            stmt = select(Comment).where(
                and_(Comment.id == comment_id, Comment.tenant_id == tenant_id)
            )
            result = await session.execute(stmt)
            comment = result.scalar_one_or_none()
            if not comment:
                return False
            await session.delete(comment)
            await session.commit()
            return True
        except Exception as e:
            await session.rollback()
            logger.error(f"Failed to delete comment: {e}")
            raise
