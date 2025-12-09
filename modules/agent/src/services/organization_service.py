"""Service layer for organization business logic."""

import logging
from typing import Dict, List, Optional, Any

from fastapi import HTTPException

from repositories.contact_repository import (
    create_contact as create_contact_repo,
    update_contact as update_contact_repo,
    delete_contact as delete_contact_repo,
    find_contact_by_id,
    find_contacts_by_account,
    link_contact_to_account,
    find_contact_account_link,
    clear_primary_contacts_for_account,
)
from repositories.organization_repository import (
    find_all_organizations,
    find_organization_by_id,
    create_organization as create_organization_repo,
    update_organization as update_organization_repo,
    delete_organization as delete_organization_repo,
    search_organizations as search_organizations_repo,
    get_organization_account_id,
    create_account_for_organization,
    update_account_industries,
    update_account_projects,
    find_organization_relationships as find_organization_relationships_repo,
    find_organization_relationship,
    create_organization_relationship as create_organization_relationship_repo,
    find_organization_relationship_by_id,
    delete_organization_relationship_by_id,
    OrganizationCreate,
    OrganizationUpdate,
    OrgContactCreate,
    OrgContactUpdate,
    OrgTechnologyCreate,
    FundingRoundCreate,
    SignalCreate,
)
from repositories.individual_repository import get_individual_account_id

# Re-export Pydantic models for controllers
__all__ = [
    "OrganizationCreate",
    "OrganizationUpdate",
    "OrgContactCreate",
    "OrgContactUpdate",
    "OrgTechnologyCreate",
    "FundingRoundCreate",
    "SignalCreate",
]
from repositories.technology_repository import (
    find_organization_technologies,
    add_organization_technology,
    remove_organization_technology,
)
from repositories.funding_round_repository import (
    find_funding_rounds_by_org,
    create_funding_round as create_funding_round_repo,
    delete_funding_round as delete_funding_round_repo,
    find_funding_round_by_id,
)
from repositories.signal_repository import (
    find_signals_by_account,
    create_signal as create_signal_repo,
    delete_signal as delete_signal_repo,
    find_signal_by_id,
)

logger = logging.getLogger(__name__)


async def get_organizations_paginated(
    tenant_id: Optional[int],
    page: int,
    page_size: int,
    order_by: str,
    order: str,
    search: Optional[str] = None,
):
    """Get organizations with pagination."""
    return await find_all_organizations(
        tenant_id=tenant_id,
        page=page,
        page_size=page_size,
        order_by=order_by,
        order=order,
        search=search,
    )


async def search_organizations(query: str, tenant_id: Optional[int]):
    """Search organizations by name."""
    return await search_organizations_repo(query, tenant_id=tenant_id)


async def get_organization(org_id: int):
    """Get organization by ID with contacts."""
    org = await find_organization_by_id(org_id)
    if not org:
        raise HTTPException(status_code=404, detail="Organization not found")
    contacts = await find_contacts_by_account(org.account_id) if org.account_id else []
    return org, contacts


async def create_organization(
    org_data: Dict[str, Any],
    tenant_id: Optional[int],
    industry_ids: Optional[List[int]],
    project_ids: Optional[List[int]],
):
    """Create organization with associated account, industries, and projects."""
    # Create an Account for the Organization if not provided
    if not org_data.get("account_id"):
        account_id = await create_account_for_organization(
            tenant_id=tenant_id,
            account_name=org_data.get("name", ""),
            created_by=org_data.get("created_by"),
            updated_by=org_data.get("updated_by"),
            industry_ids=industry_ids,
            project_ids=project_ids,
        )
        org_data["account_id"] = account_id

    try:
        return await create_organization_repo(org_data)
    except Exception as e:
        if "idx_organization_name_lower" in str(e):
            raise HTTPException(status_code=400, detail="An organization with this name already exists")
        raise


async def update_organization(
    org_id: int,
    org_data: Dict[str, Any],
    industry_ids: Optional[List[int]],
    project_ids: Optional[List[int]],
    project_ids_provided: bool,
):
    """Update organization with industries and projects."""
    org = await update_organization_repo(org_id, org_data)
    if not org:
        raise HTTPException(status_code=404, detail="Organization not found")

    # Update industries if provided
    if industry_ids is not None and org.account_id:
        await update_account_industries(org.account_id, industry_ids)

    # Update projects if explicitly provided
    if project_ids_provided and org.account_id:
        await update_account_projects(org.account_id, project_ids)

    # Re-fetch to get updated data
    return await find_organization_by_id(org_id)


async def delete_organization(org_id: int):
    """Delete an organization."""
    success = await delete_organization_repo(org_id)
    if not success:
        raise HTTPException(status_code=404, detail="Organization not found")
    return True


# Contact management

async def get_organization_contacts(org_id: int):
    """Get contacts for an organization."""
    org = await find_organization_by_id(org_id)
    if not org:
        raise HTTPException(status_code=404, detail="Organization not found")
    return await find_contacts_by_account(org.account_id) if org.account_id else []


async def create_organization_contact(
    org_id: int,
    contact_data: Dict[str, Any],
    user_id: Optional[int],
):
    """Create a contact for an organization."""
    org = await find_organization_by_id(org_id)
    if not org or not org.account_id:
        raise HTTPException(status_code=404, detail="Organization not found")

    # Create contact via repository
    contact = await create_contact_repo({
        "first_name": contact_data.get("first_name"),
        "last_name": contact_data.get("last_name"),
        "title": contact_data.get("title"),
        "department": contact_data.get("department"),
        "role": contact_data.get("role"),
        "email": contact_data.get("email"),
        "phone": contact_data.get("phone"),
        "is_primary": contact_data.get("is_primary", False),
        "notes": contact_data.get("notes"),
        "created_by": user_id,
        "updated_by": user_id,
    })

    # Link to organization's account
    await link_contact_to_account(contact.id, org.account_id, contact_data.get("is_primary", False))

    # Link to another organization's account if provided (partner relationship)
    if contact_data.get("linked_organization_id"):
        partner_account_id = await get_organization_account_id(contact_data["linked_organization_id"])
        if partner_account_id:
            await link_contact_to_account(contact.id, partner_account_id, False)

    # Link to individual's account if provided (for relationships)
    if contact_data.get("individual_id"):
        ind_account_id = await get_individual_account_id(contact_data["individual_id"])
        if ind_account_id:
            await link_contact_to_account(contact.id, ind_account_id, False)

    return await find_contact_by_id(contact.id)


async def update_organization_contact(
    org_id: int,
    contact_id: int,
    contact_data: Dict[str, Any],
    user_id: Optional[int],
):
    """Update a contact for an organization."""
    org = await find_organization_by_id(org_id)
    if not org or not org.account_id:
        raise HTTPException(status_code=404, detail="Organization not found")

    # Check contact is linked to this organization's account
    if not await find_contact_account_link(contact_id, org.account_id):
        raise HTTPException(status_code=404, detail="Contact not found for this organization")

    # Build update data
    update_data = {}
    for field in ["first_name", "last_name", "title", "department", "role", "email", "phone", "notes"]:
        if contact_data.get(field) is not None:
            update_data[field] = contact_data[field]

    # Handle is_primary specially
    if contact_data.get("is_primary"):
        await clear_primary_contacts_for_account(org.account_id, contact_id)
        update_data["is_primary"] = True
    elif "is_primary" in contact_data and contact_data["is_primary"] is not None:
        update_data["is_primary"] = contact_data["is_primary"]

    update_data["updated_by"] = user_id
    return await update_contact_repo(contact_id, update_data)


async def delete_organization_contact(org_id: int, contact_id: int):
    """Delete a contact from an organization."""
    org = await find_organization_by_id(org_id)
    if not org or not org.account_id:
        raise HTTPException(status_code=404, detail="Organization not found")

    # Check contact is linked to this organization's account
    if not await find_contact_account_link(contact_id, org.account_id):
        raise HTTPException(status_code=404, detail="Contact not found for this organization")

    success = await delete_contact_repo(contact_id)
    if not success:
        raise HTTPException(status_code=404, detail="Contact not found")
    return True


# Technology management

async def get_technologies(org_id: int):
    """Get technologies for an organization."""
    org = await find_organization_by_id(org_id)
    if not org:
        raise HTTPException(status_code=404, detail="Organization not found")
    return await find_organization_technologies(org_id)


async def add_technology(org_id: int, tech_id: int, source: Optional[str], confidence: Optional[float]):
    """Add a technology to an organization."""
    org = await find_organization_by_id(org_id)
    if not org:
        raise HTTPException(status_code=404, detail="Organization not found")
    return await add_organization_technology(
        org_id=org_id,
        tech_id=tech_id,
        source=source,
        confidence=confidence
    )


async def remove_technology(org_id: int, technology_id: int):
    """Remove a technology from an organization."""
    success = await remove_organization_technology(org_id, technology_id)
    if not success:
        raise HTTPException(status_code=404, detail="Technology not found for this organization")
    return True


# Funding round management

async def get_funding_rounds(org_id: int):
    """Get funding rounds for an organization."""
    org = await find_organization_by_id(org_id)
    if not org:
        raise HTTPException(status_code=404, detail="Organization not found")
    return await find_funding_rounds_by_org(org_id)


async def create_funding_round(org_id: int, round_data: Dict[str, Any]):
    """Create a funding round for an organization."""
    org = await find_organization_by_id(org_id)
    if not org:
        raise HTTPException(status_code=404, detail="Organization not found")
    return await create_funding_round_repo(
        org_id=org_id,
        round_type=round_data.get("round_type"),
        amount=round_data.get("amount"),
        announced_date=round_data.get("announced_date"),
        lead_investor=round_data.get("lead_investor"),
        source_url=round_data.get("source_url")
    )


async def delete_funding_round(org_id: int, round_id: int):
    """Delete a funding round."""
    funding_round = await find_funding_round_by_id(round_id)
    if not funding_round or funding_round.organization_id != org_id:
        raise HTTPException(status_code=404, detail="Funding round not found")
    await delete_funding_round_repo(round_id)
    return True


# Signal management

async def get_signals(org_id: int, signal_type: Optional[str], limit: int):
    """Get signals for an organization (via its account)."""
    org = await find_organization_by_id(org_id)
    if not org:
        raise HTTPException(status_code=404, detail="Organization not found")
    if not org.account_id:
        return []
    return await find_signals_by_account(org.account_id, signal_type=signal_type, limit=limit)


async def create_signal(org_id: int, signal_data: Dict[str, Any]):
    """Create a signal for an organization (via its account)."""
    org = await find_organization_by_id(org_id)
    if not org:
        raise HTTPException(status_code=404, detail="Organization not found")
    if not org.account_id:
        raise HTTPException(status_code=400, detail="Organization has no account")
    return await create_signal_repo(
        account_id=org.account_id,
        signal_type=signal_data.get("signal_type"),
        headline=signal_data.get("headline"),
        description=signal_data.get("description"),
        signal_date=signal_data.get("signal_date"),
        source=signal_data.get("source"),
        source_url=signal_data.get("source_url"),
        sentiment=signal_data.get("sentiment"),
        relevance_score=signal_data.get("relevance_score")
    )


async def delete_signal(org_id: int, signal_id: int):
    """Delete a signal."""
    org = await find_organization_by_id(org_id)
    if not org:
        raise HTTPException(status_code=404, detail="Organization not found")
    signal = await find_signal_by_id(signal_id)
    if not signal or signal.account_id != org.account_id:
        raise HTTPException(status_code=404, detail="Signal not found")
    await delete_signal_repo(signal_id)
    return True


# Organization-to-Organization Relationship management

async def get_organization_relationships(org_id: int):
    """Get all relationships for an organization (both directions)."""
    org = await find_organization_by_id(org_id)
    if not org:
        raise HTTPException(status_code=404, detail="Organization not found")

    return await find_organization_relationships_repo(org_id)


async def create_organization_relationship(
    from_org_id: int,
    to_org_id: int,
    relationship_type: str = None,
    notes: str = None,
):
    """Create a relationship between two organizations."""
    if from_org_id == to_org_id:
        raise HTTPException(status_code=400, detail="Cannot create relationship with self")

    from_org = await find_organization_by_id(from_org_id)
    if not from_org:
        raise HTTPException(status_code=404, detail="Source organization not found")

    to_org = await find_organization_by_id(to_org_id)
    if not to_org:
        raise HTTPException(status_code=404, detail="Target organization not found")

    # Check if relationship already exists
    if await find_organization_relationship(from_org_id, to_org_id):
        raise HTTPException(status_code=400, detail="Relationship already exists")

    return await create_organization_relationship_repo(
        from_org_id=from_org_id,
        to_org_id=to_org_id,
        relationship_type=relationship_type,
        notes=notes,
    )


async def delete_organization_relationship(from_org_id: int, relationship_id: int):
    """Delete a relationship."""
    relationship = await find_organization_relationship_by_id(relationship_id, from_org_id)
    if not relationship:
        raise HTTPException(status_code=404, detail="Relationship not found")

    await delete_organization_relationship_by_id(relationship_id)
    return True
