"""Service for fetching costs from LLM provider APIs."""

import asyncio
import logging
import os
from datetime import datetime, timedelta, timezone
from typing import Dict, List, Optional, Tuple

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


def _get_monthly_chunks(start_time: int, end_time: int) -> List[tuple[int, int]]:
    """Split a time range into monthly chunks for pagination-safe API calls."""
    chunks = []
    current_start = start_time
    while current_start < end_time:
        # 30 days in seconds
        current_end = min(current_start + (30 * 24 * 60 * 60), end_time)
        chunks.append((current_start, current_end))
        current_start = current_end
    return chunks


async def _fetch_openai_chunk(
    client: httpx.AsyncClient,
    admin_key: str,
    start_time: int,
    end_time: int,
) -> float:
    """Fetch costs for a single time chunk (no pagination needed for <=30 days)."""
    params = {
        "start_time": start_time,
        "end_time": end_time,
        "bucket_width": "1d",
        "limit": 180,
    }
    response = await client.get(
        OPENAI_COSTS_URL,
        params=params,
        headers={
            "Authorization": f"Bearer {admin_key}",
            "Content-Type": "application/json",
        },
        timeout=30.0,
    )
    response.raise_for_status()
    data = response.json()

    chunk_total = 0.0
    for bucket in data.get("data", []):
        for result in bucket.get("results", []):
            value = result.get("amount", {}).get("value", 0)
            chunk_total += float(value)
    return chunk_total


async def fetch_openai_costs(period: str = "monthly") -> Optional[Dict]:
    """Fetch costs from OpenAI Costs API."""
    admin_key = os.getenv("OPENAI_ADMIN_KEY")
    if not admin_key:
        logger.debug("OPENAI_ADMIN_KEY not configured")
        return None

    start_time, end_time = _get_period_timestamps(period)
    logger.info(f"OpenAI costs period: {datetime.fromtimestamp(start_time, tz=timezone.utc)} to {datetime.fromtimestamp(end_time, tz=timezone.utc)}")

    try:
        # Split into monthly chunks to avoid pagination issues
        chunks = _get_monthly_chunks(start_time, end_time)
        total_spend = 0.0

        async with httpx.AsyncClient() as client:
            # Fetch all chunks in parallel
            chunk_results = await asyncio.gather(
                *[_fetch_openai_chunk(client, admin_key, s, e) for s, e in chunks],
                return_exceptions=True,
            )
            for result in chunk_results:
                if isinstance(result, Exception):
                    logger.error(f"Chunk fetch failed: {result}")
                    continue
                total_spend += result

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
        logger.error(f"Failed to fetch OpenAI costs: {e}", exc_info=True)
        return None


async def _fetch_anthropic_costs_raw(
    admin_key: str,
    starting_at: str,
    ending_at: str,
) -> float:
    """Fetch raw costs from Anthropic (org-wide, no API key filtering available)."""
    total_spend = 0.0
    next_page = None

    async with httpx.AsyncClient() as client:
        while True:
            params = {
                "starting_at": starting_at,
                "ending_at": ending_at,
            }
            if next_page:
                params["page"] = next_page

            response = await client.get(
                ANTHROPIC_COSTS_URL,
                params=params,
                headers={
                    "x-api-key": admin_key,
                    "anthropic-version": "2023-06-01",
                    "Content-Type": "application/json",
                },
                timeout=30.0,
            )
            response.raise_for_status()
            data = response.json()

            for bucket in data.get("data", []):
                for result in bucket.get("results", []):
                    amount = result.get("amount", "0")
                    total_spend += float(amount)

            if not data.get("has_more"):
                break
            next_page = data.get("next_page")

    return total_spend


async def fetch_anthropic_costs(period: str = "monthly") -> Optional[Dict]:
    """Fetch costs from Anthropic Cost Report API (org-wide)."""
    admin_key = os.getenv("ANTHROPIC_ADMIN_KEY")
    if not admin_key:
        logger.debug("ANTHROPIC_ADMIN_KEY not configured")
        return None

    start_time, end_time = _get_period_timestamps(period)
    starting_at = datetime.fromtimestamp(start_time, tz=timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ")
    ending_at = datetime.fromtimestamp(end_time, tz=timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ")

    try:
        total_spend = await _fetch_anthropic_costs_raw(admin_key, starting_at, ending_at)
        return {
            "provider": "anthropic",
            "spend": round(total_spend, 4),
            "currency": "USD",
            "period": period,
            "scope": "organization",
            "fetched_at": datetime.now(timezone.utc).isoformat(),
        }
    except httpx.HTTPStatusError as e:
        logger.error(f"Anthropic Costs API error: {e.response.status_code} - {e.response.text}")
        return None
    except Exception as e:
        logger.error(f"Failed to fetch Anthropic costs: {e}")
        return None


async def fetch_anthropic_org_costs(period: str = "monthly") -> Optional[Dict]:
    """Fetch total org-wide costs from Anthropic (includes Claude Code, etc.)."""
    # Same as fetch_anthropic_costs since API key filtering isn't available
    return await fetch_anthropic_costs(period)


async def get_combined_costs(period: str = "monthly") -> Dict:
    """Fetch costs from all providers in parallel and combine."""
    openai_costs, anthropic_costs = await asyncio.gather(
        fetch_openai_costs(period),
        fetch_anthropic_costs(period),
    )

    total_spend = 0.0
    if openai_costs:
        total_spend += openai_costs.get("spend", 0)
    if anthropic_costs:
        total_spend += anthropic_costs.get("spend", 0)

    return {
        "period": period,
        "fetched_at": datetime.now(timezone.utc).isoformat(),
        "openai": openai_costs,
        "anthropic": anthropic_costs,
        "total_spend": round(total_spend, 4),
    }


async def get_org_costs(period: str = "monthly") -> Dict:
    """Fetch org-wide costs (includes Claude Code, subscriptions, etc.)."""
    anthropic_costs = await fetch_anthropic_org_costs(period)

    return {
        "period": period,
        "fetched_at": datetime.now(timezone.utc).isoformat(),
        "anthropic": anthropic_costs,
        "total_spend": round(anthropic_costs.get("spend", 0) if anthropic_costs else 0, 4),
    }
