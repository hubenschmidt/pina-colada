"""Repository layer for node config data access."""

import logging
from typing import Dict, Any, Optional, List
from sqlalchemy import select
from models.NodeConfig import NodeConfig, NodeConfigHistory
from lib.db import async_get_session

logger = logging.getLogger(__name__)


async def find_all_node_configs(node_type: Optional[str] = None) -> List[NodeConfig]:
    """Find all node configs, optionally filtered by node_type."""
    async with async_get_session() as session:
        stmt = select(NodeConfig).where(NodeConfig.is_active == True)

        if node_type:
            stmt = stmt.where(NodeConfig.node_type == node_type)

        stmt = stmt.order_by(NodeConfig.node_type, NodeConfig.node_name)
        result = await session.execute(stmt)
        return list(result.scalars().all())


async def find_node_config_by_id(config_id: int) -> Optional[NodeConfig]:
    """Find node config by ID."""
    async with async_get_session() as session:
        stmt = select(NodeConfig).where(NodeConfig.id == config_id)
        result = await session.execute(stmt)
        return result.scalar_one_or_none()


async def find_node_config_by_name(node_type: str, node_name: str) -> Optional[NodeConfig]:
    """Find active node config by type and name."""
    async with async_get_session() as session:
        stmt = select(NodeConfig).where(
            NodeConfig.node_type == node_type,
            NodeConfig.node_name == node_name,
            NodeConfig.is_active == True
        )
        result = await session.execute(stmt)
        return result.scalar_one_or_none()


async def create_node_config(data: Dict[str, Any]) -> NodeConfig:
    """Create a new node config."""
    async with async_get_session() as session:
        try:
            config = NodeConfig(
                node_type=data["node_type"],
                node_name=data["node_name"],
                system_prompt=data["system_prompt"],
                updated_by=data.get("updated_by", "system")
            )
            session.add(config)
            await session.flush()

            # Create history entry
            history = NodeConfigHistory(
                config_id=config.id,
                node_type=config.node_type,
                node_name=config.node_name,
                system_prompt=config.system_prompt,
                changed_by=config.updated_by,
                change_type="create"
            )
            session.add(history)

            await session.commit()
            await session.refresh(config)
            return config
        except Exception as e:
            await session.rollback()
            logger.error(f"Failed to create node config: {e}")
            raise


async def update_node_config(config_id: int, data: Dict[str, Any]) -> Optional[NodeConfig]:
    """Update an existing node config."""
    async with async_get_session() as session:
        try:
            stmt = select(NodeConfig).where(NodeConfig.id == config_id)
            result = await session.execute(stmt)
            config = result.scalar_one_or_none()

            if not config:
                return None

            if "system_prompt" in data:
                config.system_prompt = data["system_prompt"]
            if "is_active" in data:
                config.is_active = data["is_active"]
            if "updated_by" in data:
                config.updated_by = data["updated_by"]

            # Create history entry
            history = NodeConfigHistory(
                config_id=config.id,
                node_type=config.node_type,
                node_name=config.node_name,
                system_prompt=config.system_prompt,
                changed_by=data.get("updated_by", "system"),
                change_type="update"
            )
            session.add(history)

            await session.commit()
            await session.refresh(config)
            return config
        except Exception as e:
            await session.rollback()
            logger.error(f"Failed to update node config: {e}")
            raise


async def delete_node_config(config_id: int, deleted_by: str = "system") -> bool:
    """Soft delete a node config by setting is_active to False."""
    async with async_get_session() as session:
        try:
            stmt = select(NodeConfig).where(NodeConfig.id == config_id)
            result = await session.execute(stmt)
            config = result.scalar_one_or_none()

            if not config:
                return False

            config.is_active = False
            config.updated_by = deleted_by

            # Create history entry
            history = NodeConfigHistory(
                config_id=config.id,
                node_type=config.node_type,
                node_name=config.node_name,
                system_prompt=config.system_prompt,
                changed_by=deleted_by,
                change_type="delete"
            )
            session.add(history)

            await session.commit()
            return True
        except Exception as e:
            await session.rollback()
            logger.error(f"Failed to delete node config: {e}")
            raise
