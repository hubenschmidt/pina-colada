"""Service for fetching costs from LLM provider APIs."""

import logging
import os
from datetime import datetime, timedelta, timezone
from typing import Dict, Optional

import httpx

logger = logging.getLogger(__name__)

OPENAI_COSTS_URL = "https://api.openai.com/v1/organization/costs"
ANTHROPIC_COSTS_URL = "https://api.anthropic.com/v1/organizations/cost_report"


def _get_period_timestamps(period: str) -> tuple[int, int]:
    """Get start and end timestamps for the given period."""
    now = datetime.now(timezone.utc)
    period_map = {
        "daily": timedelta(days=1),
        "weekly": timedelta(weeks=1),
        "monthly": timedelta(days=30),
        "quarterly": timedelta(days=90),
        "annual": timedelta(days=365),
    }
    delta = period_map.get(period, timedelta(days=30))
    start = now - delta
    return int(start.timestamp()), int(now.timestamp())


async def fetch_openai_costs(period: str = "monthly") -> Optional[Dict]:
    """Fetch costs from OpenAI Costs API."""
    admin_key = os.getenv("OPENAI_ADMIN_KEY")
    if not admin_key:
        logger.debug("OPENAI_ADMIN_KEY not configured")
        return None

    start_time, end_time = _get_period_timestamps(period)

    try:
        async with httpx.AsyncClient() as client:
            response = await client.get(
                OPENAI_COSTS_URL,
                params={
                    "start_time": start_time,
                    "end_time": end_time,
                    "bucket_width": "1d",
                },
                headers={
                    "Authorization": f"Bearer {admin_key}",
                    "Content-Type": "application/json",
                },
                timeout=30.0,
            )
            response.raise_for_status()
            data = response.json()

            # Sum up costs from all buckets
            total_spend = 0.0
            for bucket in data.get("data", []):
                for result in bucket.get("results", []):
                    total_spend += result.get("amount", {}).get("value", 0)

            return {
                "provider": "openai",
                "spend": round(total_spend, 4),
                "currency": "USD",
                "period": period,
                "fetched_at": datetime.now(timezone.utc).isoformat(),
            }
    except httpx.HTTPStatusError as e:
        logger.error(f"OpenAI Costs API error: {e.response.status_code} - {e.response.text}")
        return None
    except Exception as e:
        logger.error(f"Failed to fetch OpenAI costs: {e}")
        return None


async def fetch_anthropic_costs(period: str = "monthly") -> Optional[Dict]:
    """Fetch costs from Anthropic Cost Report API."""
    admin_key = os.getenv("ANTHROPIC_ADMIN_KEY")
    if not admin_key:
        logger.debug("ANTHROPIC_ADMIN_KEY not configured")
        return None

    start_time, end_time = _get_period_timestamps(period)
    start_date = datetime.fromtimestamp(start_time, tz=timezone.utc).strftime("%Y-%m-%d")
    end_date = datetime.fromtimestamp(end_time, tz=timezone.utc).strftime("%Y-%m-%d")

    try:
        async with httpx.AsyncClient() as client:
            response = await client.get(
                ANTHROPIC_COSTS_URL,
                params={
                    "start_date": start_date,
                    "end_date": end_date,
                },
                headers={
                    "x-api-key": admin_key,
                    "anthropic-version": "2023-06-01",
                    "Content-Type": "application/json",
                },
                timeout=30.0,
            )
            response.raise_for_status()
            data = response.json()

            # Sum up costs from response
            total_spend = 0.0
            for item in data.get("data", []):
                total_spend += item.get("cost_usd", 0)

            return {
                "provider": "anthropic",
                "spend": round(total_spend, 4),
                "currency": "USD",
                "period": period,
                "fetched_at": datetime.now(timezone.utc).isoformat(),
            }
    except httpx.HTTPStatusError as e:
        logger.error(f"Anthropic Costs API error: {e.response.status_code} - {e.response.text}")
        return None
    except Exception as e:
        logger.error(f"Failed to fetch Anthropic costs: {e}")
        return None


async def get_combined_costs(period: str = "monthly") -> Dict:
    """Fetch costs from all providers and combine."""
    openai_costs = await fetch_openai_costs(period)
    anthropic_costs = await fetch_anthropic_costs(period)

    result = {
        "period": period,
        "fetched_at": datetime.now(timezone.utc).isoformat(),
        "openai": openai_costs,
        "anthropic": anthropic_costs,
        "total_spend": 0.0,
    }

    if openai_costs:
        result["total_spend"] += openai_costs.get("spend", 0)
    if anthropic_costs:
        result["total_spend"] += anthropic_costs.get("spend", 0)

    result["total_spend"] = round(result["total_spend"], 4)

    return result
