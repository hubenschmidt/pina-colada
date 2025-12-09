"""Service layer for contact business logic."""

import logging
from typing import Optional, List, Dict, Any

from fastapi import HTTPException

from repositories.contact_repository import (
    search_contacts_and_individuals,
    delete_contact as delete_contact_repo,
    find_all_contacts_paginated,
    find_contact_by_id,
    create_contact_with_accounts,
    update_contact_with_accounts,
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
    contact = await find_contact_by_id(contact_id)
    if not contact:
        raise HTTPException(status_code=404, detail="Contact not found")
    return contact


async def create_contact(
    contact_data: Dict[str, Any],
    user_id: Optional[int],
    account_ids: Optional[List[int]],
):
    """Create a contact with optional account links."""
    data = {
        "first_name": contact_data.get("first_name"),
        "last_name": contact_data.get("last_name"),
        "email": contact_data.get("email"),
        "phone": contact_data.get("phone"),
        "title": contact_data.get("title"),
        "department": contact_data.get("department"),
        "role": contact_data.get("role"),
        "notes": contact_data.get("notes"),
        "is_primary": contact_data.get("is_primary", False),
        "created_by": user_id,
        "updated_by": user_id,
    }
    return await create_contact_with_accounts(data, account_ids)


async def update_contact(
    contact_id: int,
    contact_data: Dict[str, Any],
    user_id: Optional[int],
    account_ids: Optional[List[int]],
):
    """Update a contact with account links."""
    data = dict(contact_data)
    data["updated_by"] = user_id

    contact = await update_contact_with_accounts(contact_id, data, account_ids)
    if not contact:
        raise HTTPException(status_code=404, detail="Contact not found")
    return contact


async def delete_contact(contact_id: int):
    """Delete a contact."""
    success = await delete_contact_repo(contact_id)
    if not success:
        raise HTTPException(status_code=404, detail="Contact not found")
    return True
