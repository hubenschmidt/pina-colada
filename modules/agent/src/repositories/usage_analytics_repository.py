"""Repository layer for usage analytics data access."""

from datetime import datetime, timedelta, timezone
from typing import Dict, List, Optional

from sqlalchemy import and_, func, select
from lib.db import async_get_session
from models.UsageAnalytics import UsageAnalytics


def _get_period_start(period: str) -> datetime:
    """Calculate start datetime for given period."""
    now = datetime.now(timezone.utc)
    period_map = {
        "daily": timedelta(days=1),
        "weekly": timedelta(weeks=1),
        "monthly": timedelta(days=30),
        "quarterly": timedelta(days=90),
        "annual": timedelta(days=365),
    }
    delta = period_map.get(period, timedelta(days=30))
    return now - delta


async def insert_usage(
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
) -> UsageAnalytics:
    """Insert a usage analytics record."""
    async with async_get_session() as session:
        usage = UsageAnalytics(
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
        session.add(usage)
        await session.commit()
        await session.refresh(usage)
        return usage


async def insert_usage_batch(records: List[Dict]) -> None:
    """Insert multiple usage records in a single transaction."""
    if not records:
        return
    async with async_get_session() as session:
        for record in records:
            usage = UsageAnalytics(
                tenant_id=record["tenant_id"],
                user_id=record["user_id"],
                input_tokens=record.get("input_tokens", 0),
                output_tokens=record.get("output_tokens", 0),
                total_tokens=record.get("total_tokens", 0),
                node_name=record.get("node_name"),
                tool_name=record.get("tool_name"),
                model_name=record.get("model_name"),
                conversation_id=record.get("conversation_id"),
                message_id=record.get("message_id"),
            )
            session.add(usage)
        await session.commit()


async def get_user_usage(user_id: int, period: str) -> List[Dict]:
    """Get token usage for a user grouped by model for cost calculation."""
    async with async_get_session() as session:
        period_start = _get_period_start(period)
        stmt = (
            select(
                UsageAnalytics.model_name,
                func.coalesce(func.sum(UsageAnalytics.input_tokens), 0).label("input_tokens"),
                func.coalesce(func.sum(UsageAnalytics.output_tokens), 0).label("output_tokens"),
                func.coalesce(func.sum(UsageAnalytics.total_tokens), 0).label("total_tokens"),
            )
            .where(
                and_(
                    UsageAnalytics.user_id == user_id,
                    UsageAnalytics.created_at >= period_start,
                )
            )
            .group_by(UsageAnalytics.model_name)
        )
        result = await session.execute(stmt)
        return [
            {
                "model_name": row.model_name,
                "input_tokens": row.input_tokens,
                "output_tokens": row.output_tokens,
                "total_tokens": row.total_tokens,
            }
            for row in result.all()
        ]


async def get_tenant_usage(tenant_id: int, period: str) -> List[Dict]:
    """Get token usage for a tenant grouped by model for cost calculation."""
    async with async_get_session() as session:
        period_start = _get_period_start(period)
        stmt = (
            select(
                UsageAnalytics.model_name,
                func.coalesce(func.sum(UsageAnalytics.input_tokens), 0).label("input_tokens"),
                func.coalesce(func.sum(UsageAnalytics.output_tokens), 0).label("output_tokens"),
                func.coalesce(func.sum(UsageAnalytics.total_tokens), 0).label("total_tokens"),
            )
            .where(
                and_(
                    UsageAnalytics.tenant_id == tenant_id,
                    UsageAnalytics.created_at >= period_start,
                )
            )
            .group_by(UsageAnalytics.model_name)
        )
        result = await session.execute(stmt)
        return [
            {
                "model_name": row.model_name,
                "input_tokens": row.input_tokens,
                "output_tokens": row.output_tokens,
                "total_tokens": row.total_tokens,
            }
            for row in result.all()
        ]


async def get_usage_timeseries(
    tenant_id: int,
    period: str,
    user_id: Optional[int] = None,
) -> List[Dict]:
    """Get usage data grouped by date for charting."""
    async with async_get_session() as session:
        period_start = _get_period_start(period)
        conditions = [
            UsageAnalytics.tenant_id == tenant_id,
            UsageAnalytics.created_at >= period_start,
        ]
        if user_id:
            conditions.append(UsageAnalytics.user_id == user_id)

        date_trunc_expr = func.date_trunc("day", UsageAnalytics.created_at)
        stmt = (
            select(
                date_trunc_expr.label("date"),
                func.sum(UsageAnalytics.input_tokens).label("input_tokens"),
                func.sum(UsageAnalytics.output_tokens).label("output_tokens"),
                func.sum(UsageAnalytics.total_tokens).label("total_tokens"),
            )
            .where(and_(*conditions))
            .group_by(date_trunc_expr)
            .order_by(date_trunc_expr)
        )
        result = await session.execute(stmt)
        return [
            {
                "date": row.date.isoformat() if row.date else None,
                "input_tokens": row.input_tokens or 0,
                "output_tokens": row.output_tokens or 0,
                "total_tokens": row.total_tokens or 0,
            }
            for row in result.all()
        ]


async def get_usage_by_node(tenant_id: int, period: str) -> List[Dict]:
    """Get usage breakdown by node name with model info for cost calculation."""
    async with async_get_session() as session:
        period_start = _get_period_start(period)
        # Group by node AND model to enable per-model cost calculation
        stmt = (
            select(
                UsageAnalytics.node_name,
                UsageAnalytics.model_name,
                func.count().label("request_count"),
                func.count(func.distinct(UsageAnalytics.conversation_id)).label(
                    "conversation_count"
                ),
                func.sum(UsageAnalytics.input_tokens).label("input_tokens"),
                func.sum(UsageAnalytics.output_tokens).label("output_tokens"),
                func.sum(UsageAnalytics.total_tokens).label("total_tokens"),
            )
            .where(
                and_(
                    UsageAnalytics.tenant_id == tenant_id,
                    UsageAnalytics.created_at >= period_start,
                    UsageAnalytics.node_name.isnot(None),
                )
            )
            .group_by(UsageAnalytics.node_name, UsageAnalytics.model_name)
            .order_by(func.sum(UsageAnalytics.total_tokens).desc())
        )
        result = await session.execute(stmt)
        return [
            {
                "node_name": row.node_name,
                "model_name": row.model_name,
                "request_count": row.request_count,
                "conversation_count": row.conversation_count,
                "input_tokens": row.input_tokens or 0,
                "output_tokens": row.output_tokens or 0,
                "total_tokens": row.total_tokens or 0,
            }
            for row in result.all()
        ]


async def get_usage_by_tool(tenant_id: int, period: str) -> List[Dict]:
    """Get usage breakdown by tool name (developer analytics)."""
    async with async_get_session() as session:
        period_start = _get_period_start(period)
        stmt = (
            select(
                UsageAnalytics.tool_name,
                func.sum(UsageAnalytics.input_tokens).label("input_tokens"),
                func.sum(UsageAnalytics.output_tokens).label("output_tokens"),
                func.sum(UsageAnalytics.total_tokens).label("total_tokens"),
            )
            .where(
                and_(
                    UsageAnalytics.tenant_id == tenant_id,
                    UsageAnalytics.created_at >= period_start,
                    UsageAnalytics.tool_name.isnot(None),
                )
            )
            .group_by(UsageAnalytics.tool_name)
            .order_by(func.sum(UsageAnalytics.total_tokens).desc())
        )
        result = await session.execute(stmt)
        return [
            {
                "tool_name": row.tool_name,
                "input_tokens": row.input_tokens or 0,
                "output_tokens": row.output_tokens or 0,
                "total_tokens": row.total_tokens or 0,
            }
            for row in result.all()
        ]


async def get_usage_by_model(tenant_id: int, period: str) -> List[Dict]:
    """Get usage breakdown by model name (developer analytics)."""
    async with async_get_session() as session:
        period_start = _get_period_start(period)
        stmt = (
            select(
                UsageAnalytics.model_name,
                func.count().label("request_count"),
                func.count(func.distinct(UsageAnalytics.conversation_id)).label(
                    "conversation_count"
                ),
                func.sum(UsageAnalytics.input_tokens).label("input_tokens"),
                func.sum(UsageAnalytics.output_tokens).label("output_tokens"),
                func.sum(UsageAnalytics.total_tokens).label("total_tokens"),
            )
            .where(
                and_(
                    UsageAnalytics.tenant_id == tenant_id,
                    UsageAnalytics.created_at >= period_start,
                    UsageAnalytics.model_name.isnot(None),
                )
            )
            .group_by(UsageAnalytics.model_name)
            .order_by(func.sum(UsageAnalytics.total_tokens).desc())
        )
        result = await session.execute(stmt)
        return [
            {
                "model_name": row.model_name,
                "request_count": row.request_count,
                "conversation_count": row.conversation_count,
                "input_tokens": row.input_tokens or 0,
                "output_tokens": row.output_tokens or 0,
                "total_tokens": row.total_tokens or 0,
            }
            for row in result.all()
        ]
