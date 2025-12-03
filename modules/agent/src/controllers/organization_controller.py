"""Controller layer for organization routing to services."""

import logging
from typing import List, Optional, Dict, Any

from fastapi import Request

from lib.decorators import handle_http_exceptions
from repositories.organization_repository import (
    OrganizationCreate,
    OrganizationUpdate,
    OrgContactCreate,
    OrgContactUpdate,
    OrgTechnologyCreate,
    FundingRoundCreate,
    SignalCreate,
)
from services.organization_service import (
    get_organizations_paginated,
    search_organizations as search_organizations_service,
    get_organization as get_organization_service,
    create_organization as create_organization_service,
    update_organization as update_organization_service,
    delete_organization as delete_organization_service,
    get_organization_contacts as get_contacts_service,
    create_organization_contact as create_contact_service,
    update_organization_contact as update_contact_service,
    delete_organization_contact as delete_contact_service,
    get_technologies as get_technologies_service,
    add_technology as add_technology_service,
    remove_technology as remove_technology_service,
    get_funding_rounds as get_funding_rounds_service,
    create_funding_round as create_funding_round_service,
    delete_funding_round as delete_funding_round_service,
    get_signals as get_signals_service,
    create_signal as create_signal_service,
    delete_signal as delete_signal_service,
)

logger = logging.getLogger(__name__)

# Re-export for routes
__all__ = [
    "OrganizationCreate",
    "OrganizationUpdate",
    "OrgContactCreate",
    "OrgContactUpdate",
    "OrgTechnologyCreate",
    "FundingRoundCreate",
    "SignalCreate",
]


def _to_paged_response(count: int, page: int, limit: int, items: List) -> dict:
    """Convert to paged response format."""
    return {
        "items": items,
        "currentPage": page,
        "totalPages": max(1, (count + limit - 1) // limit),
        "total": count,
        "pageSize": limit,
    }


def _contact_to_dict(contact) -> dict:
    """Convert contact ORM to response dictionary."""
    first_name = contact.first_name
    last_name = contact.last_name
    if not first_name and contact.individuals:
        first_name = contact.individuals[0].first_name
    if not last_name and contact.individuals:
        last_name = contact.individuals[0].last_name

    return {
        "id": contact.id,
        "first_name": first_name or "",
        "last_name": last_name or "",
        "title": contact.title,
        "department": contact.department,
        "role": contact.role,
        "email": contact.email,
        "phone": contact.phone,
        "is_primary": contact.is_primary,
        "notes": contact.notes,
        "individuals": [{"id": i.id, "first_name": i.first_name, "last_name": i.last_name} for i in (contact.individuals or [])],
        "organizations": [{"id": o.id, "name": o.name} for o in (contact.organizations or [])],
        "created_at": contact.created_at.isoformat() if contact.created_at else None,
        "updated_at": contact.updated_at.isoformat() if contact.updated_at else None,
    }


def _technology_to_dict(org_tech) -> dict:
    """Convert technology ORM to response dictionary."""
    return {
        "organization_id": org_tech.organization_id,
        "technology_id": org_tech.technology_id,
        "technology": {
            "id": org_tech.technology.id,
            "name": org_tech.technology.name,
            "category": org_tech.technology.category,
            "vendor": org_tech.technology.vendor,
        } if org_tech.technology else None,
        "detected_at": org_tech.detected_at.isoformat() if org_tech.detected_at else None,
        "source": org_tech.source,
        "confidence": float(org_tech.confidence) if org_tech.confidence else None,
        "created_at": org_tech.created_at.isoformat() if org_tech.created_at else None,
        "updated_at": org_tech.updated_at.isoformat() if org_tech.updated_at else None,
    }


def _funding_round_to_dict(fr) -> dict:
    """Convert funding round ORM to response dictionary."""
    return {
        "id": fr.id,
        "organization_id": fr.organization_id,
        "round_type": fr.round_type,
        "amount": fr.amount,
        "announced_date": fr.announced_date.isoformat() if fr.announced_date else None,
        "lead_investor": fr.lead_investor,
        "source_url": fr.source_url,
        "created_at": fr.created_at.isoformat() if fr.created_at else None,
        "updated_at": fr.updated_at.isoformat() if fr.updated_at else None,
    }


def _signal_to_dict(s) -> dict:
    """Convert signal ORM to response dictionary."""
    return {
        "id": s.id,
        "organization_id": s.organization_id,
        "signal_type": s.signal_type,
        "headline": s.headline,
        "description": s.description,
        "signal_date": s.signal_date.isoformat() if s.signal_date else None,
        "source": s.source,
        "source_url": s.source_url,
        "sentiment": s.sentiment,
        "relevance_score": float(s.relevance_score) if s.relevance_score else None,
        "created_at": s.created_at.isoformat() if s.created_at else None,
        "updated_at": s.updated_at.isoformat() if s.updated_at else None,
    }


def _org_to_list_response(org) -> dict:
    """Convert organization ORM to response dictionary - optimized for list/table view."""
    industries = []
    if org.account and org.account.industries:
        industries = [ind.name for ind in org.account.industries]

    return {
        "id": org.id,
        "name": org.name,
        "website": org.website,
        "industries": industries,
        "employee_count_range": org.employee_count_range.label if org.employee_count_range else None,
        "funding_stage": org.funding_stage.label if org.funding_stage else None,
        "description": org.description,
    }


def _org_to_search_response(org) -> dict:
    """Convert organization ORM to response dictionary - minimal for search/autocomplete."""
    industries = []
    if org.account and org.account.industries:
        industries = [ind.name for ind in org.account.industries]

    return {
        "id": org.id,
        "name": org.name,
        "industries": industries,
    }


def _org_to_detail_response(org, contacts=None, include_research=False) -> dict:
    """Convert organization ORM to full response dictionary - for detail/edit views."""
    industries = []
    if org.account and org.account.industries:
        industries = [ind.name for ind in org.account.industries]

    projects = []
    if org.account and org.account.projects:
        projects = [{"id": p.id, "name": p.name} for p in org.account.projects]

    result = {
        "id": org.id,
        "name": org.name,
        "website": org.website,
        "phone": org.phone,
        "industries": industries,
        "projects": projects,
        "employee_count": org.employee_count,
        "employee_count_range_id": org.employee_count_range_id,
        "employee_count_range": org.employee_count_range.label if org.employee_count_range else None,
        "funding_stage_id": org.funding_stage_id,
        "funding_stage": org.funding_stage.label if org.funding_stage else None,
        "description": org.description,
        "revenue_range_id": org.revenue_range_id,
        "revenue_range": org.revenue_range.label if org.revenue_range else None,
        "founding_year": org.founding_year,
        "headquarters_city": org.headquarters_city,
        "headquarters_state": org.headquarters_state,
        "headquarters_country": org.headquarters_country,
        "company_type": org.company_type,
        "linkedin_url": org.linkedin_url,
        "crunchbase_url": org.crunchbase_url,
        "created_at": org.created_at.isoformat() if org.created_at else None,
        "updated_at": org.updated_at.isoformat() if org.updated_at else None,
    }

    if contacts is not None:
        result["contacts"] = [_contact_to_dict(c) for c in contacts]
        seen_ind_ids = set()
        related_individuals = []
        for contact in contacts:
            for ind in (contact.individuals or []):
                if ind.id not in seen_ind_ids:
                    seen_ind_ids.add(ind.id)
                    related_individuals.append({
                        "id": ind.id,
                        "name": f"{ind.first_name} {ind.last_name}".strip(),
                        "type": "individual",
                    })
        result["relationships"] = related_individuals

    if include_research:
        result["technologies"] = [_technology_to_dict(t) for t in (org.technologies or [])]
        result["funding_rounds"] = [_funding_round_to_dict(fr) for fr in (org.funding_rounds or [])]
        result["signals"] = [_signal_to_dict(s) for s in (org.signals or [])]

    return result


# Organization CRUD

@handle_http_exceptions
async def get_organizations(
    request: Request,
    page: int,
    limit: int,
    order_by: str,
    order: str,
    search: Optional[str] = None,
) -> dict:
    """Get organizations with pagination."""
    tenant_id = request.state.tenant_id
    organizations, total = await get_organizations_paginated(
        tenant_id, page, limit, order_by, order, search
    )
    items = [_org_to_list_response(org) for org in organizations]
    return _to_paged_response(total, page, limit, items)


@handle_http_exceptions
async def search_organizations(request: Request, query: str) -> List[dict]:
    """Search organizations by name."""
    tenant_id = request.state.tenant_id
    organizations = await search_organizations_service(query, tenant_id)
    return [_org_to_search_response(org) for org in organizations]


@handle_http_exceptions
async def get_organization(org_id: int) -> dict:
    """Get organization by ID with contacts and research data."""
    org, contacts = await get_organization_service(org_id)
    return _org_to_detail_response(org, contacts=contacts, include_research=True)


@handle_http_exceptions
async def create_organization(request: Request, data: OrganizationCreate) -> dict:
    """Create a new organization."""
    tenant_id = request.state.tenant_id
    user_id = request.state.user_id
    org_data = data.model_dump(exclude_none=True)
    industry_ids = org_data.pop("industry_ids", None)
    project_ids = org_data.pop("project_ids", None)
    org_data["created_by"] = user_id
    org_data["updated_by"] = user_id
    org = await create_organization_service(org_data, tenant_id, industry_ids, project_ids)
    return _org_to_detail_response(org)


@handle_http_exceptions
async def update_organization(request: Request, org_id: int, data: OrganizationUpdate) -> dict:
    """Update an organization."""
    user_id = request.state.user_id
    org_data = data.model_dump(exclude_unset=True)
    industry_ids = org_data.pop("industry_ids", None)
    project_ids = org_data.pop("project_ids", None)
    project_ids_provided = "project_ids" in data.model_dump(exclude_unset=True)
    org_data["updated_by"] = user_id
    org = await update_organization_service(org_id, org_data, industry_ids, project_ids, project_ids_provided)
    return _org_to_detail_response(org)


@handle_http_exceptions
async def delete_organization(org_id: int) -> dict:
    """Delete an organization."""
    await delete_organization_service(org_id)
    return {"success": True}


# Contact management

@handle_http_exceptions
async def get_organization_contacts(org_id: int) -> List[dict]:
    """Get contacts for an organization."""
    contacts = await get_contacts_service(org_id)
    return [_contact_to_dict(c) for c in contacts]


@handle_http_exceptions
async def create_organization_contact(request: Request, org_id: int, data: OrgContactCreate) -> dict:
    """Create a contact for an organization."""
    user_id = request.state.user_id
    contact = await create_contact_service(org_id, data.model_dump(), user_id)
    return _contact_to_dict(contact)


@handle_http_exceptions
async def update_organization_contact(
    request: Request,
    org_id: int,
    contact_id: int,
    data: OrgContactUpdate,
) -> dict:
    """Update a contact for an organization."""
    user_id = request.state.user_id
    contact = await update_contact_service(org_id, contact_id, data.model_dump(exclude_unset=True), user_id)
    return _contact_to_dict(contact)


@handle_http_exceptions
async def delete_organization_contact(org_id: int, contact_id: int) -> dict:
    """Delete a contact from an organization."""
    await delete_contact_service(org_id, contact_id)
    return {"success": True}


# Technology management

@handle_http_exceptions
async def get_organization_technologies(org_id: int) -> dict:
    """Get technologies for an organization."""
    technologies = await get_technologies_service(org_id)
    return {"technologies": [_technology_to_dict(t) for t in technologies]}


@handle_http_exceptions
async def add_organization_technology(
    org_id: int,
    tech_id: int,
    source: Optional[str],
    confidence: Optional[float],
) -> dict:
    """Add a technology to an organization."""
    org_tech = await add_technology_service(org_id, tech_id, source, confidence)
    return {"organization_technology": _technology_to_dict(org_tech)}


@handle_http_exceptions
async def remove_organization_technology(org_id: int, technology_id: int) -> dict:
    """Remove a technology from an organization."""
    await remove_technology_service(org_id, technology_id)
    return {"success": True}


# Funding round management

@handle_http_exceptions
async def get_organization_funding_rounds(org_id: int) -> dict:
    """Get funding rounds for an organization."""
    funding_rounds = await get_funding_rounds_service(org_id)
    return {"funding_rounds": [_funding_round_to_dict(fr) for fr in funding_rounds]}


@handle_http_exceptions
async def create_organization_funding_round(org_id: int, round_data: Dict[str, Any]) -> dict:
    """Create a funding round for an organization."""
    funding_round = await create_funding_round_service(org_id, round_data)
    return {"funding_round": _funding_round_to_dict(funding_round)}


@handle_http_exceptions
async def delete_organization_funding_round(org_id: int, round_id: int) -> dict:
    """Delete a funding round."""
    await delete_funding_round_service(org_id, round_id)
    return {"success": True}


# Signal management

@handle_http_exceptions
async def get_organization_signals(org_id: int, signal_type: Optional[str], limit: int) -> dict:
    """Get signals for an organization."""
    signals = await get_signals_service(org_id, signal_type, limit)
    return {"signals": [_signal_to_dict(s) for s in signals]}


@handle_http_exceptions
async def create_organization_signal(org_id: int, signal_data: Dict[str, Any]) -> dict:
    """Create a signal for an organization."""
    signal = await create_signal_service(org_id, signal_data)
    return {"signal": _signal_to_dict(signal)}


@handle_http_exceptions
async def delete_organization_signal(org_id: int, signal_id: int) -> dict:
    """Delete a signal."""
    await delete_signal_service(org_id, signal_id)
    return {"success": True}
