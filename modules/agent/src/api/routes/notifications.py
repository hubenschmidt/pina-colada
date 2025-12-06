"""Routes for notification API endpoints."""

from fastapi import APIRouter, Request, Query

from controllers.notification_controller import (
    get_notifications,
    get_unread_count,
    mark_as_read,
    mark_entity_as_read,
)
from schemas.notification import MarkEntityReadRequest, MarkReadRequest


router = APIRouter(prefix="/notifications", tags=["notifications"])


@router.get("/count")
async def get_unread_count_route(request: Request):
    """Get unread notification count for the current user."""
    return await get_unread_count(request)


@router.get("")
async def get_notifications_route(request: Request, limit: int = Query(default=20, le=100)):
    """Get notifications for the current user."""
    return await get_notifications(request, limit)


@router.post("/mark-read")
async def mark_read_route(request: Request, data: MarkReadRequest):
    """Mark specific notifications as read."""
    return await mark_as_read(request, data)


@router.post("/mark-entity-read")
async def mark_entity_read_route(request: Request, data: MarkEntityReadRequest):
    """Mark all notifications for an entity as read."""
    return await mark_entity_as_read(request, data)
