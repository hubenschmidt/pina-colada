"""Routes for node configs API endpoints."""

from typing import Optional, Type
from fastapi import APIRouter, Query, Request
from pydantic import BaseModel, create_model
from lib.auth import require_auth
from lib.error_logging import log_errors
from controllers.node_config_controller import (
    get_node_configs,
    get_node_config,
    create_node_config,
    update_node_config,
    delete_node_config,
)

router = APIRouter(prefix="/node-configs", tags=["node-configs"])


def _make_node_config_create_model() -> Type[BaseModel]:
    """Create NodeConfigCreate model functionally."""
    return create_model(
        "NodeConfigCreate",
        node_type=(str, ...),
        node_name=(str, ...),
        system_prompt=(str, ...),
    )


def _make_node_config_update_model() -> Type[BaseModel]:
    """Create NodeConfigUpdate model functionally."""
    return create_model(
        "NodeConfigUpdate",
        system_prompt=(Optional[str], None),
        is_active=(Optional[bool], None),
    )


NodeConfigCreate = _make_node_config_create_model()
NodeConfigUpdate = _make_node_config_update_model()


@router.get("")
@log_errors
@require_auth
async def get_node_configs_route(
    request: Request,
    node_type: Optional[str] = Query(None, alias="nodeType"),
):
    """Get all node configs, optionally filtered by node_type."""
    return await get_node_configs(node_type)


@router.get("/{config_id}")
@log_errors
@require_auth
async def get_node_config_route(request: Request, config_id: int):
    """Get a node config by ID."""
    return await get_node_config(config_id)


@router.post("")
@log_errors
@require_auth
async def create_node_config_route(request: Request, config_data: NodeConfigCreate):
    """Create a new node config."""
    data = config_data.dict()
    # Get user from auth context if available
    user = getattr(request.state, "user", None)
    data["updated_by"] = user.get("email", "system") if user else "system"
    return await create_node_config(data)


@router.put("/{config_id}")
@log_errors
@require_auth
async def update_node_config_route(request: Request, config_id: int, config_data: NodeConfigUpdate):
    """Update a node config."""
    data = config_data.dict(exclude_unset=True)
    # Get user from auth context if available
    user = getattr(request.state, "user", None)
    data["updated_by"] = user.get("email", "system") if user else "system"
    return await update_node_config(config_id, data)


@router.delete("/{config_id}")
@log_errors
@require_auth
async def delete_node_config_route(request: Request, config_id: int):
    """Delete a node config (soft delete)."""
    user = getattr(request.state, "user", None)
    deleted_by = user.get("email", "system") if user else "system"
    return await delete_node_config(config_id, deleted_by)
