"""Controller layer for contact routing to services."""

from typing import Optional, List

from fastapi import Request

from lib.decorators import handle_http_exceptions
from serializers.common import to_paged_response
from serializers.contact import contact_to_list_response, contact_to_detail_response
from schemas.contact import ContactCreate, ContactUpdate
from services.contact_service import (
    get_contacts_paginated,
    search_contacts as search_contacts_service,
    get_contact as get_contact_service,
    create_contact as create_contact_service,
    update_contact as update_contact_service,
    delete_contact as delete_contact_service,
)

# Re-export for routes
__all__ = ["ContactCreate", "ContactUpdate"]


@handle_http_exceptions
async def get_contacts(
    page: int,
    limit: int,
    order_by: str,
    order: str,
    search: Optional[str] = None,
) -> dict:
    """Get contacts with pagination."""
    contacts, total = await get_contacts_paginated(page, limit, order_by, order, search)
    items = [contact_to_list_response(c) for c in contacts]
    return to_paged_response(total, page, limit, items)


@handle_http_exceptions
async def search_contacts(request: Request, query: str) -> List[dict]:
    """Search contacts and individuals."""
    tenant_id = request.state.tenant_id
    return await search_contacts_service(query, tenant_id)


@handle_http_exceptions
async def get_contact(contact_id: int) -> dict:
    """Get contact by ID."""
    contact = await get_contact_service(contact_id)
    return contact_to_detail_response(contact)


@handle_http_exceptions
async def create_contact(request: Request, data: ContactCreate) -> dict:
    """Create a new contact."""
    user_id = request.state.user_id
    contact_data = data.model_dump(exclude={"account_ids"})
    contact = await create_contact_service(contact_data, user_id, data.account_ids)
    return contact_to_detail_response(contact)


@handle_http_exceptions
async def update_contact(request: Request, contact_id: int, data: ContactUpdate) -> dict:
    """Update a contact."""
    user_id = request.state.user_id
    contact_data = data.model_dump(exclude={"account_ids"}, exclude_unset=True)
    contact = await update_contact_service(contact_id, contact_data, user_id, data.account_ids)
    return contact_to_detail_response(contact)


@handle_http_exceptions
async def delete_contact(contact_id: int) -> dict:
    """Delete a contact."""
    await delete_contact_service(contact_id)
    return {"success": True}
