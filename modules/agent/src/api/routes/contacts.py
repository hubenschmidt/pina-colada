"""Contact API routes."""

from typing import Optional

from fastapi import APIRouter, Request, Query

from controllers.contact_controller import (
    get_contacts,
    search_contacts,
    get_contact,
    create_contact,
    update_contact,
    delete_contact,
    ContactCreate,
    ContactUpdate,
)
from lib.auth import require_auth
from lib.error_logging import log_errors


router = APIRouter(prefix="/contacts", tags=["contacts"])


@router.get("")
@log_errors
@require_auth
async def get_contacts_route(
    request: Request,
    page: int = Query(1, ge=1),
    limit: int = Query(50, ge=1, le=100),
    order_by: str = Query("updated_at", alias="orderBy"),
    order: str = Query("DESC", regex="^(ASC|DESC)$"),
    search: Optional[str] = Query(None),
):
    """Get contacts with pagination, sorting, and search."""
    return await get_contacts(page, limit, order_by, order, search)


@router.get("/search")
@log_errors
@require_auth
async def search_contacts_route(request: Request, q: Optional[str] = Query(None, min_length=1)):
    """Search for contacts/individuals by name or email."""
    if not q:
        return []
    return await search_contacts(request, q)


@router.get("/{contact_id}")
@log_errors
@require_auth
async def get_contact_route(request: Request, contact_id: int):
    """Get a single contact by ID."""
    return await get_contact(contact_id)


@router.post("")
@log_errors
@require_auth
async def create_contact_route(request: Request, data: ContactCreate):
    """Create a new contact with optional individual/organization links."""
    return await create_contact(request, data)


@router.put("/{contact_id}")
@log_errors
@require_auth
async def update_contact_route(request: Request, contact_id: int, data: ContactUpdate):
    """Update an existing contact."""
    return await update_contact(request, contact_id, data)


@router.delete("/{contact_id}")
@log_errors
@require_auth
async def delete_contact_route(request: Request, contact_id: int):
    """Delete a contact."""
    return await delete_contact(contact_id)
