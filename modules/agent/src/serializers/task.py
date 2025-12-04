"""Serializers for task-related models."""

from typing import Optional, List


def format_datetime(value) -> Optional[str]:
    """Format datetime to ISO string."""
    if not value:
        return None
    return value.isoformat() if hasattr(value, "isoformat") else str(value)


def format_date(value) -> Optional[str]:
    """Format date to ISO string."""
    if not value:
        return None
    date_str = value.isoformat() if hasattr(value, "isoformat") else str(value)
    return date_str[:10] if date_str else None


def format_decimal(value) -> Optional[float]:
    """Format Decimal to float."""
    if value is None:
        return None
    return float(value)


def to_paged_response_with_scope(count: int, page: int, limit: int, items: List, scope: Optional[dict] = None) -> dict:
    """Convert to paged response format with optional scope."""
    response = {
        "items": items,
        "currentPage": page,
        "totalPages": max(1, (count + limit - 1) // limit),
        "total": count,
        "pageSize": limit,
    }
    if scope:
        response["scope"] = scope
    return response


def task_to_list_response(task, entity_info: Optional[dict] = None) -> dict:
    """Convert Task model to dictionary - optimized for list/table view."""
    if entity_info is None:
        entity_info = {
            "type": None,
            "id": None,
            "display_name": None,
            "url": None,
            "lead_type": None,
        }

    return {
        "id": task.id,
        "title": task.title,
        "due_date": format_date(task.due_date),
        "status": {
            "id": task.current_status.id,
            "name": task.current_status.name,
        } if task.current_status else None,
        "priority": {
            "id": task.priority.id,
            "name": task.priority.name,
        } if task.priority else None,
        "entity": entity_info,
        "created_at": format_datetime(task.created_at),
        "updated_at": format_datetime(task.updated_at),
    }


def task_to_detail_response(task, entity_info: Optional[dict] = None) -> dict:
    """Convert Task model to dictionary - full detail view."""
    if entity_info is None:
        entity_info = {
            "type": None,
            "id": None,
            "display_name": None,
            "url": None,
            "lead_type": None,
        }

    return {
        "id": task.id,
        "title": task.title,
        "description": task.description,
        "start_date": format_date(task.start_date),
        "due_date": format_date(task.due_date),
        "estimated_hours": format_decimal(task.estimated_hours),
        "actual_hours": format_decimal(task.actual_hours),
        "complexity": task.complexity,
        "sort_order": task.sort_order,
        "completed_at": format_datetime(task.completed_at),
        "assigned_to_individual_id": task.assigned_to_individual_id,
        "status": {
            "id": task.current_status.id,
            "name": task.current_status.name,
        } if task.current_status else None,
        "priority": {
            "id": task.priority.id,
            "name": task.priority.name,
        } if task.priority else None,
        "entity": entity_info,
        "created_at": format_datetime(task.created_at),
        "updated_at": format_datetime(task.updated_at),
    }
