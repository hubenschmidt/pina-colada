"""Controller layer for organization routing to services."""

from typing import List, Optional

from fastapi import Request

from lib.decorators import handle_http_exceptions
from serializers.common import to_paged_response, contact_to_dict
from serializers.organization import (
    technology_to_dict,
    funding_round_to_dict,
    signal_to_dict,
    org_to_list_response,
    org_to_search_response,
    org_to_detail_response,
)
from services.organization_service import (
    OrganizationCreate,
    OrganizationUpdate,
    OrgContactCreate,
    OrgContactUpdate,
    OrgTechnologyCreate,
    FundingRoundCreate,
    SignalCreate,
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
    items = [org_to_list_response(org) for org in organizations]
    return to_paged_response(total, page, limit, items)


@handle_http_exceptions
async def search_organizations(request: Request, query: str) -> List[dict]:
    """Search organizations by name."""
    tenant_id = request.state.tenant_id
    organizations = await search_organizations_service(query, tenant_id)
    return [org_to_search_response(org) for org in organizations]


@handle_http_exceptions
async def get_organization(org_id: int) -> dict:
    """Get organization by ID with contacts and research data."""
    org, contacts = await get_organization_service(org_id)
    return org_to_detail_response(org, contacts=contacts, include_research=True)


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
    return org_to_detail_response(org)


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
    return org_to_detail_response(org)


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
    return [contact_to_dict(c) for c in contacts]


@handle_http_exceptions
async def create_organization_contact(request: Request, org_id: int, data: OrgContactCreate) -> dict:
    """Create a contact for an organization."""
    user_id = request.state.user_id
    contact = await create_contact_service(org_id, data.model_dump(), user_id)
    return contact_to_dict(contact)


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
    return contact_to_dict(contact)


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
    return {"technologies": [technology_to_dict(t) for t in technologies]}


@handle_http_exceptions
async def add_organization_technology(
    org_id: int,
    tech_id: int,
    source: Optional[str],
    confidence: Optional[float],
) -> dict:
    """Add a technology to an organization."""
    org_tech = await add_technology_service(org_id, tech_id, source, confidence)
    return {"organization_technology": technology_to_dict(org_tech)}


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
    return {"funding_rounds": [funding_round_to_dict(fr) for fr in funding_rounds]}


@handle_http_exceptions
async def create_organization_funding_round(org_id: int, round_data: Dict[str, Any]) -> dict:
    """Create a funding round for an organization."""
    funding_round = await create_funding_round_service(org_id, round_data)
    return {"funding_round": funding_round_to_dict(funding_round)}


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
    return {"signals": [signal_to_dict(s) for s in signals]}


@handle_http_exceptions
async def create_organization_signal(org_id: int, signal_data: Dict[str, Any]) -> dict:
    """Create a signal for an organization."""
    signal = await create_signal_service(org_id, signal_data)
    return {"signal": signal_to_dict(signal)}


@handle_http_exceptions
async def delete_organization_signal(org_id: int, signal_id: int) -> dict:
    """Delete a signal."""
    await delete_signal_service(org_id, signal_id)
    return {"success": True}
