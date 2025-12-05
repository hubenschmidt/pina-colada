"""Service layer for contact business logic."""

import logging
from typing import Optional, List, Dict, Any

from fastapi import HTTPException
from sqlalchemy import select, delete as sql_delete

from lib.db import async_get_session
from models.Contact import Contact, ContactAccount
from repositories.contact_repository import (
    search_contacts_and_individuals,
    delete_contact as delete_contact_repo,
    find_all_contacts_paginated,
    ContactCreate,
    ContactUpdate,
)

# Re-export Pydantic models for controllers
__all__ = ["ContactCreate", "ContactUpdate"]

logger = logging.getLogger(__name__)


async def get_contacts_paginated(
    page: int,
    page_size: int,
    order_by: str,
    order: str,
    search: Optional[str] = None,
):
    """Get contacts with pagination."""
    return await find_all_contacts_paginated(
        page=page,
        page_size=page_size,
        search=search,
        order_by=order_by,
        order=order,
    )


async def search_contacts(query: str, tenant_id: Optional[int]):
    """Search contacts and individuals by name or email."""
    return await search_contacts_and_individuals(query, tenant_id=tenant_id)


async def get_contact(contact_id: int):
    """Get contact by ID."""
    async with async_get_session() as session:
        stmt = select(Contact).where(Contact.id == contact_id)
        result = await session.execute(stmt)
        contact = result.scalar_one_or_none()
        if not contact:
            raise HTTPException(status_code=404, detail="Contact not found")
        return contact


async def create_contact(
    contact_data: Dict[str, Any],
    user_id: Optional[int],
    account_ids: Optional[List[int]],
):
    """Create a contact with optional account links."""
    async with async_get_session() as session:
        contact = Contact(
            first_name=contact_data.get("first_name"),
            last_name=contact_data.get("last_name"),
            email=contact_data.get("email"),
            phone=contact_data.get("phone"),
            title=contact_data.get("title"),
            department=contact_data.get("department"),
            role=contact_data.get("role"),
            notes=contact_data.get("notes"),
            is_primary=contact_data.get("is_primary", False),
            created_by=user_id,
            updated_by=user_id,
        )
        session.add(contact)
        await session.flush()

        if account_ids:
            for acc_id in account_ids:
                link = ContactAccount(contact_id=contact.id, account_id=acc_id)
                session.add(link)

        await session.commit()

        stmt = select(Contact).where(Contact.id == contact.id)
        result = await session.execute(stmt)
        return result.scalar_one()


async def update_contact(
    contact_id: int,
    contact_data: Dict[str, Any],
    user_id: Optional[int],
    account_ids: Optional[List[int]],
):
    """Update a contact with account links."""
    async with async_get_session() as session:
        stmt = select(Contact).where(Contact.id == contact_id)
        result = await session.execute(stmt)
        contact = result.scalar_one_or_none()
        if not contact:
            raise HTTPException(status_code=404, detail="Contact not found")

        for field in ["first_name", "last_name", "email", "phone", "title", "department", "role", "notes", "is_primary"]:
            if contact_data.get(field) is not None:
                setattr(contact, field, contact_data[field])

        contact.updated_by = user_id

        if account_ids is not None:
            await session.execute(
                sql_delete(ContactAccount).where(ContactAccount.contact_id == contact_id)
            )
            for acc_id in account_ids:
                link = ContactAccount(contact_id=contact_id, account_id=acc_id)
                session.add(link)

        await session.commit()

        stmt = select(Contact).where(Contact.id == contact_id)
        result = await session.execute(stmt)
        return result.scalar_one()


async def delete_contact(contact_id: int):
    """Delete a contact."""
    success = await delete_contact_repo(contact_id)
    if not success:
        raise HTTPException(status_code=404, detail="Contact not found")
    return True
