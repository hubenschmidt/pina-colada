"""Service layer for partnership business logic."""

import logging
from typing import Dict, List, Optional, Any
from datetime import datetime
from fastapi import HTTPException
from repositories.partnership_repository import (
    find_all_partnerships,
    create_partnership as create_partnership_repo,
    find_partnership_by_id,
    update_partnership as update_partnership_repo,
    delete_partnership as delete_partnership_repo,
)
from repositories.organization_repository import get_or_create_organization
from repositories.individual_repository import get_or_create_individual
from repositories.deal_repository import get_or_create_deal
from repositories.status_repository import find_status_by_name
from repositories.industry_repository import find_industry_by_name

logger = logging.getLogger(__name__)


async def _resolve_individual_account(data: Dict[str, Any], tenant_id: Optional[str]) -> tuple[int, str]:
    """Resolve account for Individual account type."""
    account_name = data.get("account", "")
    if not account_name:
        raise HTTPException(status_code=400, detail="account is required for Individual account type")
    parts = account_name.strip().split(", ")
    if len(parts) == 2:
        last_name, first_name = parts
        individual = await get_or_create_individual(first_name, last_name, tenant_id)
        if not individual.account_id:
            raise HTTPException(status_code=400, detail=f"Individual {account_name} has no account")
        return individual.account_id, account_name
    parts = account_name.strip().split(" ", 1)
    first_name = parts[0]
    last_name = parts[1] if len(parts) > 1 else ""
    individual = await get_or_create_individual(first_name, last_name, tenant_id)
    if not individual.account_id:
        raise HTTPException(status_code=400, detail=f"Individual {account_name} has no account")
    return individual.account_id, account_name


async def _resolve_organization_account(data: Dict[str, Any], tenant_id: Optional[str]) -> tuple[int, str]:
    """Resolve account for Organization account type."""
    organization_name = data.get("account")
    if not organization_name:
        raise HTTPException(status_code=400, detail="account is required")
    industry_ids = data.get("industry_ids")
    industry_input = data.get("industry")
    if not industry_ids and industry_input:
        industry_names = industry_input if isinstance(industry_input, list) else [industry_input]
        industry_ids = [
            ind.id for name in industry_names if name
            for ind in [await find_industry_by_name(name)] if ind
        ]
    org = await get_or_create_organization(organization_name, tenant_id, industry_ids if industry_ids else None)
    if not org.account_id:
        raise HTTPException(status_code=400, detail=f"Organization {organization_name} has no account")
    return org.account_id, org.name


def _parse_date(date_str: str | None) -> datetime | None:
    """Parse date string to datetime."""
    if not date_str:
        return None
    try:
        return datetime.fromisoformat(date_str.replace("Z", "+00:00"))
    except (ValueError, TypeError):
        return None


async def get_partnerships_paginated(
    page: int, limit: int, order_by: str, order: str, search: Optional[str] = None, tenant_id: Optional[int] = None, project_id: Optional[int] = None
) -> tuple[List[Any], int]:
    """Get all partnerships with pagination."""
    return await find_all_partnerships(
        page=page,
        page_size=limit,
        search=search,
        order_by=order_by,
        order=order,
        tenant_id=tenant_id,
        project_id=project_id
    )


async def create_partnership(data: Dict[str, Any]) -> Any:
    """Create a new partnership."""
    account_type = data.get("account_type", "Organization")
    tenant_id = data.get("tenant_id")
    resolve_account = _resolve_individual_account if account_type == "Individual" else _resolve_organization_account
    account_id, account_name = await resolve_account(data, tenant_id)

    deal_id = data.get("deal_id")
    if not deal_id:
        deal = await get_or_create_deal("Partnerships Pipeline")
        deal_id = deal.id

    status_id = data.get("current_status_id")
    if not status_id:
        status_name = data.get("status", "Exploring")
        status = await find_status_by_name(status_name, "partnership")
        status_id = status.id if status else None

    user_id = data.get("user_id")
    create_data: Dict[str, Any] = {
        "account_id": account_id,
        "account_name": account_name,
        "deal_id": deal_id,
        "current_status_id": status_id,
        "title": data.get("title", ""),
        "partnership_name": data.get("partnership_name", ""),
        "partnership_type": data.get("partnership_type"),
        "start_date": _parse_date(data.get("start_date")),
        "end_date": _parse_date(data.get("end_date")),
        "notes": data.get("notes"),
        "source": data.get("source", "manual"),
        "tenant_id": tenant_id,
        "project_ids": data.get("project_ids") or [],
        "created_by": user_id,
        "updated_by": user_id,
    }

    if not create_data["project_ids"]:
        raise HTTPException(status_code=400, detail="At least one project must be specified")

    return await create_partnership_repo(create_data)


async def get_partnership(partnership_id: str) -> Any:
    """Get a partnership by ID."""
    try:
        partnership_id_int = int(partnership_id)
    except ValueError:
        raise HTTPException(status_code=400, detail="Invalid partnership ID format")

    partnership = await find_partnership_by_id(partnership_id_int)
    if not partnership:
        raise HTTPException(status_code=404, detail="Partnership not found")

    return partnership


async def update_partnership(partnership_id: str, data: Dict[str, Any]) -> Any:
    """Update a partnership."""
    try:
        partnership_id_int = int(partnership_id)
    except ValueError:
        raise HTTPException(status_code=400, detail="Invalid partnership ID format")

    update_data: Dict[str, Any] = {}
    user_id = data.get("user_id")
    if user_id:
        update_data["updated_by"] = user_id

    organization_name = data.get("account")
    if organization_name:
        org = await get_or_create_organization(organization_name)
        update_data["organization_id"] = org.id

    allowed_fields = [
        "title", "partnership_name", "partnership_type",
        "notes", "source", "project_ids",
    ]
    update_data.update({k: data[k] for k in allowed_fields if k in data})

    if "start_date" in data:
        update_data["start_date"] = _parse_date(data["start_date"])
    if "end_date" in data:
        update_data["end_date"] = _parse_date(data["end_date"])

    if "current_status_id" in data:
        update_data["current_status_id"] = data["current_status_id"]
    elif "status" in data:
        status = await find_status_by_name(data["status"], "partnership")
        update_data["current_status_id"] = status.id if status else None

    updated = await update_partnership_repo(partnership_id_int, update_data)
    if not updated:
        raise HTTPException(status_code=404, detail="Partnership not found")

    return updated


async def delete_partnership(partnership_id: str) -> bool:
    """Delete a partnership."""
    try:
        partnership_id_int = int(partnership_id)
    except ValueError:
        raise HTTPException(status_code=400, detail="Invalid partnership ID format")

    deleted = await delete_partnership_repo(partnership_id_int)
    if not deleted:
        raise HTTPException(status_code=404, detail="Partnership not found")

    return deleted
