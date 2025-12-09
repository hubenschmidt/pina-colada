"""Service layer for individual business logic."""

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
from repositories.organization_repository import get_organization_account_id
from repositories.individual_repository import (
    find_all_individuals_paginated,
    find_individual_by_id,
    create_individual as create_individual_repo,
    update_individual as update_individual_repo,
    delete_individual as delete_individual_repo,
    search_individuals as search_individuals_repo,
    get_individual_account_id,
    check_duplicate_individual,
    create_account_for_individual,
    update_account_industries,
    update_account_projects,
    find_individual_relationships,
    find_individual_relationship,
    create_individual_relationship as create_individual_relationship_repo,
    find_individual_relationship_by_id,
    delete_individual_relationship_by_id,
    IndividualCreate,
    IndividualUpdate,
    IndContactCreate,
    IndContactUpdate,
)

# Re-export Pydantic models for controllers
__all__ = [
    "IndividualCreate",
    "IndividualUpdate",
    "IndContactCreate",
    "IndContactUpdate",
]

logger = logging.getLogger(__name__)


def _normalize_linkedin_url(url: str | None) -> str | None:
    """Ensure LinkedIn URL has https:// prefix."""
    if not url:
        return url
    url = url.strip()
    if not url:
        return None
    if url.startswith("http://") or url.startswith("https://"):
        return url
    return f"https://{url}"


async def get_individuals_paginated(
    tenant_id: Optional[int],
    page: int,
    page_size: int,
    order_by: str,
    order: str,
    search: Optional[str] = None,
):
    """Get individuals with pagination."""
    return await find_all_individuals_paginated(
        tenant_id=tenant_id,
        page=page,
        page_size=page_size,
        order_by=order_by,
        order=order,
        search=search,
    )


async def search_individuals(query: str, tenant_id: Optional[int]):
    """Search individuals by name or email."""
    return await search_individuals_repo(query, tenant_id=tenant_id)


async def get_individual(individual_id: int):
    """Get individual by ID with contacts."""
    individual = await find_individual_by_id(individual_id)
    if not individual:
        raise HTTPException(status_code=404, detail="Individual not found")
    contacts = await find_contacts_by_account(individual.account_id) if individual.account_id else []
    return individual, contacts


async def create_individual(
    ind_data: Dict[str, Any],
    tenant_id: Optional[int],
    industry_ids: Optional[List[int]],
    project_ids: Optional[List[int]],
):
    """Create individual with associated account, industries, and projects."""
    # Normalize LinkedIn URL
    if "linkedin_url" in ind_data:
        ind_data["linkedin_url"] = _normalize_linkedin_url(ind_data["linkedin_url"])

    # Check for duplicate individual
    is_duplicate = await check_duplicate_individual(
        ind_data.get("first_name", ""),
        ind_data.get("last_name", ""),
        ind_data.get("email"),
        ind_data.get("linkedin_url"),
        ind_data.get("phone"),
    )
    if is_duplicate:
        raise HTTPException(status_code=400, detail="Individual already exists")

    # Create an Account for the Individual if not provided
    if not ind_data.get("account_id"):
        account_name = f"{ind_data.get('first_name', '')} {ind_data.get('last_name', '')}".strip()
        account_id = await create_account_for_individual(
            tenant_id=tenant_id,
            account_name=account_name,
            created_by=ind_data.get("created_by"),
            updated_by=ind_data.get("updated_by"),
            industry_ids=industry_ids,
            project_ids=project_ids,
        )
        ind_data["account_id"] = account_id

    individual = await create_individual_repo(ind_data)
    # Re-fetch to ensure relationships are loaded
    return await find_individual_by_id(individual.id)


async def update_individual(
    individual_id: int,
    ind_data: Dict[str, Any],
    industry_ids: Optional[List[int]],
    project_ids: Optional[List[int]],
    project_ids_provided: bool,
):
    """Update individual with industries and projects."""
    # Normalize LinkedIn URL
    if "linkedin_url" in ind_data:
        ind_data["linkedin_url"] = _normalize_linkedin_url(ind_data["linkedin_url"])

    individual = await update_individual_repo(individual_id, ind_data)
    if not individual:
        raise HTTPException(status_code=404, detail="Individual not found")

    # Update industries if provided
    if industry_ids is not None and individual.account_id:
        await update_account_industries(individual.account_id, industry_ids)

    # Update projects if explicitly provided
    if project_ids_provided and individual.account_id:
        await update_account_projects(individual.account_id, project_ids)

    # Re-fetch to get updated data
    return await find_individual_by_id(individual_id)


async def delete_individual(individual_id: int):
    """Delete an individual."""
    success = await delete_individual_repo(individual_id)
    if not success:
        raise HTTPException(status_code=404, detail="Individual not found")
    return True


# Contact management

async def get_individual_contacts(individual_id: int):
    """Get contacts for an individual."""
    individual = await find_individual_by_id(individual_id)
    if not individual:
        raise HTTPException(status_code=404, detail="Individual not found")
    return await find_contacts_by_account(individual.account_id) if individual.account_id else []


async def create_individual_contact(
    individual_id: int,
    contact_data: Dict[str, Any],
    user_id: Optional[int],
):
    """Create a contact for an individual."""
    individual = await find_individual_by_id(individual_id)
    if not individual or not individual.account_id:
        raise HTTPException(status_code=404, detail="Individual not found")

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

    # Link to individual's account
    await link_contact_to_account(contact.id, individual.account_id, contact_data.get("is_primary", False))

    # Link to another individual's account if provided (peer relationship)
    if contact_data.get("linked_individual_id"):
        peer_account_id = await get_individual_account_id(contact_data["linked_individual_id"])
        if peer_account_id:
            await link_contact_to_account(contact.id, peer_account_id, False)

    # Link to organization's account if provided
    if contact_data.get("organization_id"):
        org_account_id = await get_organization_account_id(contact_data["organization_id"])
        if org_account_id:
            await link_contact_to_account(contact.id, org_account_id, contact_data.get("is_primary", False))

    return await find_contact_by_id(contact.id)


async def update_individual_contact(
    individual_id: int,
    contact_id: int,
    contact_data: Dict[str, Any],
    user_id: Optional[int],
):
    """Update a contact for an individual."""
    individual = await find_individual_by_id(individual_id)
    if not individual or not individual.account_id:
        raise HTTPException(status_code=404, detail="Individual not found")

    # Check contact is linked to this individual's account
    if not await find_contact_account_link(contact_id, individual.account_id):
        raise HTTPException(status_code=404, detail="Contact not found for this individual")

    # Build update data
    update_data = {}
    for field in ["first_name", "last_name", "title", "department", "role", "email", "phone", "notes"]:
        if contact_data.get(field) is not None:
            update_data[field] = contact_data[field]

    # Handle is_primary specially
    if contact_data.get("is_primary"):
        await clear_primary_contacts_for_account(individual.account_id, contact_id)
        update_data["is_primary"] = True
    elif "is_primary" in contact_data and contact_data["is_primary"] is not None:
        update_data["is_primary"] = contact_data["is_primary"]

    update_data["updated_by"] = user_id
    return await update_contact_repo(contact_id, update_data)


async def delete_individual_contact(individual_id: int, contact_id: int):
    """Delete a contact from an individual."""
    individual = await find_individual_by_id(individual_id)
    if not individual or not individual.account_id:
        raise HTTPException(status_code=404, detail="Individual not found")

    # Check contact is linked to this individual's account
    if not await find_contact_account_link(contact_id, individual.account_id):
        raise HTTPException(status_code=404, detail="Contact not found for this individual")

    success = await delete_contact_repo(contact_id)
    if not success:
        raise HTTPException(status_code=404, detail="Contact not found")
    return True


# Individual-to-Individual Relationship management

async def get_individual_relationships(individual_id: int):
    """Get all relationships for an individual (both directions)."""
    individual = await find_individual_by_id(individual_id)
    if not individual:
        raise HTTPException(status_code=404, detail="Individual not found")

    return await find_individual_relationships(individual_id)


async def create_individual_relationship(
    from_individual_id: int,
    to_individual_id: int,
    relationship_type: Optional[str] = None,
    notes: Optional[str] = None,
):
    """Create a relationship between two individuals."""
    if from_individual_id == to_individual_id:
        raise HTTPException(status_code=400, detail="Cannot create relationship with self")

    from_ind = await find_individual_by_id(from_individual_id)
    if not from_ind:
        raise HTTPException(status_code=404, detail="Source individual not found")

    to_ind = await find_individual_by_id(to_individual_id)
    if not to_ind:
        raise HTTPException(status_code=404, detail="Target individual not found")

    # Check if relationship already exists
    if await find_individual_relationship(from_individual_id, to_individual_id):
        raise HTTPException(status_code=400, detail="Relationship already exists")

    return await create_individual_relationship_repo(
        from_individual_id=from_individual_id,
        to_individual_id=to_individual_id,
        relationship_type=relationship_type,
        notes=notes,
    )


async def delete_individual_relationship(from_individual_id: int, relationship_id: int):
    """Delete a relationship."""
    relationship = await find_individual_relationship_by_id(relationship_id, from_individual_id)
    if not relationship:
        raise HTTPException(status_code=404, detail="Relationship not found")

    await delete_individual_relationship_by_id(relationship_id)
    return True


# Signal management

async def get_individual_signals(individual_id: int, signal_type: Optional[str], limit: int):
    """Get signals for an individual (via its account)."""
    from repositories.signal_repository import find_signals_by_account

    individual = await find_individual_by_id(individual_id)
    if not individual:
        raise HTTPException(status_code=404, detail="Individual not found")
    if not individual.account_id:
        return []
    return await find_signals_by_account(individual.account_id, signal_type=signal_type, limit=limit)


async def create_individual_signal(individual_id: int, signal_data: Dict[str, Any]):
    """Create a signal for an individual (via its account)."""
    from repositories.signal_repository import create_signal as create_signal_repo

    individual = await find_individual_by_id(individual_id)
    if not individual:
        raise HTTPException(status_code=404, detail="Individual not found")
    if not individual.account_id:
        raise HTTPException(status_code=400, detail="Individual has no account")
    return await create_signal_repo(
        account_id=individual.account_id,
        signal_type=signal_data.get("signal_type"),
        headline=signal_data.get("headline"),
        description=signal_data.get("description"),
        signal_date=signal_data.get("signal_date"),
        source=signal_data.get("source"),
        source_url=signal_data.get("source_url"),
        sentiment=signal_data.get("sentiment"),
        relevance_score=signal_data.get("relevance_score")
    )


async def delete_individual_signal(individual_id: int, signal_id: int):
    """Delete a signal."""
    from repositories.signal_repository import find_signal_by_id, delete_signal as delete_signal_repo

    individual = await find_individual_by_id(individual_id)
    if not individual:
        raise HTTPException(status_code=404, detail="Individual not found")
    signal = await find_signal_by_id(signal_id)
    if not signal or signal.account_id != individual.account_id:
        raise HTTPException(status_code=404, detail="Signal not found")
    await delete_signal_repo(signal_id)
    return True
