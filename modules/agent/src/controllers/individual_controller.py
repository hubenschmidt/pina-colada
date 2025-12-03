"""Controller layer for individual routing to services."""

import logging
from typing import List, Optional, Dict, Any

from fastapi import Request

from lib.decorators import handle_http_exceptions
from services.individual_service import (
    IndividualCreate,
    IndividualUpdate,
    IndContactCreate,
    IndContactUpdate,
    get_individuals_paginated,
    search_individuals as search_individuals_service,
    get_individual as get_individual_service,
    create_individual as create_individual_service,
    update_individual as update_individual_service,
    delete_individual as delete_individual_service,
    get_individual_contacts as get_contacts_service,
    create_individual_contact as create_contact_service,
    update_individual_contact as update_contact_service,
    delete_individual_contact as delete_contact_service,
)

logger = logging.getLogger(__name__)

# Re-export for routes
__all__ = ["IndividualCreate", "IndividualUpdate", "IndContactCreate", "IndContactUpdate"]


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


def _get_industries(ind) -> list:
    """Extract industry names from individual's account."""
    if not ind.account or not ind.account.industries:
        return []
    return [industry.name for industry in ind.account.industries]


def _get_projects(ind) -> list:
    """Extract projects info from individual's account."""
    if not ind.account or not ind.account.projects:
        return []
    return [{"id": p.id, "name": p.name} for p in ind.account.projects]


def _get_related_orgs_from_contacts(contacts) -> list:
    """Extract unique organizations from contacts."""
    seen_ids = set()
    orgs = []
    for contact in contacts:
        for org in (contact.organizations or []):
            if org.id not in seen_ids:
                seen_ids.add(org.id)
                orgs.append({"id": org.id, "name": org.name, "type": "organization"})
    return orgs


def _ind_to_list_response(ind) -> dict:
    """Convert individual ORM to response dictionary - optimized for list/table view."""
    return {
        "id": ind.id,
        "first_name": ind.first_name,
        "last_name": ind.last_name,
        "title": ind.title,
        "industries": _get_industries(ind),
        "email": ind.email,
        "phone": ind.phone,
        "updated_at": ind.updated_at.isoformat() if ind.updated_at else None,
    }


def _ind_to_search_response(ind) -> dict:
    """Convert individual ORM to response dictionary - minimal for search/autocomplete."""
    return {
        "id": ind.id,
        "first_name": ind.first_name,
        "last_name": ind.last_name,
        "email": ind.email,
    }


def _ind_to_detail_response(ind, contacts=None, include_research=False) -> dict:
    """Convert individual ORM to full response dictionary - for detail/edit views."""
    result = {
        "id": ind.id,
        "first_name": ind.first_name,
        "last_name": ind.last_name,
        "email": ind.email,
        "phone": ind.phone,
        "linkedin_url": ind.linkedin_url,
        "title": ind.title,
        "description": ind.description,
        "industries": _get_industries(ind),
        "projects": _get_projects(ind),
        "twitter_url": ind.twitter_url,
        "github_url": ind.github_url,
        "bio": ind.bio,
        "seniority_level": ind.seniority_level,
        "department": ind.department,
        "is_decision_maker": ind.is_decision_maker,
        "reports_to_id": ind.reports_to_id,
        "created_at": ind.created_at.isoformat() if ind.created_at else None,
        "updated_at": ind.updated_at.isoformat() if ind.updated_at else None,
    }

    if contacts is not None:
        result["contacts"] = [_contact_to_dict(c) for c in contacts]
        result["relationships"] = _get_related_orgs_from_contacts(contacts)

    if include_research:
        if ind.reports_to:
            result["reports_to"] = {
                "id": ind.reports_to.id,
                "first_name": ind.reports_to.first_name,
                "last_name": ind.reports_to.last_name,
            }
        result["direct_reports"] = [
            {"id": dr.id, "first_name": dr.first_name, "last_name": dr.last_name, "title": dr.title}
            for dr in (ind.direct_reports or [])
        ]

    return result


# Individual CRUD

@handle_http_exceptions
async def get_individuals(
    request: Request,
    page: int,
    limit: int,
    order_by: str,
    order: str,
    search: Optional[str] = None,
) -> dict:
    """Get individuals with pagination."""
    tenant_id = request.state.tenant_id
    individuals, total = await get_individuals_paginated(
        tenant_id, page, limit, order_by, order, search
    )
    items = [_ind_to_list_response(ind) for ind in individuals]
    return _to_paged_response(total, page, limit, items)


@handle_http_exceptions
async def search_individuals(request: Request, query: str) -> List[dict]:
    """Search individuals by name or email."""
    tenant_id = request.state.tenant_id
    individuals = await search_individuals_service(query, tenant_id)
    return [_ind_to_search_response(ind) for ind in individuals]


@handle_http_exceptions
async def get_individual(individual_id: int) -> dict:
    """Get individual by ID with contacts and research data."""
    ind, contacts = await get_individual_service(individual_id)
    return _ind_to_detail_response(ind, contacts=contacts, include_research=True)


@handle_http_exceptions
async def create_individual(request: Request, data: IndividualCreate) -> dict:
    """Create a new individual."""
    tenant_id = request.state.tenant_id
    user_id = request.state.user_id
    ind_data = data.model_dump(exclude_none=True)
    industry_ids = ind_data.pop("industry_ids", None)
    project_ids = ind_data.pop("project_ids", None)
    ind_data["created_by"] = user_id
    ind_data["updated_by"] = user_id
    ind = await create_individual_service(ind_data, tenant_id, industry_ids, project_ids)
    return _ind_to_detail_response(ind, contacts=[])


@handle_http_exceptions
async def update_individual(request: Request, individual_id: int, data: IndividualUpdate) -> dict:
    """Update an individual."""
    user_id = request.state.user_id
    ind_data = data.model_dump(exclude_unset=True)
    industry_ids = ind_data.pop("industry_ids", None)
    project_ids = ind_data.pop("project_ids", None)
    project_ids_provided = "project_ids" in data.model_dump(exclude_unset=True)
    ind_data["updated_by"] = user_id
    ind = await update_individual_service(individual_id, ind_data, industry_ids, project_ids, project_ids_provided)
    return _ind_to_detail_response(ind)


@handle_http_exceptions
async def delete_individual(individual_id: int) -> dict:
    """Delete an individual."""
    await delete_individual_service(individual_id)
    return {"success": True}


# Contact management

@handle_http_exceptions
async def get_individual_contacts(individual_id: int) -> List[dict]:
    """Get contacts for an individual."""
    contacts = await get_contacts_service(individual_id)
    return [_contact_to_dict(c) for c in contacts]


@handle_http_exceptions
async def create_individual_contact(request: Request, individual_id: int, data: IndContactCreate) -> dict:
    """Create a contact for an individual."""
    user_id = request.state.user_id
    contact = await create_contact_service(individual_id, data.model_dump(), user_id)
    return _contact_to_dict(contact)


@handle_http_exceptions
async def update_individual_contact(
    request: Request,
    individual_id: int,
    contact_id: int,
    data: IndContactUpdate,
) -> dict:
    """Update a contact for an individual."""
    user_id = request.state.user_id
    contact = await update_contact_service(individual_id, contact_id, data.model_dump(exclude_unset=True), user_id)
    return _contact_to_dict(contact)


@handle_http_exceptions
async def delete_individual_contact(individual_id: int, contact_id: int) -> dict:
    """Delete a contact from an individual."""
    await delete_contact_service(individual_id, contact_id)
    return {"success": True}
