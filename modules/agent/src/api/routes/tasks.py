"""Routes for tasks API endpoints."""

from typing import Optional, List, Literal
from datetime import date
from decimal import Decimal
from fastapi import APIRouter, Request, HTTPException, Query
from pydantic import BaseModel, field_validator

from lib.auth import require_auth
from lib.error_logging import log_errors
from services.task_service import (
    get_tasks_paginated,
    get_task as get_task_service,
    get_tasks_by_entity,
    create_task as create_task_service,
    update_task as update_task_service,
    delete_task as delete_task_service,
    resolve_entity_display,
    batch_resolve_entity_display,
    get_task_statuses,
    get_task_priorities,
)


FIBONACCI_VALUES = (1, 2, 3, 5, 8, 13, 21)


class TaskCreate(BaseModel):
    title: str
    description: Optional[str] = None
    taskable_type: Optional[str] = None
    taskable_id: Optional[int] = None
    current_status_id: Optional[int] = None
    priority_id: Optional[int] = None
    start_date: Optional[date] = None
    due_date: Optional[date] = None
    estimated_hours: Optional[Decimal] = None
    actual_hours: Optional[Decimal] = None
    complexity: Optional[int] = None
    sort_order: Optional[int] = None
    assigned_to_individual_id: Optional[int] = None

    @field_validator("complexity")
    @classmethod
    def validate_complexity(cls, v):
        if v is not None and v not in FIBONACCI_VALUES:
            raise ValueError(f"complexity must be one of {FIBONACCI_VALUES}")
        return v


class TaskUpdate(BaseModel):
    title: Optional[str] = None
    description: Optional[str] = None
    taskable_type: Optional[str] = None
    taskable_id: Optional[int] = None
    current_status_id: Optional[int] = None
    priority_id: Optional[int] = None
    start_date: Optional[date] = None
    due_date: Optional[date] = None
    estimated_hours: Optional[Decimal] = None
    actual_hours: Optional[Decimal] = None
    complexity: Optional[int] = None
    sort_order: Optional[int] = None
    completed_at: Optional[str] = None
    assigned_to_individual_id: Optional[int] = None

    @field_validator("complexity")
    @classmethod
    def validate_complexity(cls, v):
        if v is not None and v not in FIBONACCI_VALUES:
            raise ValueError(f"complexity must be one of {FIBONACCI_VALUES}")
        return v


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


async def _task_to_list_dict(task, entity_info: Optional[dict] = None) -> dict:
    """Convert Task model to dictionary - optimized for list/table view.
    
    Only returns fields needed for table columns:
    Task, Linked To, Status, Priority, Due Date, Created, Updated
    """
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


async def _task_to_dict(task, entity_info: Optional[dict] = None) -> dict:
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


router = APIRouter(prefix="/tasks", tags=["tasks"])


@router.get("/statuses")
@require_auth
@log_errors
async def get_task_statuses_route(request: Request):
    """Get all task statuses."""
    statuses = await get_task_statuses()
    return [{"id": s.id, "name": s.name} for s in statuses]


@router.get("/priorities")
@require_auth
@log_errors
async def get_task_priorities_route(request: Request):
    """Get all task priorities."""
    priorities = await get_task_priorities()
    return [{"id": p.id, "name": p.name} for p in priorities]


@router.get("")
@require_auth
@log_errors
async def get_tasks_route(
    request: Request,
    project_id: Optional[int] = Query(None, alias="projectId"),
    scope: str = Query("global"),
    page: int = Query(1, ge=1),
    limit: int = Query(20, ge=1, le=100),
    order_by: str = Query("created_at", alias="orderBy"),
    order: str = Query("DESC"),
):
    """Get all tasks with optional project scope filtering.

    - scope="project" + project_id: Tasks within project's entity graph
    - scope="global": All tasks in tenant
    """
    tenant_id = request.state.tenant_id

    # Determine project_id based on scope
    effective_project_id = None
    if scope == "project" and project_id:
        effective_project_id = project_id

    tasks, total = await get_tasks_paginated(
        tenant_id=tenant_id,
        project_id=effective_project_id,
        page=page,
        page_size=limit,
        order_by=order_by,
        order=order,
    )

    # Batch resolve entity display info for all tasks
    entity_info_map = await batch_resolve_entity_display(tasks)

    # Convert tasks to response format (slimmed down for list view)
    items = []
    for task in tasks:
        key = (task.taskable_type, task.taskable_id) if task.taskable_type and task.taskable_id else None
        entity_info = entity_info_map.get(key) if key else None
        items.append(await _task_to_list_dict(task, entity_info))

    # Build scope info
    scope_info = {"type": scope}
    if effective_project_id:
        scope_info["project_id"] = effective_project_id

    return _to_paged_response(total, page, limit, items, scope_info)


@router.get("/entity/{entity_type}/{entity_id}")
@require_auth
@log_errors
async def get_tasks_by_entity_route(
    request: Request,
    entity_type: str,
    entity_id: int,
):
    """Get all tasks for a specific entity."""
    tasks = await get_tasks_by_entity(entity_type, entity_id)
    entity_info_map = await batch_resolve_entity_display(tasks)
    items = []
    for task in tasks:
        key = (task.taskable_type, task.taskable_id) if task.taskable_type and task.taskable_id else None
        entity_info = entity_info_map.get(key) if key else None
        items.append(await _task_to_list_dict(task, entity_info))
    return {"items": items}


@router.get("/{task_id}")
@require_auth
@log_errors
async def get_task_route(request: Request, task_id: str):
    """Get a single task by ID."""
    task = await get_task_service(task_id)
    return await _task_to_dict(task)


@router.post("")
@require_auth
@log_errors
async def create_task_route(request: Request, data: TaskCreate):
    """Create a new task."""
    tenant_id = request.state.tenant_id
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
    }

    task = await create_task_service(task_data)
    return await _task_to_dict(task)


@router.put("/{task_id}")
@require_auth
@log_errors
async def update_task_route(request: Request, task_id: str, data: TaskUpdate):
    """Update a task."""
    update_data = {k: v for k, v in data.model_dump().items() if v is not None}

    task = await update_task_service(task_id, update_data)
    return await _task_to_dict(task)


@router.delete("/{task_id}")
@require_auth
@log_errors
async def delete_task_route(request: Request, task_id: str):
    """Delete a task."""
    await delete_task_service(task_id)
    return {"status": "deleted"}
