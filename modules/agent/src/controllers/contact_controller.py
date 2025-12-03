"""Controller layer for contact routing to services."""

import logging
from typing import Optional, List, Dict, Any

from fastapi import Request

from lib.decorators import handle_http_exceptions
from services.contact_service import (
    ContactCreate,
    ContactUpdate,
    get_contacts_paginated,
    search_contacts as search_contacts_service,
    get_contact as get_contact_service,
    create_contact as create_contact_service,
    update_contact as update_contact_service,
    delete_contact as delete_contact_service,
)

logger = logging.getLogger(__name__)

# Re-export for routes
__all__ = ["ContactCreate", "ContactUpdate"]


def _to_paged_response(count: int, page: int, limit: int, items: List) -> dict:
    """Convert to paged response format."""
    return {
        "items": items,
        "currentPage": page,
        "totalPages": max(1, (count + limit - 1) // limit),
        "total": count,
        "pageSize": limit,
    }


def _contact_to_list_response(contact) -> dict:
    """Convert Contact model to dict - optimized for list/table view."""
    organizations = [
        {"id": org.id, "name": org.name}
        for org in (contact.organizations or [])
    ]

    return {
        "id": contact.id,
        "first_name": contact.first_name or "",
        "last_name": contact.last_name or "",
        "title": contact.title,
        "email": contact.email,
        "phone": contact.phone,
        "organizations": organizations,
        "updated_at": contact.updated_at.isoformat() if contact.updated_at else None,
    }


def _contact_to_detail_response(contact) -> dict:
    """Convert Contact model to dict with linked individuals and organizations."""
    first_name = contact.first_name
    last_name = contact.last_name

    if not first_name and contact.individuals:
        first_name = contact.individuals[0].first_name
    if not last_name and contact.individuals:
        last_name = contact.individuals[0].last_name

    individuals = [
        {"id": ind.id, "first_name": ind.first_name, "last_name": ind.last_name}
        for ind in (contact.individuals or [])
    ]

    organizations = [
        {"id": org.id, "name": org.name}
        for org in (contact.organizations or [])
    ]

    return {
        "id": contact.id,
        "first_name": first_name or "",
        "last_name": last_name or "",
        "title": contact.title,
        "department": contact.department,
        "role": contact.role,
        "email": contact.email,
        "phone": contact.phone,
        "is_primary": contact.is_primary,
        "notes": contact.notes,
        "individuals": individuals,
        "organizations": organizations,
        "created_at": contact.created_at.isoformat() if contact.created_at else None,
        "updated_at": contact.updated_at.isoformat() if contact.updated_at else None,
    }


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
    items = [_contact_to_list_response(c) for c in contacts]
    return _to_paged_response(total, page, limit, items)


@handle_http_exceptions
async def search_contacts(request: Request, query: str) -> List[dict]:
    """Search contacts and individuals."""
    tenant_id = request.state.tenant_id
    return await search_contacts_service(query, tenant_id)


@handle_http_exceptions
async def get_contact(contact_id: int) -> dict:
    """Get contact by ID."""
    contact = await get_contact_service(contact_id)
    return _contact_to_detail_response(contact)


@handle_http_exceptions
async def create_contact(request: Request, data: ContactCreate) -> dict:
    """Create a new contact."""
    user_id = request.state.user_id
    contact_data = data.model_dump(exclude={"individual_ids", "organization_ids"})
    contact = await create_contact_service(contact_data, user_id, data.individual_ids, data.organization_ids)
    return _contact_to_detail_response(contact)


@handle_http_exceptions
async def update_contact(request: Request, contact_id: int, data: ContactUpdate) -> dict:
    """Update a contact."""
    user_id = request.state.user_id
    contact_data = data.model_dump(exclude={"individual_ids", "organization_ids"}, exclude_unset=True)
    contact = await update_contact_service(contact_id, contact_data, user_id, data.individual_ids, data.organization_ids)
    return _contact_to_detail_response(contact)


@handle_http_exceptions
async def delete_contact(contact_id: int) -> dict:
    """Delete a contact."""
    await delete_contact_service(contact_id)
    return {"success": True}
