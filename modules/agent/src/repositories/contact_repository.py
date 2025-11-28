"""Repository layer for contact data access."""

import logging
from typing import List, Optional, Dict, Any
from sqlalchemy import select, text
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
    existing = await _find_existing_contact(individual_id, organization_id, first_name, last_name)
    if existing:
        # Update is_primary if it changed
        if existing.is_primary != is_primary:
            await _update_contact_primary(existing.id, is_primary)
            existing.is_primary = is_primary
        return existing

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


async def search_contacts_and_individuals(query: str, tenant_id: Optional[int] = None) -> List[Dict[str, Any]]:
    """
    Search both Individuals and existing Contacts.
    Returns unified results - Contacts first, then Individuals without contacts.
    """
    async with async_get_session() as session:
        search_pattern = f"%{query}%"

        # Search Contacts directly (they have their own first_name/last_name)
        # UNION with Individuals that don't have a Contact yet
        sql = text("""
            SELECT DISTINCT
                NULL::BIGINT as individual_id,
                c.id as contact_id,
                c.first_name,
                c.last_name,
                c.email,
                c.phone,
                'contact' as source
            FROM "Contact" c
            WHERE (
                c.first_name ILIKE :pattern
                OR c.last_name ILIKE :pattern
                OR c.email ILIKE :pattern
                OR CONCAT(c.first_name, ' ', c.last_name) ILIKE :pattern
            )

            UNION

            SELECT DISTINCT
                i.id as individual_id,
                NULL::BIGINT as contact_id,
                i.first_name,
                i.last_name,
                i.email,
                i.phone,
                'individual' as source
            FROM "Individual" i
            WHERE (
                i.first_name ILIKE :pattern
                OR i.last_name ILIKE :pattern
                OR i.email ILIKE :pattern
                OR CONCAT(i.first_name, ' ', i.last_name) ILIKE :pattern
            )

            ORDER BY first_name, last_name
            LIMIT 20
        """)
        result = await session.execute(sql, {"pattern": search_pattern})
        rows = result.fetchall()
        return [
            {
                "individual_id": row.individual_id,
                "contact_id": row.contact_id,
                "first_name": row.first_name,
                "last_name": row.last_name,
                "email": row.email,
                "phone": row.phone,
                "source": row.source,
            }
            for row in rows
        ]
