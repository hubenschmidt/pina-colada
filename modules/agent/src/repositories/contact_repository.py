"""Repository layer for contact data access."""

import logging
from typing import List, Optional, Dict, Any
from sqlalchemy import select
from sqlalchemy.orm import selectinload
from models.Contact import Contact
from models.Individual import Individual
from lib.db import async_get_session

logger = logging.getLogger(__name__)


async def _find_contact_by_individual_only(session, individual_id: int) -> Optional[Contact]:
    """Find contact linked to individual without organization filter."""
    from models.Contact import ContactIndividual

    stmt = (
        select(Contact)
        .join(ContactIndividual, Contact.id == ContactIndividual.contact_id)
        .where(ContactIndividual.individual_id == individual_id)
    )
    result = await session.execute(stmt)
    return result.scalar_one_or_none()


async def _find_contact_by_individual_and_org(session, individual_id: int, organization_id: int) -> Optional[Contact]:
    """Find contact linked to both individual and organization."""
    from models.Contact import ContactIndividual, ContactOrganization

    stmt = (
        select(Contact)
        .join(ContactIndividual, Contact.id == ContactIndividual.contact_id)
        .where(ContactIndividual.individual_id == individual_id)
        .join(ContactOrganization, Contact.id == ContactOrganization.contact_id)
        .where(ContactOrganization.organization_id == organization_id)
    )
    result = await session.execute(stmt)
    return result.scalar_one_or_none()


async def find_contact_by_individual_and_org(
    individual_id: int, organization_id: Optional[int] = None
) -> Optional[Contact]:
    """Find contact by individual and organization via junction tables."""
    async with async_get_session() as session:
        if organization_id is None:
            return await _find_contact_by_individual_only(session, individual_id)
        return await _find_contact_by_individual_and_org(session, individual_id, organization_id)


async def create_contact(data: Dict[str, Any]) -> Contact:
    """Create a new contact record."""
    async with async_get_session() as session:
        try:
            contact = Contact(**data)
            session.add(contact)
            await session.commit()
            await session.refresh(contact)
            return contact
        except Exception as e:
            await session.rollback()
            logger.error(f"Failed to create contact: {e}")
            raise


def _maybe_add_org_link(session, contact_id: int, organization_id: Optional[int], is_primary: bool) -> None:
    """Add organization link if organization_id is provided."""
    from models.Contact import ContactOrganization

    if organization_id is None:
        return
    org_link = ContactOrganization(
        contact_id=contact_id,
        organization_id=organization_id,
        is_primary=is_primary
    )
    session.add(org_link)


def _maybe_add_individual_link(session, contact_id: int, individual_id: Optional[int]) -> None:
    """Add individual link if individual_id is provided."""
    from models.Contact import ContactIndividual

    if individual_id is None:
        return
    ind_link = ContactIndividual(contact_id=contact_id, individual_id=individual_id)
    session.add(ind_link)


async def find_contact_by_org_and_name(
    organization_id: int, first_name: str, last_name: str
) -> Optional[Contact]:
    """Find contact by organization and name."""
    from models.Contact import ContactOrganization

    async with async_get_session() as session:
        stmt = (
            select(Contact)
            .join(ContactOrganization, Contact.id == ContactOrganization.contact_id)
            .where(ContactOrganization.organization_id == organization_id)
            .where(Contact.first_name == first_name)
            .where(Contact.last_name == last_name)
        )
        result = await session.execute(stmt)
        return result.scalar_one_or_none()


async def _find_existing_contact(
    individual_id: Optional[int],
    organization_id: Optional[int],
    first_name: Optional[str],
    last_name: Optional[str]
) -> Optional[Contact]:
    """Find existing contact by individual or by org+name."""
    if individual_id is not None:
        return await find_contact_by_individual_and_org(individual_id, organization_id)
    if organization_id is not None and first_name and last_name:
        return await find_contact_by_org_and_name(organization_id, first_name, last_name)
    return None


async def _update_contact_primary(contact_id: int, is_primary: bool) -> None:
    """Update contact's is_primary field."""
    async with async_get_session() as session:
        stmt = select(Contact).where(Contact.id == contact_id)
        result = await session.execute(stmt)
        contact = result.scalar_one_or_none()
        if not contact:
            return
        contact.is_primary = is_primary
        await session.commit()


async def get_or_create_contact(
    contact_id: Optional[int] = None,
    individual_id: Optional[int] = None,
    organization_id: Optional[int] = None,
    first_name: Optional[str] = None,
    last_name: Optional[str] = None,
    email: Optional[str] = None,
    phone: Optional[str] = None,
    title: Optional[str] = None,
    is_primary: bool = False
) -> Contact:
    """Get or create a contact, optionally linked to individual and/or organization."""

    async def _sync_primary_status(contact: Contact) -> Contact:
        if contact.is_primary != is_primary:
            await _update_contact_primary(contact.id, is_primary)
            contact.is_primary = is_primary
        return contact

    # If contact_id is provided, find by ID first
    if contact_id is not None:
        existing = await find_contact_by_id(contact_id)
        if existing:
            return await _sync_primary_status(existing)

    # Otherwise try to find by individual/org/name
    existing = await _find_existing_contact(individual_id, organization_id, first_name, last_name)
    if existing:
        return await _sync_primary_status(existing)

    async with async_get_session() as session:
        try:
            contact = Contact(
                first_name=first_name,
                last_name=last_name,
                email=email,
                phone=phone,
                title=title,
                is_primary=is_primary
            )
            session.add(contact)
            await session.flush()

            _maybe_add_individual_link(session, contact.id, individual_id)
            _maybe_add_org_link(session, contact.id, organization_id, is_primary)

            await session.commit()
            await session.refresh(contact)
            return contact
        except Exception as e:
            await session.rollback()
            logger.error(f"Failed to create contact: {e}")
            raise


async def find_contacts_by_individual(individual_id: int) -> List[Contact]:
    """Find all contacts linked to an individual."""
    from models.Contact import ContactIndividual
    async with async_get_session() as session:
        stmt = (
            select(Contact)
            .join(ContactIndividual, Contact.id == ContactIndividual.contact_id)
            .where(ContactIndividual.individual_id == individual_id)
            .options(selectinload(Contact.organizations), selectinload(Contact.individuals))
            .order_by(Contact.is_primary.desc(), Contact.id)
        )
        result = await session.execute(stmt)
        return list(result.scalars().all())


async def find_contacts_by_organization(organization_id: int) -> List[Contact]:
    """Find all contacts linked to an organization."""
    from models.Contact import ContactOrganization
    async with async_get_session() as session:
        stmt = (
            select(Contact)
            .join(ContactOrganization, Contact.id == ContactOrganization.contact_id)
            .where(ContactOrganization.organization_id == organization_id)
            .options(selectinload(Contact.organizations), selectinload(Contact.individuals))
            .order_by(Contact.is_primary.desc(), Contact.id)
        )
        result = await session.execute(stmt)
        return list(result.scalars().all())


async def find_contact_by_id(contact_id: int) -> Optional[Contact]:
    """Find a contact by ID."""
    async with async_get_session() as session:
        return await session.get(Contact, contact_id)


async def update_contact(contact_id: int, data: Dict[str, Any]) -> Optional[Contact]:
    """Update an existing contact."""
    async with async_get_session() as session:
        try:
            contact = await session.get(Contact, contact_id)
            if not contact:
                return None

            for key, value in data.items():
                if hasattr(contact, key):
                    setattr(contact, key, value)

            await session.commit()
            await session.refresh(contact)
            return contact
        except Exception as e:
            await session.rollback()
            logger.error(f"Failed to update contact: {e}")
            raise


async def delete_contact(contact_id: int) -> bool:
    """Delete a contact by ID."""
    async with async_get_session() as session:
        try:
            contact = await session.get(Contact, contact_id)
            if not contact:
                return False
            await session.delete(contact)
            await session.commit()
            return True
        except Exception as e:
            await session.rollback()
            logger.error(f"Failed to delete contact: {e}")
            raise


def _contact_to_search_result(contact: Contact) -> Dict[str, Any]:
    """Convert Contact to search result dict."""
    # Get account name from linked organizations
    account_name = None
    if contact.organizations:
        account_name = contact.organizations[0].name
    return {
        "individual_id": None,
        "contact_id": contact.id,
        "first_name": contact.first_name,
        "last_name": contact.last_name,
        "email": contact.email,
        "phone": contact.phone,
        "account_name": account_name,
        "source": "contact",
    }


def _individual_to_search_result(individual: Individual) -> Dict[str, Any]:
    """Convert Individual to search result dict."""
    account_name = individual.account.name if individual.account else None
    return {
        "individual_id": individual.id,
        "contact_id": None,
        "first_name": individual.first_name,
        "last_name": individual.last_name,
        "email": individual.email,
        "phone": individual.phone,
        "account_name": account_name,
        "source": "individual",
    }


async def search_contacts_and_individuals(query: str, tenant_id: Optional[int] = None) -> List[Dict[str, Any]]:
    """
    Search both Individuals and existing Contacts.
    Returns unified results - Contacts first, then Individuals.
    """
    from sqlalchemy import or_, func

    async with async_get_session() as session:
        search_pattern = f"%{query}%"

        # Search Contacts (with organizations for account name)
        contact_stmt = (
            select(Contact)
            .options(selectinload(Contact.organizations))
            .where(
                or_(
                    Contact.first_name.ilike(search_pattern),
                    Contact.last_name.ilike(search_pattern),
                    Contact.email.ilike(search_pattern),
                    func.concat(Contact.first_name, " ", Contact.last_name).ilike(search_pattern),
                )
            )
            .order_by(Contact.first_name, Contact.last_name)
            .limit(10)
        )
        contact_result = await session.execute(contact_stmt)
        contacts = list(contact_result.scalars().all())

        # Search Individuals (with account for account name)
        individual_stmt = (
            select(Individual)
            .options(selectinload(Individual.account))
            .where(
                or_(
                    Individual.first_name.ilike(search_pattern),
                    Individual.last_name.ilike(search_pattern),
                    Individual.email.ilike(search_pattern),
                    func.concat(Individual.first_name, " ", Individual.last_name).ilike(search_pattern),
                )
            )
            .order_by(Individual.first_name, Individual.last_name)
            .limit(10)
        )
        individual_result = await session.execute(individual_stmt)
        individuals = list(individual_result.scalars().all())

        # Combine results: contacts first, then individuals
        results = [_contact_to_search_result(c) for c in contacts]
        results.extend(_individual_to_search_result(i) for i in individuals)
        return results[:20]
