"""Service layer for individual business logic."""

import logging
from typing import Dict, List, Optional, Any

from fastapi import HTTPException
from sqlalchemy import select, delete, func, or_, update
from sqlalchemy.orm import selectinload

from lib.db import async_get_session
from models.Account import Account
from models.Industry import AccountIndustry
from models.AccountProject import AccountProject
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


async def _check_duplicate_individual(
    first_name: str,
    last_name: str,
    email: str | None,
    linkedin_url: str | None,
    phone: str | None,
) -> bool:
    """Check if an individual with same name and (email, linkedin, or phone) already exists."""
    if not email and not linkedin_url and not phone:
        return False

    async with async_get_session() as session:
        conditions = [
            func.lower(Individual.first_name) == first_name.lower(),
            func.lower(Individual.last_name) == last_name.lower(),
        ]

        match_conditions = []
        if email:
            match_conditions.append(func.lower(Individual.email) == email.lower())
        if linkedin_url:
            match_conditions.append(func.lower(Individual.linkedin_url) == linkedin_url.lower())
        if phone:
            match_conditions.append(Individual.phone == phone)

        stmt = select(Individual).where(*conditions, or_(*match_conditions))
        result = await session.execute(stmt)
        return result.scalar_one_or_none() is not None


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
    is_duplicate = await _check_duplicate_individual(
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
        async with async_get_session() as session:
            account_name = f"{ind_data.get('first_name', '')} {ind_data.get('last_name', '')}".strip()
            account = Account(
                tenant_id=tenant_id,
                name=account_name,
                created_by=ind_data.get("created_by"),
                updated_by=ind_data.get("updated_by"),
            )
            session.add(account)
            await session.flush()

            # Link industries to the account
            if industry_ids:
                for industry_id in industry_ids:
                    session.add(AccountIndustry(
                        account_id=account.id,
                        industry_id=industry_id
                    ))

            # Link projects to the account
            if project_ids:
                for project_id in project_ids:
                    account_project = AccountProject(account_id=account.id, project_id=project_id)
                    session.add(account_project)

            await session.commit()
            await session.refresh(account)
            ind_data["account_id"] = account.id

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
        async with async_get_session() as session:
            await session.execute(
                delete(AccountIndustry).where(AccountIndustry.account_id == individual.account_id)
            )
            for industry_id in industry_ids:
                session.add(AccountIndustry(
                    account_id=individual.account_id,
                    industry_id=industry_id
                ))
            await session.commit()

    # Update projects if explicitly provided
    if project_ids_provided and individual.account_id:
        async with async_get_session() as session:
            await session.execute(
                delete(AccountProject).where(AccountProject.account_id == individual.account_id)
            )
            if project_ids:
                for project_id in project_ids:
                    account_project = AccountProject(account_id=individual.account_id, project_id=project_id)
                    session.add(account_project)
            await session.commit()

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
    from models.IndividualRelationship import IndividualRelationship

    individual = await find_individual_by_id(individual_id)
    if not individual:
        raise HTTPException(status_code=404, detail="Individual not found")

    async with async_get_session() as session:
        # Get outgoing relationships
        stmt = (
            select(IndividualRelationship)
            .options(selectinload(IndividualRelationship.to_individual))
            .where(IndividualRelationship.from_individual_id == individual_id)
        )
        result = await session.execute(stmt)
        outgoing = list(result.scalars().all())

        # Get incoming relationships
        stmt = (
            select(IndividualRelationship)
            .options(selectinload(IndividualRelationship.from_individual))
            .where(IndividualRelationship.to_individual_id == individual_id)
        )
        result = await session.execute(stmt)
        incoming = list(result.scalars().all())

        return outgoing, incoming


async def create_individual_relationship(
    from_individual_id: int,
    to_individual_id: int,
    relationship_type: Optional[str] = None,
    notes: Optional[str] = None,
):
    """Create a relationship between two individuals."""
    from models.IndividualRelationship import IndividualRelationship

    if from_individual_id == to_individual_id:
        raise HTTPException(status_code=400, detail="Cannot create relationship with self")

    from_ind = await find_individual_by_id(from_individual_id)
    if not from_ind:
        raise HTTPException(status_code=404, detail="Source individual not found")

    to_ind = await find_individual_by_id(to_individual_id)
    if not to_ind:
        raise HTTPException(status_code=404, detail="Target individual not found")

    async with async_get_session() as session:
        # Check if relationship already exists
        stmt = select(IndividualRelationship).where(
            IndividualRelationship.from_individual_id == from_individual_id,
            IndividualRelationship.to_individual_id == to_individual_id,
        )
        result = await session.execute(stmt)
        if result.scalar_one_or_none():
            raise HTTPException(status_code=400, detail="Relationship already exists")

        relationship = IndividualRelationship(
            from_individual_id=from_individual_id,
            to_individual_id=to_individual_id,
            relationship_type=relationship_type,
            notes=notes,
        )
        session.add(relationship)
        await session.commit()
        await session.refresh(relationship)
        return relationship


async def delete_individual_relationship(from_individual_id: int, relationship_id: int):
    """Delete a relationship."""
    from models.IndividualRelationship import IndividualRelationship

    async with async_get_session() as session:
        stmt = select(IndividualRelationship).where(
            IndividualRelationship.id == relationship_id,
            or_(
                IndividualRelationship.from_individual_id == from_individual_id,
                IndividualRelationship.to_individual_id == from_individual_id,
            )
        )
        result = await session.execute(stmt)
        relationship = result.scalar_one_or_none()
        if not relationship:
            raise HTTPException(status_code=404, detail="Relationship not found")

        await session.delete(relationship)
        await session.commit()
        return True
