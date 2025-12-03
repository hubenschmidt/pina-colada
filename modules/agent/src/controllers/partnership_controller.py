"""Controller layer for partnership routing to services."""

import logging
from typing import List, Optional, Dict, Any

from fastapi import Request

from lib.serialization import model_to_dict
from lib.decorators import handle_http_exceptions
from lib.date_utils import format_date, format_datetime, format_display_date
from repositories.partnership_repository import PartnershipCreate, PartnershipUpdate
from services.partnership_service import (
    get_partnerships_paginated,
    create_partnership as create_partnership_service,
    get_partnership as get_partnership_service,
    update_partnership as update_partnership_service,
    delete_partnership as delete_partnership_service,
)

logger = logging.getLogger(__name__)

# Re-export for routes
__all__ = ["PartnershipCreate", "PartnershipUpdate"]


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
        "created_at": contact.created_at.isoformat() if contact.created_at else None,
        "updated_at": contact.updated_at.isoformat() if contact.updated_at else None,
    }


def _get_industries(partnership) -> list:
    """Get industry names from partnership's account."""
    if not partnership.lead or not partnership.lead.account or not partnership.lead.account.industries:
        return []
    return [ind.name for ind in partnership.lead.account.industries]


def _partnership_to_list_dict(partnership) -> Dict[str, Any]:
    """Convert partnership ORM to dict - optimized for list/table view.

    Only returns fields needed for table columns:
    Account, Partnership Name, Type, Status, Start Date, End Date, Updated
    """
    p_dict = model_to_dict(partnership, include_relationships=True)

    lead = p_dict.get("lead") or {}
    account = lead.get("account") or {}
    company, _ = _extract_company_info(
        account.get("organizations") or [],
        account.get("individuals") or [],
    )

    status = "Exploring"
    if partnership.lead and partnership.lead.current_status:
        status = partnership.lead.current_status.name

    return {
        "id": str(p_dict.get("id", "")),
        "account": company,
        "partnership_name": p_dict.get("partnership_name", ""),
        "partnership_type": p_dict.get("partnership_type"),
        "status": status,
        "start_date": format_date(p_dict.get("start_date")),
        "formatted_start_date": format_display_date(p_dict.get("start_date")),
        "end_date": format_date(p_dict.get("end_date")),
        "formatted_end_date": format_display_date(p_dict.get("end_date")),
        "updated_at": format_datetime(p_dict.get("updated_at")),
        "formatted_updated_at": format_display_date(p_dict.get("updated_at")),
    }


def _partnership_to_response_dict(partnership) -> Dict[str, Any]:
    """Convert partnership ORM to response dictionary - full detail view."""
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
        "start_date": format_date(p_dict.get("start_date")),
        "formatted_start_date": format_display_date(p_dict.get("start_date")),
        "end_date": format_date(p_dict.get("end_date")),
        "formatted_end_date": format_display_date(p_dict.get("end_date")),
        "description": p_dict.get("description"),
        "status": status,
        "source": lead.get("source", "manual"),
        "created_at": format_datetime(created_at),
        "updated_at": format_datetime(p_dict.get("updated_at")),
        "contacts": contacts,
        "industry": _get_industries(partnership),
        "project_ids": project_ids,
    }


@handle_http_exceptions
async def get_partnerships(
    request: Request,
    page: int,
    limit: int,
    order_by: str,
    order: str,
    search: Optional[str] = None,
    project_id: Optional[int] = None,
) -> dict:
    """Get all partnerships with pagination."""
    tenant_id = getattr(request.state, "tenant_id", None)
    paginated, total_count = await get_partnerships_paginated(
        page, limit, order_by, order, search, tenant_id, project_id
    )
    items = [_partnership_to_list_dict(p) for p in paginated]
    return _to_paged_response(total_count, page, limit, items)


@handle_http_exceptions
async def create_partnership(request: Request, data: PartnershipCreate) -> Dict[str, Any]:
    """Create a new partnership."""
    partner_data = data.dict()
    partner_data["tenant_id"] = getattr(request.state, "tenant_id", None)
    partner_data["user_id"] = getattr(request.state, "user_id", None)
    created = await create_partnership_service(partner_data)
    return _partnership_to_response_dict(created)


@handle_http_exceptions
async def get_partnership(partnership_id: str) -> Dict[str, Any]:
    """Get a partnership by ID."""
    partnership = await get_partnership_service(partnership_id)
    return _partnership_to_response_dict(partnership)


@handle_http_exceptions
async def update_partnership(request: Request, partnership_id: str, data: PartnershipUpdate) -> Dict[str, Any]:
    """Update a partnership."""
    update_data = data.dict(exclude_unset=True)
    update_data["user_id"] = getattr(request.state, "user_id", None)
    updated = await update_partnership_service(partnership_id, update_data)
    return _partnership_to_response_dict(updated)


@handle_http_exceptions
async def delete_partnership(partnership_id: str) -> dict:
    """Delete a partnership."""
    await delete_partnership_service(partnership_id)
    return {"success": True}
