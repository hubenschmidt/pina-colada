"""Service layer for user business logic."""

from typing import Optional

from fastapi import HTTPException

from repositories.user_repository import (
    find_user_by_email,
    find_tenant_by_id,
    find_individual_by_id,
    find_project_by_id,
    get_selected_project_id as get_selected_project_id_repo,
    set_selected_project as set_selected_project_repo,
)


async def get_user_tenant(email: str) -> dict:
    """Get tenant info for a user by email."""
    user = await find_user_by_email(email)
    if not user:
        raise HTTPException(status_code=404, detail="User not found")

    if not user.tenant_id:
        raise HTTPException(status_code=404, detail="No tenant associated with user")

    tenant = await find_tenant_by_id(user.tenant_id)
    if not tenant:
        raise HTTPException(status_code=404, detail="Tenant not found")

    individual = None
    if user.individual_id:
        individual = await find_individual_by_id(user.individual_id)

    return {
        "tenant": tenant,
        "individual": individual,
    }


async def set_selected_project(user_id: int, tenant_id: int, project_id: Optional[int]) -> Optional[int]:
    """Set user's selected project with tenant validation."""
    if project_id is not None:
        project = await find_project_by_id(project_id)
        if not project:
            raise HTTPException(status_code=404, detail="Project not found")
        if project.tenant_id != tenant_id:
            raise HTTPException(status_code=403, detail="Project does not belong to tenant")

    user = await set_selected_project_repo(user_id, project_id)
    if not user:
        raise HTTPException(status_code=404, detail="User not found")
    return user.selected_project_id


async def get_selected_project_id(user_id: int) -> Optional[int]:
    """Get user's selected project ID."""
    return await get_selected_project_id_repo(user_id)
