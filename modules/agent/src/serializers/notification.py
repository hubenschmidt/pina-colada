"""Serializers for notification-related models."""

from typing import Optional

from repositories.notification_repository import (
    get_entity_project_id_sync,
    get_lead_type_sync,
)

ENTITY_URL_MAP = {
    "Lead": "/leads",
    "Deal": "/deals",
    "Task": "/tasks",
    "Individual": "/accounts/individuals",
    "Organization": "/accounts/organizations",
}


def _get_entity_url(entity_type: str, entity_id: int, comment_id: Optional[int] = None) -> str:
    """Get URL for an entity, optionally with comment anchor."""
    if entity_type == "Lead":
        lead_type = get_lead_type_sync(entity_id)
        url = f"/leads/{lead_type}/{entity_id}"
    else:
        base_path = ENTITY_URL_MAP.get(entity_type, f"/{entity_type.lower()}s")
        url = f"{base_path}/{entity_id}"

    if not comment_id:
        return url
    return f"{url}#comment-{comment_id}"


def _get_entity_display_name(comment) -> str:
    """Get display name for the entity a comment belongs to."""
    return f"{comment.commentable_type} #{comment.commentable_id}"


def notification_to_response(notification) -> dict:
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
            "url": _get_entity_url(comment.commentable_type, comment.commentable_id, comment.id) if comment else None,
            "project_id": get_entity_project_id_sync(comment.commentable_type, comment.commentable_id) if comment else None,
        } if comment else None,
    }
