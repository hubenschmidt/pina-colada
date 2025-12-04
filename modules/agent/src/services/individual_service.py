"""Service layer for individual business logic."""

import logging
from typing import Dict, List, Optional, Any

from fastapi import HTTPException
from sqlalchemy import select, delete, func, or_, update

from lib.db import async_get_session
from models.Account import Account
from models.Industry import AccountIndustry
from models.AccountProject import AccountProject
from models.Contact import Contact, ContactIndividual, ContactOrganization
from models.Individual import Individual
from repositories.individual_repository import (
    find_all_individuals_paginated,
    find_individual_by_id,
    create_individual as create_individual_repo,
    update_individual as update_individual_repo,
    delete_individual as delete_individual_repo,
    search_individuals as search_individuals_repo,
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
from repositories.contact_repository import find_contacts_by_individual, delete_contact

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
    contacts = await find_contacts_by_individual(individual_id)
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
    return await find_contacts_by_individual(individual_id)


async def create_individual_contact(
    individual_id: int,
    contact_data: Dict[str, Any],
    user_id: Optional[int],
):
    """Create a contact for an individual."""
    individual = await find_individual_by_id(individual_id)
    if not individual:
        raise HTTPException(status_code=404, detail="Individual not found")

    async with async_get_session() as session:
        contact = Contact(
            first_name=contact_data.get("first_name"),
            last_name=contact_data.get("last_name"),
            title=contact_data.get("title"),
            department=contact_data.get("department"),
            role=contact_data.get("role"),
            email=contact_data.get("email"),
            phone=contact_data.get("phone"),
            is_primary=contact_data.get("is_primary", False),
            notes=contact_data.get("notes"),
            created_by=user_id,
            updated_by=user_id,
        )
        session.add(contact)
        await session.flush()

        # Link to individual
        ind_link = ContactIndividual(contact_id=contact.id, individual_id=individual_id)
        session.add(ind_link)

        # Link to organization if provided
        if contact_data.get("organization_id"):
            org_link = ContactOrganization(
                contact_id=contact.id,
                organization_id=contact_data["organization_id"],
                is_primary=contact_data.get("is_primary", False),
            )
            session.add(org_link)

        await session.commit()

        stmt = select(Contact).where(Contact.id == contact.id)
        result = await session.execute(stmt)
        return result.scalar_one()


async def update_individual_contact(
    individual_id: int,
    contact_id: int,
    contact_data: Dict[str, Any],
    user_id: Optional[int],
):
    """Update a contact for an individual."""
    async with async_get_session() as session:
        stmt = select(ContactIndividual).where(
            ContactIndividual.contact_id == contact_id,
            ContactIndividual.individual_id == individual_id
        )
        result = await session.execute(stmt)
        link = result.scalar_one_or_none()
        if not link:
            raise HTTPException(status_code=404, detail="Contact not found for this individual")

        stmt = select(Contact).where(Contact.id == contact_id)
        result = await session.execute(stmt)
        contact = result.scalar_one()

        # Update fields if provided
        for field in ["first_name", "last_name", "title", "department", "role", "email", "phone", "notes"]:
            if contact_data.get(field) is not None:
                setattr(contact, field, contact_data[field])

        # Handle is_primary specially
        if contact_data.get("is_primary"):
            stmt = select(ContactIndividual.contact_id).where(
                ContactIndividual.individual_id == individual_id,
                ContactIndividual.contact_id != contact_id
            )
            result = await session.execute(stmt)
            other_contact_ids = [row[0] for row in result.fetchall()]
            if other_contact_ids:
                await session.execute(
                    update(Contact).where(Contact.id.in_(other_contact_ids)).values(is_primary=False)
                )
            contact.is_primary = True
        elif "is_primary" in contact_data and contact_data["is_primary"] is not None:
            contact.is_primary = contact_data["is_primary"]

        contact.updated_by = user_id
        await session.commit()

        stmt = select(Contact).where(Contact.id == contact_id)
        result = await session.execute(stmt)
        return result.scalar_one()


async def delete_individual_contact(individual_id: int, contact_id: int):
    """Delete a contact from an individual."""
    async with async_get_session() as session:
        stmt = select(ContactIndividual).where(
            ContactIndividual.contact_id == contact_id,
            ContactIndividual.individual_id == individual_id
        )
        result = await session.execute(stmt)
        link = result.scalar_one_or_none()
        if not link:
            raise HTTPException(status_code=404, detail="Contact not found for this individual")

    success = await delete_contact(contact_id)
    if not success:
        raise HTTPException(status_code=404, detail="Contact not found")
    return True
