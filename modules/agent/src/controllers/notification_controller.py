"""Controller layer for notification routing to services."""

from fastapi import Request

from lib.decorators import handle_http_exceptions
from serializers.notification import notification_to_response
from services.notification_service import (
    MarkReadRequest,
    MarkEntityReadRequest,
    get_unread_count as get_unread_count_service,
    get_notifications as get_notifications_service,
    mark_as_read as mark_as_read_service,
    mark_entity_as_read as mark_entity_as_read_service,
)

# Re-export for routes
__all__ = ["MarkReadRequest", "MarkEntityReadRequest"]


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
        "notifications": [notification_to_response(n) for n in notifications],
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
