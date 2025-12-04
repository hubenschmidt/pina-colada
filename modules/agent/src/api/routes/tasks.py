"""Routes for tasks API endpoints."""

from typing import Optional

from fastapi import APIRouter, Request, Query

from controllers.task_controller import (
    create_task,
    delete_task,
    get_task,
    get_task_priorities,
    get_task_statuses,
    get_tasks,
    get_tasks_by_entity,
    update_task,
)
from schemas.task import TaskCreate, TaskUpdate
from lib.auth import require_auth
from lib.error_logging import log_errors


router = APIRouter(prefix="/tasks", tags=["tasks"])


@router.get("/statuses")
@require_auth
@log_errors
async def get_task_statuses_route(request: Request):
    """Get all task statuses."""
    return await get_task_statuses()


@router.get("/priorities")
@require_auth
@log_errors
async def get_task_priorities_route(request: Request):
    """Get all task priorities."""
    return await get_task_priorities()


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
    search: Optional[str] = Query(None),
):
    """Get tasks with optional project scope filtering."""
    return await get_tasks(request, page, limit, order_by, order, scope, project_id, search)


@router.get("/entity/{entity_type}/{entity_id}")
@require_auth
@log_errors
async def get_tasks_by_entity_route(request: Request, entity_type: str, entity_id: int):
    """Get all tasks for a specific entity."""
    return await get_tasks_by_entity(entity_type, entity_id)


@router.get("/{task_id}")
@require_auth
@log_errors
async def get_task_route(request: Request, task_id: str):
    """Get a single task by ID."""
    return await get_task(task_id)


@router.post("")
@require_auth
@log_errors
async def create_task_route(request: Request, data: TaskCreate):
    """Create a new task."""
    return await create_task(request, data)


@router.put("/{task_id}")
@require_auth
@log_errors
async def update_task_route(request: Request, task_id: str, data: TaskUpdate):
    """Update a task."""
    return await update_task(request, task_id, data)


@router.delete("/{task_id}")
@require_auth
@log_errors
async def delete_task_route(request: Request, task_id: str):
    """Delete a task."""
    return await delete_task(task_id)
