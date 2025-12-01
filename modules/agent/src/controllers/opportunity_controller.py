"""Controller layer for opportunity routing to services."""

import logging
from typing import List, Optional, Dict, Any
from datetime import datetime
from lib.serialization import model_to_dict
from lib.decorators import handle_http_exceptions
from services.opportunity_service import (
    get_opportunities_paginated,
    create_opportunity as create_opportunity_service,
    get_opportunity as get_opportunity_service,
    update_opportunity as update_opportunity_service,
    delete_opportunity as delete_opportunity_service,
)

logger = logging.getLogger(__name__)


def _to_paged_response(count: int, page: int, limit: int, items: List) -> dict:
    """Convert to paged response format."""
    return {
        "items": items,
        "currentPage": page,
        "totalPages": max(1, (count + limit - 1) // limit),
        "total": count,
        "pageSize": limit,
    }


def _extract_company_info(organizations: list, individuals: list) -> tuple[str, str]:
    """Extract company name and type from account data."""
    if organizations:
        return organizations[0].get("name", ""), "Organization"
    if individuals:
        ind = individuals[0]
        first_name = ind.get("first_name", "")
        last_name = ind.get("last_name", "")
        return f"{last_name}, {first_name}".strip(", "), "Individual"
    return "", "Organization"


def _format_date(value) -> str:
    """Format datetime to ISO string, truncated to date."""
    if not value:
        return ""
    date_str = value.isoformat() if isinstance(value, datetime) else str(value)
    return date_str[:10] if date_str else ""


def _format_datetime(value) -> str:
    """Format datetime to full ISO string."""
    if not value:
        return ""
    return value.isoformat() if isinstance(value, datetime) else str(value)


def _get_account_contacts(opp) -> list:
    """Get contacts from opportunity's account."""
    if not opp.lead or not opp.lead.account:
        return []

    if opp.lead.account.organizations:
        return opp.lead.account.organizations[0].contacts or []

    if opp.lead.account.individuals:
        return opp.lead.account.individuals[0].contacts or []

    return []


def _build_contact_dict(contact) -> dict:
    """Build contact dictionary from ORM contact."""
    first_name = contact.first_name or ""
    last_name = contact.last_name or ""

    if not first_name and contact.individuals:
        first_name = contact.individuals[0].first_name or ""
    if not last_name and contact.individuals:
        last_name = contact.individuals[0].last_name or ""

    return {
        "id": contact.id,
        "first_name": first_name,
        "last_name": last_name,
        "email": contact.email or "",
        "phone": contact.phone or "",
        "title": contact.title,
        "is_primary": contact.is_primary,
    }


def _get_industries(opp) -> list:
    """Get industry names from opportunity's account."""
    if not opp.lead or not opp.lead.account or not opp.lead.account.industries:
        return []
    return [ind.name for ind in opp.lead.account.industries]


def _opportunity_to_response_dict(opp) -> Dict[str, Any]:
    """Convert opportunity ORM to response dictionary."""
    opp_dict = model_to_dict(opp, include_relationships=True)

    lead = opp_dict.get("lead") or {}
    account = lead.get("account") or {}
    company, company_type = _extract_company_info(
        account.get("organizations") or [],
        account.get("individuals") or [],
    )

    contacts = [_build_contact_dict(c) for c in _get_account_contacts(opp)]

    status = "Qualifying"
    if opp.lead and opp.lead.current_status:
        status = opp.lead.current_status.name

    created_at = opp_dict.get("created_at", "")

    project_ids = []
    if opp.lead and opp.lead.projects:
        project_ids = [p.id for p in opp.lead.projects]

    return {
        "id": str(opp_dict.get("id", "")),
        "account": company,
        "account_type": company_type,
        "title": lead.get("title", ""),
        "opportunity_name": opp_dict.get("opportunity_name", ""),
        "estimated_value": float(opp_dict.get("estimated_value")) if opp_dict.get("estimated_value") else None,
        "probability": float(opp_dict.get("probability")) if opp_dict.get("probability") else None,
        "expected_close_date": _format_date(opp_dict.get("expected_close_date")),
        "description": opp_dict.get("description"),
        "status": status,
        "source": lead.get("source", "manual"),
        "created_at": _format_datetime(created_at),
        "updated_at": opp_dict.get("updated_at", ""),
        "contacts": contacts,
        "industry": _get_industries(opp),
        "project_ids": project_ids,
    }


@handle_http_exceptions
async def get_opportunities(
    page: int, limit: int, order_by: str, order: str, search: Optional[str] = None, tenant_id: Optional[int] = None, project_id: Optional[int] = None
) -> dict:
    """Get all opportunities with pagination."""
    paginated, total_count = await get_opportunities_paginated(
        page, limit, order_by, order, search, tenant_id, project_id
    )
    items = [_opportunity_to_response_dict(opp) for opp in paginated]
    return _to_paged_response(total_count, page, limit, items)


@handle_http_exceptions
async def create_opportunity(data: Dict[str, Any]) -> Dict[str, Any]:
    """Create a new opportunity."""
    created = await create_opportunity_service(data)
    return _opportunity_to_response_dict(created)


@handle_http_exceptions
async def get_opportunity(opp_id: str) -> Dict[str, Any]:
    """Get an opportunity by ID."""
    opp = await get_opportunity_service(opp_id)
    return _opportunity_to_response_dict(opp)


@handle_http_exceptions
async def update_opportunity(opp_id: str, data: Dict[str, Any]) -> Dict[str, Any]:
    """Update an opportunity."""
    updated = await update_opportunity_service(opp_id, data)
    return _opportunity_to_response_dict(updated)


@handle_http_exceptions
async def delete_opportunity(opp_id: str) -> dict:
    """Delete an opportunity."""
    await delete_opportunity_service(opp_id)
    return {"success": True}
