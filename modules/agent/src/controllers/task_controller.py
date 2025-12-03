"""Controller layer for task routing to services."""

import logging
from typing import Optional, List, Dict, Any

from fastapi import Request

from lib.decorators import handle_http_exceptions
from services.task_service import (
    TaskCreate,
    TaskUpdate,
    get_tasks_paginated,
    get_task as get_task_service,
    get_tasks_by_entity as get_tasks_by_entity_service,
    create_task as create_task_service,
    update_task as update_task_service,
    delete_task as delete_task_service,
    batch_resolve_entity_display,
    get_task_statuses as get_statuses_service,
    get_task_priorities as get_priorities_service,
)

logger = logging.getLogger(__name__)

# Re-export for routes
__all__ = ["TaskCreate", "TaskUpdate"]


def _format_datetime(value) -> Optional[str]:
    """Format datetime to ISO string."""
    if not value:
        return None
    return value.isoformat() if hasattr(value, "isoformat") else str(value)


def _format_date(value) -> Optional[str]:
    """Format date to ISO string."""
    if not value:
        return None
    date_str = value.isoformat() if hasattr(value, "isoformat") else str(value)
    return date_str[:10] if date_str else None


def _format_decimal(value) -> Optional[float]:
    """Format Decimal to float."""
    if value is None:
        return None
    return float(value)


def _to_paged_response(count: int, page: int, limit: int, items: List, scope: Optional[dict] = None) -> dict:
    """Convert to paged response format."""
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


def _task_to_list_response(task, entity_info: Optional[dict] = None) -> dict:
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
        "due_date": _format_date(task.due_date),
        "status": {
            "id": task.current_status.id,
            "name": task.current_status.name,
        } if task.current_status else None,
        "priority": {
            "id": task.priority.id,
            "name": task.priority.name,
        } if task.priority else None,
        "entity": entity_info,
        "created_at": _format_datetime(task.created_at),
        "updated_at": _format_datetime(task.updated_at),
    }


def _task_to_detail_response(task, entity_info: Optional[dict] = None) -> dict:
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
        "start_date": _format_date(task.start_date),
        "due_date": _format_date(task.due_date),
        "estimated_hours": _format_decimal(task.estimated_hours),
        "actual_hours": _format_decimal(task.actual_hours),
        "complexity": task.complexity,
        "sort_order": task.sort_order,
        "completed_at": _format_datetime(task.completed_at),
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
        "created_at": _format_datetime(task.created_at),
        "updated_at": _format_datetime(task.updated_at),
    }


@handle_http_exceptions
async def get_task_statuses() -> List[dict]:
    """Get all task statuses."""
    statuses = await get_statuses_service()
    return [{"id": s.id, "name": s.name} for s in statuses]


@handle_http_exceptions
async def get_task_priorities() -> List[dict]:
    """Get all task priorities."""
    priorities = await get_priorities_service()
    return [{"id": p.id, "name": p.name} for p in priorities]


@handle_http_exceptions
async def get_tasks(
    request: Request,
    page: int,
    limit: int,
    order_by: str,
    order: str,
    scope: str = "global",
    project_id: Optional[int] = None,
    search: Optional[str] = None,
) -> dict:
    """Get tasks with pagination."""
    tenant_id = request.state.tenant_id
    effective_project_id = project_id if scope == "project" and project_id else None

    tasks, total = await get_tasks_paginated(
        tenant_id=tenant_id,
        project_id=effective_project_id,
        page=page,
        page_size=limit,
        order_by=order_by,
        order=order,
        search=search,
    )

    entity_info_map = await batch_resolve_entity_display(tasks)

    items = []
    for task in tasks:
        key = (task.taskable_type, task.taskable_id) if task.taskable_type and task.taskable_id else None
        entity_info = entity_info_map.get(key) if key else None
        items.append(_task_to_list_response(task, entity_info))

    scope_info = {"type": scope}
    if effective_project_id:
        scope_info["project_id"] = effective_project_id

    return _to_paged_response(total, page, limit, items, scope_info)


@handle_http_exceptions
async def get_tasks_by_entity(entity_type: str, entity_id: int) -> dict:
    """Get all tasks for a specific entity."""
    tasks = await get_tasks_by_entity_service(entity_type, entity_id)
    entity_info_map = await batch_resolve_entity_display(tasks)

    items = []
    for task in tasks:
        key = (task.taskable_type, task.taskable_id) if task.taskable_type and task.taskable_id else None
        entity_info = entity_info_map.get(key) if key else None
        items.append(_task_to_list_response(task, entity_info))

    return {"items": items}


@handle_http_exceptions
async def get_task(task_id: str) -> dict:
    """Get a single task by ID."""
    task = await get_task_service(task_id)
    return _task_to_detail_response(task)


@handle_http_exceptions
async def create_task(request: Request, data: TaskCreate) -> dict:
    """Create a new task."""
    tenant_id = request.state.tenant_id
    user_id = request.state.user_id
    task_data = {
        "tenant_id": tenant_id,
        "title": data.title,
        "description": data.description,
        "taskable_type": data.taskable_type,
        "taskable_id": data.taskable_id,
        "current_status_id": data.current_status_id,
        "priority_id": data.priority_id,
        "start_date": data.start_date,
        "due_date": data.due_date,
        "estimated_hours": data.estimated_hours,
        "actual_hours": data.actual_hours,
        "complexity": data.complexity,
        "sort_order": data.sort_order,
        "assigned_to_individual_id": data.assigned_to_individual_id,
        "created_by": user_id,
        "updated_by": user_id,
    }
    task = await create_task_service(task_data)
    return _task_to_detail_response(task)


@handle_http_exceptions
async def update_task(request: Request, task_id: str, data: TaskUpdate) -> dict:
    """Update a task."""
    user_id = request.state.user_id
    update_data = {k: v for k, v in data.model_dump().items() if v is not None}
    update_data["updated_by"] = user_id
    task = await update_task_service(task_id, update_data)
    return _task_to_detail_response(task)


@handle_http_exceptions
async def delete_task(task_id: str) -> dict:
    """Delete a task."""
    await delete_task_service(task_id)
    return {"status": "deleted"}
