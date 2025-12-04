"""Service layer for organization business logic."""

import logging
from typing import Dict, List, Optional, Any

from fastapi import HTTPException
from sqlalchemy import select, delete, update
from sqlalchemy.exc import IntegrityError

from lib.db import async_get_session
from models.Account import Account
from models.Industry import AccountIndustry
from models.AccountProject import AccountProject
from models.Contact import Contact, ContactOrganization
from repositories.organization_repository import (
    find_all_organizations,
    find_organization_by_id,
    create_organization as create_organization_repo,
    update_organization as update_organization_repo,
    delete_organization as delete_organization_repo,
    search_organizations as search_organizations_repo,
    OrganizationCreate,
    OrganizationUpdate,
    OrgContactCreate,
    OrgContactUpdate,
    OrgTechnologyCreate,
    FundingRoundCreate,
    SignalCreate,
)

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
from repositories.contact_repository import find_contacts_by_organization, delete_contact
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
from repositories.company_signal_repository import (
    find_signals_by_org,
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
    contacts = await find_contacts_by_organization(org_id)
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
        async with async_get_session() as session:
            account = Account(
                tenant_id=tenant_id,
                name=org_data.get("name", ""),
                created_by=org_data.get("created_by"),
                updated_by=org_data.get("updated_by"),
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
            org_data["account_id"] = account.id

    try:
        return await create_organization_repo(org_data)
    except IntegrityError as e:
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
        async with async_get_session() as session:
            await session.execute(
                delete(AccountIndustry).where(AccountIndustry.account_id == org.account_id)
            )
            for industry_id in industry_ids:
                session.add(AccountIndustry(
                    account_id=org.account_id,
                    industry_id=industry_id
                ))
            await session.commit()

    # Update projects if explicitly provided
    if project_ids_provided and org.account_id:
        async with async_get_session() as session:
            await session.execute(
                delete(AccountProject).where(AccountProject.account_id == org.account_id)
            )
            if project_ids:
                for project_id in project_ids:
                    account_project = AccountProject(account_id=org.account_id, project_id=project_id)
                    session.add(account_project)
            await session.commit()

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
    return await find_contacts_by_organization(org_id)


async def create_organization_contact(
    org_id: int,
    contact_data: Dict[str, Any],
    user_id: Optional[int],
):
    """Create a contact for an organization."""
    org = await find_organization_by_id(org_id)
    if not org:
        raise HTTPException(status_code=404, detail="Organization not found")

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

        org_link = ContactOrganization(
            contact_id=contact.id,
            organization_id=org_id,
            is_primary=contact_data.get("is_primary", False),
        )
        session.add(org_link)
        await session.commit()

        stmt = select(Contact).where(Contact.id == contact.id)
        result = await session.execute(stmt)
        return result.scalar_one()


async def update_organization_contact(
    org_id: int,
    contact_id: int,
    contact_data: Dict[str, Any],
    user_id: Optional[int],
):
    """Update a contact for an organization."""
    async with async_get_session() as session:
        stmt = select(ContactOrganization).where(
            ContactOrganization.contact_id == contact_id,
            ContactOrganization.organization_id == org_id
        )
        result = await session.execute(stmt)
        link = result.scalar_one_or_none()
        if not link:
            raise HTTPException(status_code=404, detail="Contact not found for this organization")

        stmt = select(Contact).where(Contact.id == contact_id)
        result = await session.execute(stmt)
        contact = result.scalar_one()

        # Update fields if provided
        for field in ["first_name", "last_name", "title", "department", "role", "email", "phone", "notes"]:
            if contact_data.get(field) is not None:
                setattr(contact, field, contact_data[field])

        # Handle is_primary specially
        if contact_data.get("is_primary"):
            stmt = select(ContactOrganization.contact_id).where(
                ContactOrganization.organization_id == org_id,
                ContactOrganization.contact_id != contact_id
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


async def delete_organization_contact(org_id: int, contact_id: int):
    """Delete a contact from an organization."""
    async with async_get_session() as session:
        stmt = select(ContactOrganization).where(
            ContactOrganization.contact_id == contact_id,
            ContactOrganization.organization_id == org_id
        )
        result = await session.execute(stmt)
        link = result.scalar_one_or_none()
        if not link:
            raise HTTPException(status_code=404, detail="Contact not found for this organization")

    success = await delete_contact(contact_id)
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
    """Get signals for an organization."""
    org = await find_organization_by_id(org_id)
    if not org:
        raise HTTPException(status_code=404, detail="Organization not found")
    return await find_signals_by_org(org_id, signal_type=signal_type, limit=limit)


async def create_signal(org_id: int, signal_data: Dict[str, Any]):
    """Create a signal for an organization."""
    org = await find_organization_by_id(org_id)
    if not org:
        raise HTTPException(status_code=404, detail="Organization not found")
    return await create_signal_repo(
        org_id=org_id,
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
    signal = await find_signal_by_id(signal_id)
    if not signal or signal.organization_id != org_id:
        raise HTTPException(status_code=404, detail="Signal not found")
    await delete_signal_repo(signal_id)
    return True
