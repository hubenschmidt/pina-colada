"""Contact API routes."""

from typing import Optional, List
from fastapi import APIRouter, Request, Query, HTTPException
from pydantic import BaseModel
from sqlalchemy import select, delete as sql_delete
from sqlalchemy.orm import selectinload

from lib.error_logging import log_errors
from lib.auth import require_auth
from lib.validators import validate_phone
from lib.db import async_get_session
from repositories.contact_repository import (
    search_contacts_and_individuals,
    delete_contact,
    find_all_contacts_paginated,
)
from models.Contact import Contact, ContactIndividual, ContactOrganization
from models.Individual import Individual
from models.Organization import Organization


class ContactCreate(BaseModel):
    first_name: str
    last_name: str
    email: Optional[str] = None
    phone: Optional[str] = None
    title: Optional[str] = None
    department: Optional[str] = None
    role: Optional[str] = None
    notes: Optional[str] = None
    individual_ids: Optional[List[int]] = None
    organization_ids: Optional[List[int]] = None
    is_primary: bool = False

    @classmethod
    def validate_phone_format(cls, v):
        return validate_phone(v) if v else v


class ContactUpdate(BaseModel):
    first_name: Optional[str] = None
    last_name: Optional[str] = None
    email: Optional[str] = None
    phone: Optional[str] = None
    title: Optional[str] = None
    department: Optional[str] = None
    role: Optional[str] = None
    notes: Optional[str] = None
    individual_ids: Optional[List[int]] = None
    organization_ids: Optional[List[int]] = None
    is_primary: Optional[bool] = None

    @classmethod
    def validate_phone_format(cls, v):
        return validate_phone(v) if v else v


router = APIRouter(prefix="/contacts", tags=["contacts"])


def _contact_to_dict(contact: Contact) -> dict:
    """Convert Contact model to dict with linked individuals and organizations."""
    first_name = contact.first_name
    last_name = contact.last_name

    # Fallback to first linked Individual if contact doesn't have name
    if not first_name and contact.individuals:
        first_name = contact.individuals[0].first_name
    if not last_name and contact.individuals:
        last_name = contact.individuals[0].last_name

    # Build individuals list
    individuals = [
        {"id": ind.id, "first_name": ind.first_name, "last_name": ind.last_name}
        for ind in (contact.individuals or [])
    ]

    # Build organizations list
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


def _contact_to_list_dict(contact: Contact) -> dict:
    """Convert Contact model to dict - optimized for list/table view.

    Only returns fields needed for table columns:
    Last Name, First Name, Title, Account, Email, Phone
    """
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
    contacts, total = await find_all_contacts_paginated(
        page=page,
        page_size=limit,
        search=search,
        order_by=order_by,
        order=order,
    )

    items = [_contact_to_list_dict(c) for c in contacts]
    total_pages = (total + limit - 1) // limit if limit > 0 else 1

    return {
        "items": items,
        "currentPage": page,
        "totalPages": total_pages,
        "total": total,
        "pageSize": limit,
    }


@router.get("/search")
@log_errors
@require_auth
async def search_contacts_route(request: Request, q: Optional[str] = Query(None, min_length=1)):
    """
    Search for contacts/individuals by name or email.
    Returns individuals that can be linked as contacts.
    """
    if not q:
        return []
    tenant_id = getattr(request.state, "tenant_id", None)
    results = await search_contacts_and_individuals(q, tenant_id=tenant_id)
    return results


@router.get("/{contact_id}")
@log_errors
@require_auth
async def get_contact_route(request: Request, contact_id: int):
    """Get a single contact by ID."""
    async with async_get_session() as session:
        stmt = select(Contact).where(Contact.id == contact_id)
        result = await session.execute(stmt)
        contact = result.scalar_one_or_none()
        if not contact:
            raise HTTPException(status_code=404, detail="Contact not found")
        return _contact_to_dict(contact)


@router.post("")
@log_errors
@require_auth
async def create_contact_route(request: Request, data: ContactCreate):
    """Create a new contact with optional individual/organization links."""
    user_id = getattr(request.state, "user_id", None)
    async with async_get_session() as session:
        # Create the contact
        contact = Contact(
            first_name=data.first_name,
            last_name=data.last_name,
            email=data.email,
            phone=data.phone,
            title=data.title,
            department=data.department,
            role=data.role,
            notes=data.notes,
            is_primary=data.is_primary,
            created_by=user_id,
            updated_by=user_id,
        )
        session.add(contact)
        await session.flush()

        # Link to individuals
        if data.individual_ids:
            for ind_id in data.individual_ids:
                link = ContactIndividual(contact_id=contact.id, individual_id=ind_id)
                session.add(link)

        # Link to organizations
        if data.organization_ids:
            for org_id in data.organization_ids:
                link = ContactOrganization(contact_id=contact.id, organization_id=org_id)
                session.add(link)

        await session.commit()

        # Re-fetch to get relationships
        stmt = select(Contact).where(Contact.id == contact.id)
        result = await session.execute(stmt)
        contact = result.scalar_one()
        return _contact_to_dict(contact)


@router.put("/{contact_id}")
@log_errors
@require_auth
async def update_contact_route(request: Request, contact_id: int, data: ContactUpdate):
    """Update an existing contact."""
    user_id = getattr(request.state, "user_id", None)
    async with async_get_session() as session:
        stmt = select(Contact).where(Contact.id == contact_id)
        result = await session.execute(stmt)
        contact = result.scalar_one_or_none()
        if not contact:
            raise HTTPException(status_code=404, detail="Contact not found")

        # Update basic fields
        if data.first_name is not None:
            contact.first_name = data.first_name
        if data.last_name is not None:
            contact.last_name = data.last_name
        if data.email is not None:
            contact.email = data.email
        if data.phone is not None:
            contact.phone = data.phone
        if data.title is not None:
            contact.title = data.title
        if data.department is not None:
            contact.department = data.department
        if data.role is not None:
            contact.role = data.role
        if data.notes is not None:
            contact.notes = data.notes
        if data.is_primary is not None:
            contact.is_primary = data.is_primary
        contact.updated_by = user_id

        # Update individual links if provided
        if data.individual_ids is not None:
            await session.execute(
                sql_delete(ContactIndividual).where(ContactIndividual.contact_id == contact_id)
            )
            for ind_id in data.individual_ids:
                link = ContactIndividual(contact_id=contact_id, individual_id=ind_id)
                session.add(link)

        # Update organization links if provided
        if data.organization_ids is not None:
            await session.execute(
                sql_delete(ContactOrganization).where(ContactOrganization.contact_id == contact_id)
            )
            for org_id in data.organization_ids:
                link = ContactOrganization(contact_id=contact_id, organization_id=org_id)
                session.add(link)

        await session.commit()

        # Re-fetch to get updated relationships
        stmt = select(Contact).where(Contact.id == contact_id)
        result = await session.execute(stmt)
        contact = result.scalar_one()
        return _contact_to_dict(contact)


@router.delete("/{contact_id}")
@log_errors
@require_auth
async def delete_contact_route(request: Request, contact_id: int):
    """Delete a contact."""
    success = await delete_contact(contact_id)
    if not success:
        raise HTTPException(status_code=404, detail="Contact not found")
    return {"success": True}
