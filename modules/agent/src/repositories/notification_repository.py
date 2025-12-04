"""Repository layer for notification data access."""

import logging
from datetime import datetime, timedelta
from typing import List, Optional
from sqlalchemy import and_, delete, func, select, update
from sqlalchemy.orm import joinedload
from lib.db import async_get_session
from models.Comment import Comment
from models.CommentNotification import CommentNotification
from schemas.notification import MarkEntityReadRequest, MarkReadRequest

__all__ = ["MarkReadRequest", "MarkEntityReadRequest"]

logger = logging.getLogger(__name__)


async def get_unread_count(user_id: int, tenant_id: int) -> int:
    """Get count of unread notifications for a user."""
    async with async_get_session() as session:
        stmt = (
            select(func.count(CommentNotification.id))
            .where(
                and_(
                    CommentNotification.user_id == user_id,
                    CommentNotification.tenant_id == tenant_id,
                    CommentNotification.is_read == False,
                )
            )
        )
        result = await session.execute(stmt)
        return result.scalar() or 0


async def get_notifications(
    user_id: int, tenant_id: int, limit: int = 20
) -> List[CommentNotification]:
    """Get notifications with comment details for a user."""
    async with async_get_session() as session:
        stmt = (
            select(CommentNotification)
            .options(
                joinedload(CommentNotification.comment).joinedload(Comment.creator)
            )
            .where(
                and_(
                    CommentNotification.user_id == user_id,
                    CommentNotification.tenant_id == tenant_id,
                )
            )
            .order_by(CommentNotification.created_at.desc())
            .limit(limit)
        )
        result = await session.execute(stmt)
        return list(result.unique().scalars().all())


async def create_notification(
    tenant_id: int,
    user_id: int,
    comment_id: int,
    notification_type: str,
) -> Optional[CommentNotification]:
    """Create a notification. Returns None if duplicate (user already notified for this comment)."""
    async with async_get_session() as session:
        try:
            notification = CommentNotification(
                tenant_id=tenant_id,
                user_id=user_id,
                comment_id=comment_id,
                notification_type=notification_type,
            )
            session.add(notification)
            await session.commit()
            await session.refresh(notification)
            return notification
        except Exception as e:
            await session.rollback()
            # Unique constraint violation - user already notified for this comment
            if "unique" in str(e).lower():
                logger.debug(f"Duplicate notification skipped: user={user_id}, comment={comment_id}")
                return None
            logger.error(f"Failed to create notification: {e}")
            raise


async def mark_as_read(notification_ids: List[int], user_id: int, tenant_id: int) -> int:
    """Mark specific notifications as read. Returns count of updated rows."""
    async with async_get_session() as session:
        try:
            stmt = (
                update(CommentNotification)
                .where(
                    and_(
                        CommentNotification.id.in_(notification_ids),
                        CommentNotification.user_id == user_id,
                        CommentNotification.tenant_id == tenant_id,
                        CommentNotification.is_read == False,
                    )
                )
                .values(is_read=True, read_at=datetime.utcnow())
            )
            result = await session.execute(stmt)
            await session.commit()
            return result.rowcount
        except Exception as e:
            await session.rollback()
            logger.error(f"Failed to mark notifications as read: {e}")
            raise


async def mark_entity_as_read(
    user_id: int, tenant_id: int, entity_type: str, entity_id: int
) -> int:
    """Mark all notifications for an entity as read. Returns count of updated rows."""
    async with async_get_session() as session:
        try:
            # Get comment IDs for this entity
            comment_subq = (
                select(Comment.id)
                .where(
                    and_(
                        Comment.commentable_type == entity_type,
                        Comment.commentable_id == entity_id,
                    )
                )
            )

            stmt = (
                update(CommentNotification)
                .where(
                    and_(
                        CommentNotification.user_id == user_id,
                        CommentNotification.tenant_id == tenant_id,
                        CommentNotification.comment_id.in_(comment_subq),
                        CommentNotification.is_read == False,
                    )
                )
                .values(is_read=True, read_at=datetime.utcnow())
            )
            result = await session.execute(stmt)
            await session.commit()
            return result.rowcount
        except Exception as e:
            await session.rollback()
            logger.error(f"Failed to mark entity notifications as read: {e}")
            raise


async def get_thread_participants(
    tenant_id: int, entity_type: str, entity_id: int, exclude_user_id: Optional[int] = None
) -> List[int]:
    """Get all user IDs who have commented on an entity (for thread activity notifications)."""
    async with async_get_session() as session:
        stmt = (
            select(Comment.created_by)
            .where(
                and_(
                    Comment.tenant_id == tenant_id,
                    Comment.commentable_type == entity_type,
                    Comment.commentable_id == entity_id,
                    Comment.created_by.isnot(None),
                )
            )
            .distinct()
        )

        if exclude_user_id:
            stmt = stmt.where(Comment.created_by != exclude_user_id)

        result = await session.execute(stmt)
        return [row[0] for row in result.fetchall()]


async def cleanup_old_notifications(days: int = 30) -> int:
    """Delete notifications older than specified days. Returns count of deleted rows."""
    async with async_get_session() as session:
        try:
            cutoff = datetime.utcnow() - timedelta(days=days)
            stmt = delete(CommentNotification).where(
                CommentNotification.created_at < cutoff
            )
            result = await session.execute(stmt)
            await session.commit()
            return result.rowcount
        except Exception as e:
            await session.rollback()
            logger.error(f"Failed to cleanup old notifications: {e}")
            raise
