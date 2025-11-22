"""Controller layer for node config routing to repositories."""

import logging
from typing import List, Optional, Dict, Any
from lib.serialization import model_to_dict
from lib.decorators import handle_http_exceptions
from repositories.node_config_repository import (
    find_all_node_configs,
    find_node_config_by_id,
    create_node_config as create_node_config_repo,
    update_node_config as update_node_config_repo,
    delete_node_config as delete_node_config_repo,
)

logger = logging.getLogger(__name__)


def _config_to_response_dict(config) -> Dict[str, Any]:
    """Convert node config ORM to response dictionary."""
    config_dict = model_to_dict(config, include_relationships=False)
    return {
        "id": config_dict.get("id"),
        "node_type": config_dict.get("node_type"),
        "node_name": config_dict.get("node_name"),
        "system_prompt": config_dict.get("system_prompt"),
        "is_active": config_dict.get("is_active"),
        "created_at": str(config_dict.get("created_at", "")),
        "updated_at": str(config_dict.get("updated_at", "")),
        "updated_by": config_dict.get("updated_by"),
    }


@handle_http_exceptions
async def get_node_configs(node_type: Optional[str] = None) -> List[Dict[str, Any]]:
    """Get all node configs, optionally filtered by node_type."""
    configs = await find_all_node_configs(node_type)
    return [_config_to_response_dict(c) for c in configs]


@handle_http_exceptions
async def get_node_config(config_id: int) -> Dict[str, Any]:
    """Get a node config by ID."""
    config = await find_node_config_by_id(config_id)
    if not config:
        from fastapi import HTTPException
        raise HTTPException(status_code=404, detail="Node config not found")
    return _config_to_response_dict(config)


@handle_http_exceptions
async def create_node_config(data: Dict[str, Any]) -> Dict[str, Any]:
    """Create a new node config."""
    config = await create_node_config_repo(data)
    return _config_to_response_dict(config)


@handle_http_exceptions
async def update_node_config(config_id: int, data: Dict[str, Any]) -> Dict[str, Any]:
    """Update a node config."""
    config = await update_node_config_repo(config_id, data)
    if not config:
        from fastapi import HTTPException
        raise HTTPException(status_code=404, detail="Node config not found")
    return _config_to_response_dict(config)


@handle_http_exceptions
async def delete_node_config(config_id: int, deleted_by: str = "system") -> Dict[str, bool]:
    """Delete a node config (soft delete)."""
    success = await delete_node_config_repo(config_id, deleted_by)
    if not success:
        from fastapi import HTTPException
        raise HTTPException(status_code=404, detail="Node config not found")
    return {"success": True}
