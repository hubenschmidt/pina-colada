"""Controller layer for notification routing to services."""

import logging

from fastapi import Request

from lib.decorators import handle_http_exceptions
from services.notification_service import (
    MarkReadRequest,
    MarkEntityReadRequest,
    get_unread_count as get_unread_count_service,
    get_notifications as get_notifications_service,
    mark_as_read as mark_as_read_service,
    mark_entity_as_read as mark_entity_as_read_service,
    get_entity_url,
    get_entity_project_id,
)

logger = logging.getLogger(__name__)

# Re-export for routes
__all__ = ["MarkReadRequest", "MarkEntityReadRequest"]


def _get_entity_display_name(comment) -> str:
    """Get display name for the entity a comment belongs to."""
    return f"{comment.commentable_type} #{comment.commentable_id}"


def _notification_to_response(notification) -> dict:
    """Convert notification model to dictionary with nested comment/entity info."""
    comment = notification.comment
    creator = comment.creator if comment else None

    created_by_name = None
    if creator:
        first = creator.first_name or ""
        last = creator.last_name or ""
        created_by_name = f"{first} {last}".strip() or creator.email

    content_preview = ""
    if comment and comment.content:
        content_preview = comment.content[:100] + ("..." if len(comment.content) > 100 else "")

    return {
        "id": notification.id,
        "type": notification.notification_type,
        "is_read": notification.is_read,
        "created_at": notification.created_at.isoformat() if notification.created_at else None,
        "comment": {
            "id": comment.id if comment else None,
            "content": content_preview,
            "created_by_name": created_by_name,
            "created_at": comment.created_at.isoformat() if comment and comment.created_at else None,
        } if comment else None,
        "entity": {
            "type": comment.commentable_type if comment else None,
            "id": comment.commentable_id if comment else None,
            "display_name": _get_entity_display_name(comment) if comment else None,
            "url": get_entity_url(comment.commentable_type, comment.commentable_id, comment.id) if comment else None,
            "project_id": get_entity_project_id(comment.commentable_type, comment.commentable_id) if comment else None,
        } if comment else None,
    }


@handle_http_exceptions
async def get_unread_count(request: Request) -> dict:
    """Get unread notification count for a user."""
    user_id = request.state.user_id
    tenant_id = request.state.tenant_id
    count = await get_unread_count_service(user_id, tenant_id)
    return {"unread_count": count}


@handle_http_exceptions
async def get_notifications(request: Request, limit: int) -> dict:
    """Get notifications for a user."""
    user_id = request.state.user_id
    tenant_id = request.state.tenant_id
    notifications = await get_notifications_service(user_id, tenant_id, limit)
    unread_count = await get_unread_count_service(user_id, tenant_id)
    return {
        "notifications": [_notification_to_response(n) for n in notifications],
        "unread_count": unread_count,
    }


@handle_http_exceptions
async def mark_as_read(request: Request, data: MarkReadRequest) -> dict:
    """Mark specific notifications as read."""
    user_id = request.state.user_id
    tenant_id = request.state.tenant_id
    updated = await mark_as_read_service(data.notification_ids, user_id, tenant_id)
    return {"status": "ok", "updated": updated}


@handle_http_exceptions
async def mark_entity_as_read(request: Request, data: MarkEntityReadRequest) -> dict:
    """Mark all notifications for an entity as read."""
    user_id = request.state.user_id
    tenant_id = request.state.tenant_id
    updated = await mark_entity_as_read_service(user_id, tenant_id, data.entity_type, data.entity_id)
    return {"status": "ok", "updated": updated}
