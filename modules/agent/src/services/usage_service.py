"""Service layer for usage analytics business logic."""

from typing import Dict, List, Optional

from repositories import usage_analytics_repository as repo
from services.pricing_service import calculate_cost, get_cost_tier


def _add_cost_to_record(record: Dict) -> Dict:
    """Add cost calculation to a usage record."""
    model_name = record.get("model_name", "")
    cost = calculate_cost(
        model_name,
        record.get("input_tokens", 0),
        record.get("output_tokens", 0),
    )
    record["estimated_cost"] = round(cost, 6) if cost is not None else None
    record["cost_tier"] = get_cost_tier(cost)
    return record


def _add_cost_to_aggregate(record: Dict, model_name: Optional[str] = None) -> Dict:
    """Add cost calculation to an aggregate record (uses average pricing if no model)."""
    cost = calculate_cost(
        model_name or "",
        record.get("input_tokens", 0),
        record.get("output_tokens", 0),
    )
    record["estimated_cost"] = round(cost, 6) if cost is not None else None
    record["cost_tier"] = get_cost_tier(cost)
    return record


async def get_user_usage(user_id: int, period: str = "monthly") -> Dict:
    """Get aggregated usage for a user with cost estimate."""
    usage = await repo.get_user_usage(user_id, period)
    return _add_cost_to_aggregate(usage)


async def get_tenant_usage(tenant_id: int, period: str = "monthly") -> Dict:
    """Get aggregated usage for a tenant with cost estimate."""
    usage = await repo.get_tenant_usage(tenant_id, period)
    return _add_cost_to_aggregate(usage)


async def get_usage_timeseries(
    tenant_id: int,
    period: str = "monthly",
    user_id: Optional[int] = None,
) -> List[Dict]:
    """Get usage data grouped by date for charting."""
    data = await repo.get_usage_timeseries(tenant_id, period, user_id)
    return [_add_cost_to_aggregate(record) for record in data]


async def get_usage_by_node(tenant_id: int, period: str = "monthly") -> List[Dict]:
    """Get usage breakdown by node (developer analytics)."""
    data = await repo.get_usage_by_node(tenant_id, period)
    return [_add_cost_to_aggregate(record) for record in data]


async def get_usage_by_tool(tenant_id: int, period: str = "monthly") -> List[Dict]:
    """Get usage breakdown by tool (developer analytics)."""
    data = await repo.get_usage_by_tool(tenant_id, period)
    return [_add_cost_to_aggregate(record) for record in data]


async def get_usage_by_model(tenant_id: int, period: str = "monthly") -> List[Dict]:
    """Get usage breakdown by model (developer analytics)."""
    data = await repo.get_usage_by_model(tenant_id, period)
    return [_add_cost_to_record(record) for record in data]


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
