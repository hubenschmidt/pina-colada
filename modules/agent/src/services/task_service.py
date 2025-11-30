"""Service layer for task business logic."""

import logging
from typing import Dict, List, Optional, Any, Tuple
from fastapi import HTTPException
from repositories.task_repository import (
    find_all_tasks,
    find_task_by_id,
    find_tasks_by_entity,
    create_task as create_task_repo,
    update_task as update_task_repo,
    delete_task as delete_task_repo,
    get_entity_display_name,
)

logger = logging.getLogger(__name__)


async def get_tasks_paginated(
    tenant_id: int,
    project_id: Optional[int] = None,
    page: int = 1,
    page_size: int = 20,
    order_by: str = "created_at",
    order: str = "DESC",
) -> Tuple[List[Any], int]:
    """Get tasks with pagination and optional project scope."""
    return await find_all_tasks(
        tenant_id=tenant_id,
        project_id=project_id,
        page=page,
        page_size=page_size,
        order_by=order_by,
        order=order,
    )


async def get_task(task_id: str) -> Any:
    """Get a task by ID."""
    try:
        task_id_int = int(task_id)
    except ValueError:
        raise HTTPException(status_code=400, detail="Invalid task ID format")

    task = await find_task_by_id(task_id_int)
    if not task:
        raise HTTPException(status_code=404, detail="Task not found")

    return task


async def create_task(task_data: Dict[str, Any]) -> Any:
    """Create a new task."""
    return await create_task_repo(task_data)


async def update_task(task_id: str, task_data: Dict[str, Any]) -> Any:
    """Update a task."""
    try:
        task_id_int = int(task_id)
    except ValueError:
        raise HTTPException(status_code=400, detail="Invalid task ID format")

    task = await update_task_repo(task_id_int, task_data)
    if not task:
        raise HTTPException(status_code=404, detail="Task not found")

    return task


async def delete_task(task_id: str) -> bool:
    """Delete a task."""
    try:
        task_id_int = int(task_id)
    except ValueError:
        raise HTTPException(status_code=400, detail="Invalid task ID format")

    deleted = await delete_task_repo(task_id_int)
    if not deleted:
        raise HTTPException(status_code=404, detail="Task not found")

    return deleted


async def get_tasks_by_entity(entity_type: str, entity_id: int) -> List[Any]:
    """Get all tasks for a specific entity."""
    return await find_tasks_by_entity(entity_type, entity_id)


async def resolve_entity_display(taskable_type: Optional[str], taskable_id: Optional[int]) -> Dict[str, Any]:
    """Resolve entity display info for a task."""
    if not taskable_type or not taskable_id:
        return {
            "type": None,
            "id": None,
            "display_name": None,
            "url": None,
        }

    display_name = await get_entity_display_name(taskable_type, taskable_id)

    # Build URL based on entity type
    url_map = {
        "Project": f"/projects/{taskable_id}",
        "Deal": f"/deals/{taskable_id}",
        "Lead": f"/leads/{taskable_id}",
        "Account": f"/accounts/{taskable_id}",
    }

    return {
        "type": taskable_type,
        "id": taskable_id,
        "display_name": display_name,
        "url": url_map.get(taskable_type),
    }
