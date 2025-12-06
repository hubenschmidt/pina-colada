"""Routes for individuals API endpoints."""

from typing import Optional

from fastapi import APIRouter, Request, Query

from controllers.individual_controller import (
    create_individual,
    create_individual_contact,
    delete_individual,
    delete_individual_contact,
    get_individual,
    get_individual_contacts,
    get_individuals,
    search_individuals,
    update_individual,
    update_individual_contact,
)
from schemas.individual import (
    IndContactCreate as ContactCreate,
    IndContactUpdate as ContactUpdate,
    IndividualCreate,
    IndividualUpdate,
)


router = APIRouter(prefix="/individuals", tags=["individuals"])


# Individual CRUD

@router.get("")
async def get_individuals_route(
    request: Request,
    page: int = Query(1, ge=1),
    limit: int = Query(50, ge=1, le=100),
    order_by: str = Query("updated_at", alias="orderBy"),
    order: str = Query("DESC", regex="^(ASC|DESC)$"),
    search: Optional[str] = Query(None),
):
    """Get individuals with pagination, sorting, and search."""
    return await get_individuals(request, page, limit, order_by, order, search)


@router.get("/search")
async def search_individuals_route(request: Request, q: Optional[str] = Query(None, min_length=1)):
    """Search individuals by name or email."""
    if not q:
        return []
    return await search_individuals(request, q)


@router.get("/{individual_id}")
async def get_individual_route(request: Request, individual_id: int):
    """Get a single individual by ID."""
    return await get_individual(individual_id)


@router.post("")
async def create_individual_route(request: Request, data: IndividualCreate):
    """Create a new individual."""
    return await create_individual(request, data)


@router.put("/{individual_id}")
async def update_individual_route(request: Request, individual_id: int, data: IndividualUpdate):
    """Update an existing individual."""
    return await update_individual(request, individual_id, data)


@router.delete("/{individual_id}")
async def delete_individual_route(request: Request, individual_id: int):
    """Delete an individual."""
    return await delete_individual(individual_id)


# Contact management

@router.get("/{individual_id}/contacts")
async def get_individual_contacts_route(request: Request, individual_id: int):
    """Get all contacts for an individual."""
    return await get_individual_contacts(individual_id)


@router.post("/{individual_id}/contacts")
async def create_individual_contact_route(request: Request, individual_id: int, data: ContactCreate):
    """Add a contact to an individual."""
    return await create_individual_contact(request, individual_id, data)


@router.put("/{individual_id}/contacts/{contact_id}")
async def update_individual_contact_route(request: Request, individual_id: int, contact_id: int, data: ContactUpdate):
    """Update a contact for an individual."""
    return await update_individual_contact(request, individual_id, contact_id, data)


@router.delete("/{individual_id}/contacts/{contact_id}")
async def delete_individual_contact_route(request: Request, individual_id: int, contact_id: int):
    """Remove a contact from an individual."""
    return await delete_individual_contact(individual_id, contact_id)
