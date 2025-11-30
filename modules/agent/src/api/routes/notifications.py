"""Routes for notification API endpoints."""

from typing import List

from fastapi import APIRouter, Request, Query
from pydantic import BaseModel
from sqlalchemy import text

from lib.auth import require_auth
from lib.error_logging import log_errors
from lib.db import get_session
from repositories.notification_repository import (
    get_unread_count as repo_get_unread_count,
    get_notifications as repo_get_notifications,
    mark_as_read as repo_mark_as_read,
    mark_entity_as_read as repo_mark_entity_as_read,
)


# Entity type to URL path mapping
ENTITY_URL_MAP = {
    "Lead": "/leads",
    "Deal": "/deals",
    "Task": "/tasks",
    "Individual": "/accounts/individuals",
    "Organization": "/accounts/organizations",
}


def _get_entity_project_id(entity_type: str, entity_id: int) -> int | None:
    """Get project_id for an entity. Returns first project if multiple."""
    if not entity_type or not entity_id:
        return None

    session = get_session()
    try:
        # Direct project_id column
        if entity_type == "Deal":
            result = session.execute(
                text('SELECT project_id FROM "Deal" WHERE id = :id'),
                {"id": entity_id}
            ).fetchone()
            return result[0] if result else None

        # Tasks derive project from their parent entity (taskable)
        if entity_type == "Task":
            result = session.execute(
                text('SELECT taskable_type, taskable_id FROM "Task" WHERE id = :id'),
                {"id": entity_id}
            ).fetchone()
            if result and result[0] and result[1]:
                # Recursively get project from parent entity
                return _get_entity_project_id(result[0], result[1])
            return None

        # Junction table relationships
        if entity_type == "Lead":
            result = session.execute(
                text('SELECT project_id FROM "LeadProject" WHERE lead_id = :id LIMIT 1'),
                {"id": entity_id}
            ).fetchone()
            return result[0] if result else None

        if entity_type in ("Individual", "Organization"):
            # Get account_id first, then lookup project
            table = entity_type
            result = session.execute(
                text(f'SELECT account_id FROM "{table}" WHERE id = :id'),
                {"id": entity_id}
            ).fetchone()
            if result and result[0]:
                account_id = result[0]
                proj_result = session.execute(
                    text('SELECT project_id FROM "AccountProject" WHERE account_id = :id LIMIT 1'),
                    {"id": account_id}
                ).fetchone()
                return proj_result[0] if proj_result else None

        return None
    finally:
        session.close()


def _get_lead_url(entity_id: int) -> str:
    """Get URL for a Lead entity by looking up its type."""
    session = get_session()
    try:
        result = session.execute(
            text('SELECT type FROM "Lead" WHERE id = :id'),
            {"id": entity_id}
        ).fetchone()
        if not result or not result[0]:
            return f"/leads/{entity_id}"
        lead_type = result[0].lower() + "s"  # 'Job' -> 'jobs', 'Opportunity' -> 'opportunitys'
        # Fix pluralization
        if lead_type == "opportunitys":
            lead_type = "opportunities"
        return f"/leads/{lead_type}/{entity_id}"
    finally:
        session.close()


def _get_entity_url(entity_type: str, entity_id: int, comment_id: int = None) -> str:
    """Get URL for an entity, optionally with comment anchor."""
    if entity_type == "Lead":
        url = _get_lead_url(entity_id)
    else:
        base_path = ENTITY_URL_MAP.get(entity_type, f"/{entity_type.lower()}s")
        url = f"{base_path}/{entity_id}"

    if not comment_id:
        return url
    return f"{url}#comment-{comment_id}"


def _get_entity_display_name(comment) -> str:
    """Get display name for the entity a comment belongs to."""
    # For now, return a generic name based on type and ID
    # In future, could join to entity tables to get actual names
    return f"{comment.commentable_type} #{comment.commentable_id}"


def _notification_to_dict(notification) -> dict:
    """Convert notification model to dictionary with nested comment/entity info."""
    comment = notification.comment
    creator = comment.creator if comment else None

    created_by_name = None
    if creator:
        first = creator.first_name or ""
        last = creator.last_name or ""
        created_by_name = f"{first} {last}".strip() or creator.email

    # Truncate content for preview
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
            "url": _get_entity_url(comment.commentable_type, comment.commentable_id, comment.id) if comment else None,
            "project_id": _get_entity_project_id(comment.commentable_type, comment.commentable_id) if comment else None,
        } if comment else None,
    }


class MarkReadRequest(BaseModel):
    notification_ids: List[int]


class MarkEntityReadRequest(BaseModel):
    entity_type: str
    entity_id: int


router = APIRouter(prefix="/notifications", tags=["notifications"])


@router.get("/count")
@require_auth
@log_errors
async def get_unread_count_route(request: Request):
    """Get unread notification count for the current user."""
    user_id = request.state.user_id
    tenant_id = request.state.tenant_id

    count = await repo_get_unread_count(user_id, tenant_id)
    return {"unread_count": count}


@router.get("")
@require_auth
@log_errors
async def get_notifications_route(
    request: Request,
    limit: int = Query(default=20, le=100),
):
    """Get notifications for the current user."""
    user_id = request.state.user_id
    tenant_id = request.state.tenant_id

    notifications = await repo_get_notifications(user_id, tenant_id, limit)
    unread_count = await repo_get_unread_count(user_id, tenant_id)

    return {
        "notifications": [_notification_to_dict(n) for n in notifications],
        "unread_count": unread_count,
    }


@router.post("/mark-read")
@require_auth
@log_errors
async def mark_read_route(request: Request, data: MarkReadRequest):
    """Mark specific notifications as read."""
    user_id = request.state.user_id
    tenant_id = request.state.tenant_id

    updated = await repo_mark_as_read(data.notification_ids, user_id, tenant_id)
    return {"status": "ok", "updated": updated}


@router.post("/mark-entity-read")
@require_auth
@log_errors
async def mark_entity_read_route(request: Request, data: MarkEntityReadRequest):
    """Mark all notifications for an entity as read."""
    user_id = request.state.user_id
    tenant_id = request.state.tenant_id

    updated = await repo_mark_entity_as_read(
        user_id, tenant_id, data.entity_type, data.entity_id
    )
    return {"status": "ok", "updated": updated}
