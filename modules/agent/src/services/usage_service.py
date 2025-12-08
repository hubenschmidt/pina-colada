"""Service layer for usage analytics business logic."""

import logging
from typing import Dict, List, Optional

from repositories import usage_analytics_repository as repo

logger = logging.getLogger(__name__)


def _aggregate_usage(records: List[Dict]) -> Dict:
    """Aggregate usage records into totals."""
    totals = {
        "input_tokens": 0,
        "output_tokens": 0,
        "total_tokens": 0,
    }
    for record in records:
        totals["input_tokens"] += record.get("input_tokens", 0)
        totals["output_tokens"] += record.get("output_tokens", 0)
        totals["total_tokens"] += record.get("total_tokens", 0)
    return totals


async def get_user_usage(user_id: int, period: str = "monthly") -> Dict:
    """Get aggregated usage for a user."""
    records = await repo.get_user_usage(user_id, period)
    return _aggregate_usage(records)


async def get_tenant_usage(tenant_id: int, period: str = "monthly") -> Dict:
    """Get aggregated usage for a tenant."""
    records = await repo.get_tenant_usage(tenant_id, period)
    return _aggregate_usage(records)


async def get_usage_timeseries(
    tenant_id: int,
    period: str = "monthly",
    user_id: Optional[int] = None,
) -> List[Dict]:
    """Get usage data grouped by date for charting."""
    return await repo.get_usage_timeseries(tenant_id, period, user_id)


async def get_usage_by_node(tenant_id: int, period: str = "monthly") -> List[Dict]:
    """Get usage breakdown by node (developer analytics)."""
    raw_data = await repo.get_usage_by_node(tenant_id, period)

    # Aggregate by node (data comes grouped by node+model)
    node_totals = {}
    for record in raw_data:
        node = record["node_name"]
        if node not in node_totals:
            node_totals[node] = {
                "node_name": node,
                "request_count": 0,
                "conversation_count": 0,
                "input_tokens": 0,
                "output_tokens": 0,
                "total_tokens": 0,
            }
        node_totals[node]["request_count"] += record.get("request_count", 0)
        node_totals[node]["conversation_count"] += record.get("conversation_count", 0)
        node_totals[node]["input_tokens"] += record.get("input_tokens", 0)
        node_totals[node]["output_tokens"] += record.get("output_tokens", 0)
        node_totals[node]["total_tokens"] += record.get("total_tokens", 0)

    return sorted(node_totals.values(), key=lambda x: x["total_tokens"], reverse=True)


async def get_usage_by_model(tenant_id: int, period: str = "monthly") -> List[Dict]:
    """Get usage breakdown by model (developer analytics)."""
    return await repo.get_usage_by_model(tenant_id, period)


async def log_usage(
    tenant_id: int,
    user_id: int,
    input_tokens: int,
    output_tokens: int,
    total_tokens: int,
    node_name: Optional[str] = None,
    tool_name: Optional[str] = None,
    model_name: Optional[str] = None,
    conversation_id: Optional[int] = None,
    message_id: Optional[int] = None,
) -> None:
    """Log a usage record."""
    await repo.insert_usage(
        tenant_id=tenant_id,
        user_id=user_id,
        input_tokens=input_tokens,
        output_tokens=output_tokens,
        total_tokens=total_tokens,
        node_name=node_name,
        tool_name=tool_name,
        model_name=model_name,
        conversation_id=conversation_id,
        message_id=message_id,
    )


async def log_usage_records(
    tenant_id: int,
    user_id: int,
    records: List[Dict],
    conversation_id: Optional[int] = None,
) -> None:
    """Log usage records with model and node info."""
    logger.info(f"log_usage_records: tenant={tenant_id}, user={user_id}, count={len(records)}")
    db_records = [
        {
            "tenant_id": tenant_id,
            "user_id": user_id,
            "input_tokens": r.get("input", 0),
            "output_tokens": r.get("output", 0),
            "total_tokens": r.get("total", 0),
            "model_name": r.get("model_name"),
            "node_name": r.get("node_name"),
            "conversation_id": conversation_id,
        }
        for r in records
        if r.get("total", 0) > 0
    ]
    if db_records:
        await repo.insert_usage_batch(db_records)
        logger.info(f"Inserted {len(db_records)} usage records")
