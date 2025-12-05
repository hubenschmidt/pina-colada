"""Repository layer for contact data access."""

import logging
from typing import Any, Dict, List, Optional, Tuple
from sqlalchemy import func, or_, select
from sqlalchemy.orm import selectinload
from lib.db import async_get_session
from models.Contact import Contact, ContactAccount
from models.Individual import Individual
from models.Organization import Organization
from schemas.contact import ContactCreate, ContactUpdate

__all__ = ["ContactCreate", "ContactUpdate"]

logger = logging.getLogger(__name__)


async def find_all_contacts_paginated(
    page: int = 1,
    page_size: int = 50,
    search: Optional[str] = None,
    order_by: str = "updated_at",
    order: str = "DESC",
) -> Tuple[List[Contact], int]:
    """Find contacts with pagination, sorting, and search."""
    async with async_get_session() as session:
        stmt = select(Contact).options(selectinload(Contact.accounts))

        if search and search.strip():
            search_lower = search.strip().lower()
            stmt = stmt.where(
                or_(
                    func.lower(Contact.first_name).contains(search_lower),
                    func.lower(Contact.last_name).contains(search_lower),
                    func.lower(Contact.email).contains(search_lower),
                    func.lower(Contact.title).contains(search_lower),
                )
            )

        count_stmt = select(func.count()).select_from(stmt.subquery())
        count_result = await session.execute(count_stmt)
        total_count = count_result.scalar() or 0

        sort_map = {
            "updated_at": Contact.updated_at,
            "first_name": Contact.first_name,
            "last_name": Contact.last_name,
            "email": Contact.email,
            "phone": Contact.phone,
            "title": Contact.title,
        }
        sort_column = sort_map.get(order_by, Contact.updated_at)
        stmt = stmt.order_by(sort_column.desc() if order.upper() == "DESC" else sort_column.asc())

        offset = (page - 1) * page_size
        stmt = stmt.limit(page_size).offset(offset)

        result = await session.execute(stmt)
        contacts = list(result.scalars().all())

        return contacts, total_count


async def find_contacts_by_account(account_id: int) -> List[Contact]:
    """Find all contacts linked to an account."""
    async with async_get_session() as session:
        stmt = (
            select(Contact)
            .join(ContactAccount, Contact.id == ContactAccount.contact_id)
            .where(ContactAccount.account_id == account_id)
            .options(selectinload(Contact.accounts))
            .order_by(Contact.is_primary.desc(), Contact.id)
        )
        result = await session.execute(stmt)
        return list(result.scalars().all())


async def find_contact_by_id(contact_id: int) -> Optional[Contact]:
    """Find a contact by ID."""
    async with async_get_session() as session:
        stmt = select(Contact).options(selectinload(Contact.accounts)).where(Contact.id == contact_id)
        result = await session.execute(stmt)
        return result.scalar_one_or_none()


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
    account_name = None
    if contact.accounts:
        account_name = contact.accounts[0].name
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
    """Search both Individuals and existing Contacts."""
    async with async_get_session() as session:
        search_pattern = f"%{query}%"

        # Search Contacts (with accounts for account name)
        contact_stmt = (
            select(Contact)
            .options(selectinload(Contact.accounts))
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


async def get_or_create_contact(
    contact_id: Optional[int] = None,
    individual_id: Optional[int] = None,
    organization_id: Optional[int] = None,
    first_name: Optional[str] = None,
    last_name: Optional[str] = None,
    email: Optional[str] = None,
    phone: Optional[str] = None,
    title: Optional[str] = None,
    is_primary: bool = False,
) -> Optional[Contact]:
    """Get or create a contact, linking to account via individual_id or organization_id."""
    async with async_get_session() as session:
        try:
            # Get account_id from individual or organization
            account_id = None
            if individual_id:
                individual = await session.get(Individual, individual_id)
                if individual:
                    account_id = individual.account_id
            elif organization_id:
                org = await session.get(Organization, organization_id)
                if org:
                    account_id = org.account_id

            # If contact_id provided, update existing
            if contact_id:
                contact = await session.get(Contact, contact_id)
                if contact:
                    if first_name:
                        contact.first_name = first_name
                    if last_name:
                        contact.last_name = last_name
                    if email:
                        contact.email = email
                    if phone:
                        contact.phone = phone
                    if title:
                        contact.title = title
                    contact.is_primary = is_primary
                    await session.commit()
                    await session.refresh(contact)
                    return contact

            # Try to find existing contact by email
            if email:
                stmt = select(Contact).where(func.lower(Contact.email) == email.lower())
                result = await session.execute(stmt)
                existing = result.scalar_one_or_none()
                if existing:
                    return existing

            # Create new contact
            contact = Contact(
                first_name=first_name,
                last_name=last_name,
                email=email,
                phone=phone,
                title=title,
                is_primary=is_primary,
            )
            session.add(contact)
            await session.flush()

            # Link to account if we have one
            if account_id:
                link = ContactAccount(
                    contact_id=contact.id,
                    account_id=account_id,
                    is_primary=is_primary,
                )
                session.add(link)

            await session.commit()
            await session.refresh(contact)
            return contact
        except Exception as e:
            await session.rollback()
            logger.error(f"Failed to get or create contact: {e}")
            raise


async def link_contact_to_account(
    contact_id: int,
    account_id: int,
    is_primary: bool = False,
) -> bool:
    """Link a contact to an account via ContactAccount."""
    async with async_get_session() as session:
        try:
            # Check if link already exists
            stmt = select(ContactAccount).where(
                ContactAccount.contact_id == contact_id,
                ContactAccount.account_id == account_id
            )
            result = await session.execute(stmt)
            if result.scalar_one_or_none():
                return True  # Already linked

            link = ContactAccount(
                contact_id=contact_id,
                account_id=account_id,
                is_primary=is_primary,
            )
            session.add(link)
            await session.commit()
            return True
        except Exception as e:
            await session.rollback()
            logger.error(f"Failed to link contact to account: {e}")
            raise


async def find_contact_account_link(contact_id: int, account_id: int) -> bool:
    """Check if a contact is linked to an account."""
    async with async_get_session() as session:
        stmt = select(ContactAccount).where(
            ContactAccount.contact_id == contact_id,
            ContactAccount.account_id == account_id
        )
        result = await session.execute(stmt)
        return result.scalar_one_or_none() is not None


async def find_other_contact_ids_for_account(account_id: int, exclude_contact_id: int) -> List[int]:
    """Find contact IDs linked to an account, excluding a specific contact."""
    async with async_get_session() as session:
        stmt = select(ContactAccount.contact_id).where(
            ContactAccount.account_id == account_id,
            ContactAccount.contact_id != exclude_contact_id
        )
        result = await session.execute(stmt)
        return [row[0] for row in result.fetchall()]


async def clear_primary_contacts_for_account(account_id: int, exclude_contact_id: int) -> None:
    """Clear is_primary flag for all contacts of an account except one."""
    async with async_get_session() as session:
        from sqlalchemy import update as sql_update
        contact_ids = await find_other_contact_ids_for_account(account_id, exclude_contact_id)
        if contact_ids:
            stmt = sql_update(Contact).where(Contact.id.in_(contact_ids)).values(is_primary=False)
            await session.execute(stmt)
            await session.commit()
