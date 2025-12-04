"""Routes for organizations API endpoints."""

from typing import Optional

from fastapi import APIRouter, Request, Query

from controllers.organization_controller import (
    add_organization_technology,
    create_organization,
    create_organization_contact,
    create_organization_funding_round,
    create_organization_signal,
    delete_organization,
    delete_organization_contact,
    delete_organization_funding_round,
    delete_organization_signal,
    get_organization,
    get_organization_contacts,
    get_organization_funding_rounds,
    get_organization_signals,
    get_organization_technologies,
    get_organizations,
    remove_organization_technology,
    search_organizations,
    update_organization,
    update_organization_contact,
)
from schemas.organization import (
    FundingRoundCreate,
    OrgContactCreate,
    OrgContactUpdate,
    OrgTechnologyCreate,
    OrganizationCreate,
    OrganizationUpdate,
    SignalCreate,
)
from lib.auth import require_auth
from lib.error_logging import log_errors


router = APIRouter(prefix="/organizations", tags=["organizations"])


# Organization CRUD

@router.get("")
@log_errors
@require_auth
async def get_organizations_route(
    request: Request,
    page: int = Query(1, ge=1),
    limit: int = Query(50, ge=1, le=100),
    order_by: str = Query("updated_at", alias="orderBy"),
    order: str = Query("DESC", regex="^(ASC|DESC)$"),
    search: Optional[str] = Query(None),
):
    """Get organizations with pagination, sorting, and search."""
    return await get_organizations(request, page, limit, order_by, order, search)


@router.get("/search")
@log_errors
@require_auth
async def search_organizations_route(request: Request, q: Optional[str] = Query(None, min_length=1)):
    """Search organizations by name."""
    if not q:
        return []
    return await search_organizations(request, q)


@router.get("/{org_id}")
@log_errors
@require_auth
async def get_organization_route(request: Request, org_id: int):
    """Get a single organization by ID."""
    return await get_organization(org_id)


@router.post("")
@log_errors
@require_auth
async def create_organization_route(request: Request, data: OrganizationCreate):
    """Create a new organization."""
    return await create_organization(request, data)


@router.put("/{org_id}")
@log_errors
@require_auth
async def update_organization_route(request: Request, org_id: int, data: OrganizationUpdate):
    """Update an existing organization."""
    return await update_organization(request, org_id, data)


@router.delete("/{org_id}")
@log_errors
@require_auth
async def delete_organization_route(request: Request, org_id: int):
    """Delete an organization."""
    return await delete_organization(org_id)


# Contact management

@router.get("/{org_id}/contacts")
@log_errors
@require_auth
async def get_organization_contacts_route(request: Request, org_id: int):
    """Get all contacts for an organization."""
    return await get_organization_contacts(org_id)


@router.post("/{org_id}/contacts")
@log_errors
@require_auth
async def create_organization_contact_route(request: Request, org_id: int, data: OrgContactCreate):
    """Add a contact to an organization."""
    return await create_organization_contact(request, org_id, data)


@router.put("/{org_id}/contacts/{contact_id}")
@log_errors
@require_auth
async def update_organization_contact_route(request: Request, org_id: int, contact_id: int, data: OrgContactUpdate):
    """Update a contact for an organization."""
    return await update_organization_contact(request, org_id, contact_id, data)


@router.delete("/{org_id}/contacts/{contact_id}")
@log_errors
@require_auth
async def delete_organization_contact_route(request: Request, org_id: int, contact_id: int):
    """Remove a contact from an organization."""
    return await delete_organization_contact(org_id, contact_id)


# Technology management

@router.get("/{org_id}/technologies")
@log_errors
@require_auth
async def get_organization_technologies_route(request: Request, org_id: int):
    """Get all technologies for an organization."""
    return await get_organization_technologies(org_id)


@router.post("/{org_id}/technologies")
@log_errors
@require_auth
async def add_organization_technology_route(request: Request, org_id: int, data: OrgTechnologyCreate):
    """Add a technology to an organization."""
    return await add_organization_technology(org_id, data.technology_id, data.source, data.confidence)


@router.delete("/{org_id}/technologies/{technology_id}")
@log_errors
@require_auth
async def remove_organization_technology_route(request: Request, org_id: int, technology_id: int):
    """Remove a technology from an organization."""
    return await remove_organization_technology(org_id, technology_id)


# Funding round management

@router.get("/{org_id}/funding-rounds")
@log_errors
@require_auth
async def get_organization_funding_rounds_route(request: Request, org_id: int):
    """Get all funding rounds for an organization."""
    return await get_organization_funding_rounds(org_id)


@router.post("/{org_id}/funding-rounds")
@log_errors
@require_auth
async def create_organization_funding_round_route(request: Request, org_id: int, data: FundingRoundCreate):
    """Create a funding round for an organization."""
    return await create_organization_funding_round(org_id, data.model_dump())


@router.delete("/{org_id}/funding-rounds/{round_id}")
@log_errors
@require_auth
async def delete_organization_funding_round_route(request: Request, org_id: int, round_id: int):
    """Delete a funding round."""
    return await delete_organization_funding_round(org_id, round_id)


# Signal management

@router.get("/{org_id}/signals")
@log_errors
@require_auth
async def get_organization_signals_route(
    request: Request,
    org_id: int,
    signal_type: Optional[str] = Query(None),
    limit: int = Query(20, le=100),
):
    """Get signals for an organization."""
    return await get_organization_signals(org_id, signal_type, limit)


@router.post("/{org_id}/signals")
@log_errors
@require_auth
async def create_organization_signal_route(request: Request, org_id: int, data: SignalCreate):
    """Create a signal for an organization."""
    return await create_organization_signal(org_id, data.model_dump())


@router.delete("/{org_id}/signals/{signal_id}")
@log_errors
@require_auth
async def delete_organization_signal_route(request: Request, org_id: int, signal_id: int):
    """Delete a signal."""
    return await delete_organization_signal(org_id, signal_id)
