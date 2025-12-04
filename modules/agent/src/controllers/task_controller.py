"""Controller layer for task routing to services."""

import logging
from typing import Optional, List, Dict, Any

from fastapi import Request

from lib.decorators import handle_http_exceptions
from serializers.task import (
    to_paged_response_with_scope,
    task_to_list_response,
    task_to_detail_response,
)
from schemas.task import TaskCreate, TaskUpdate
from services.task_service import (
    batch_resolve_entity_display,
    create_task as create_task_service,
    delete_task as delete_task_service,
    get_task as get_task_service,
    get_task_priorities as get_priorities_service,
    get_task_statuses as get_statuses_service,
    get_tasks_by_entity as get_tasks_by_entity_service,
    get_tasks_paginated,
    update_task as update_task_service,
)


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
        items.append(task_to_list_response(task, entity_info))

    scope_info = {"type": scope}
    if effective_project_id:
        scope_info["project_id"] = effective_project_id

    return to_paged_response_with_scope(total, page, limit, items, scope_info)


@handle_http_exceptions
async def get_tasks_by_entity(entity_type: str, entity_id: int) -> dict:
    """Get all tasks for a specific entity."""
    tasks = await get_tasks_by_entity_service(entity_type, entity_id)
    entity_info_map = await batch_resolve_entity_display(tasks)

    items = []
    for task in tasks:
        key = (task.taskable_type, task.taskable_id) if task.taskable_type and task.taskable_id else None
        entity_info = entity_info_map.get(key) if key else None
        items.append(task_to_list_response(task, entity_info))

    return {"items": items}


@handle_http_exceptions
async def get_task(task_id: str) -> dict:
    """Get a single task by ID."""
    task = await get_task_service(task_id)
    return task_to_detail_response(task)


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
    return task_to_detail_response(task)


@handle_http_exceptions
async def update_task(request: Request, task_id: str, data: TaskUpdate) -> dict:
    """Update a task."""
    user_id = request.state.user_id
    update_data = {k: v for k, v in data.model_dump().items() if v is not None}
    update_data["updated_by"] = user_id
    task = await update_task_service(task_id, update_data)
    return task_to_detail_response(task)


@handle_http_exceptions
async def delete_task(task_id: str) -> dict:
    """Delete a task."""
    await delete_task_service(task_id)
    return {"status": "deleted"}
