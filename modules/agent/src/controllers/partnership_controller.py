"""Controller layer for partnership routing to services."""

import logging
from typing import List, Optional, Dict, Any
from datetime import datetime
from lib.serialization import model_to_dict
from lib.decorators import handle_http_exceptions
from services.partnership_service import (
    get_partnerships_paginated,
    create_partnership as create_partnership_service,
    get_partnership as get_partnership_service,
    update_partnership as update_partnership_service,
    delete_partnership as delete_partnership_service,
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


def _get_account_contacts(partnership) -> list:
    """Get contacts from partnership's account."""
    if not partnership.lead or not partnership.lead.account:
        return []

    if partnership.lead.account.organizations:
        return partnership.lead.account.organizations[0].contacts or []

    if partnership.lead.account.individuals:
        return partnership.lead.account.individuals[0].contacts or []

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


def _get_industries(partnership) -> list:
    """Get industry names from partnership's account."""
    if not partnership.lead or not partnership.lead.account or not partnership.lead.account.industries:
        return []
    return [ind.name for ind in partnership.lead.account.industries]


def _partnership_to_response_dict(partnership) -> Dict[str, Any]:
    """Convert partnership ORM to response dictionary."""
    p_dict = model_to_dict(partnership, include_relationships=True)

    lead = p_dict.get("lead") or {}
    account = lead.get("account") or {}
    company, company_type = _extract_company_info(
        account.get("organizations") or [],
        account.get("individuals") or [],
    )

    contacts = [_build_contact_dict(c) for c in _get_account_contacts(partnership)]

    status = "Exploring"
    if partnership.lead and partnership.lead.current_status:
        status = partnership.lead.current_status.name

    created_at = p_dict.get("created_at", "")

    project_ids = []
    if partnership.lead and partnership.lead.projects:
        project_ids = [p.id for p in partnership.lead.projects]

    return {
        "id": str(p_dict.get("id", "")),
        "account": company,
        "account_type": company_type,
        "title": lead.get("title", ""),
        "partnership_name": p_dict.get("partnership_name", ""),
        "partnership_type": p_dict.get("partnership_type"),
        "start_date": _format_date(p_dict.get("start_date")),
        "end_date": _format_date(p_dict.get("end_date")),
        "description": p_dict.get("description"),
        "status": status,
        "source": lead.get("source", "manual"),
        "created_at": _format_datetime(created_at),
        "updated_at": p_dict.get("updated_at", ""),
        "contacts": contacts,
        "industry": _get_industries(partnership),
        "project_ids": project_ids,
    }


@handle_http_exceptions
async def get_partnerships(
    page: int, limit: int, order_by: str, order: str, search: Optional[str] = None, tenant_id: Optional[int] = None, project_id: Optional[int] = None
) -> dict:
    """Get all partnerships with pagination."""
    paginated, total_count = await get_partnerships_paginated(
        page, limit, order_by, order, search, tenant_id, project_id
    )
    items = [_partnership_to_response_dict(p) for p in paginated]
    return _to_paged_response(total_count, page, limit, items)


@handle_http_exceptions
async def create_partnership(data: Dict[str, Any]) -> Dict[str, Any]:
    """Create a new partnership."""
    created = await create_partnership_service(data)
    return _partnership_to_response_dict(created)


@handle_http_exceptions
async def get_partnership(partnership_id: str) -> Dict[str, Any]:
    """Get a partnership by ID."""
    partnership = await get_partnership_service(partnership_id)
    return _partnership_to_response_dict(partnership)


@handle_http_exceptions
async def update_partnership(partnership_id: str, data: Dict[str, Any]) -> Dict[str, Any]:
    """Update a partnership."""
    updated = await update_partnership_service(partnership_id, data)
    return _partnership_to_response_dict(updated)


@handle_http_exceptions
async def delete_partnership(partnership_id: str) -> dict:
    """Delete a partnership."""
    await delete_partnership_service(partnership_id)
    return {"success": True}
