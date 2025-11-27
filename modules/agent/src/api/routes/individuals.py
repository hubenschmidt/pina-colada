"""Routes for individuals API endpoints."""

from typing import Optional, List
from fastapi import APIRouter, Request, Query, HTTPException
from pydantic import BaseModel, field_validator
from lib.auth import require_auth
from lib.error_logging import log_errors
from lib.validators import validate_phone
from repositories.individual_repository import (
    find_all_individuals,
    find_individual_by_id,
    create_individual,
    update_individual,
    delete_individual,
    search_individuals,
)
from repositories.contact_repository import (
    find_contacts_by_individual,
    create_contact,
    update_contact,
    delete_contact,
    find_contact_by_id,
)


class IndividualCreate(BaseModel):
    first_name: str
    last_name: str
    email: Optional[str] = None
    phone: Optional[str] = None
    linkedin_url: Optional[str] = None
    title: Optional[str] = None
    notes: Optional[str] = None
    account_id: Optional[int] = None

    @field_validator("phone")
    @classmethod
    def validate_phone_format(cls, v):
        return validate_phone(v)


class IndividualUpdate(BaseModel):
    first_name: Optional[str] = None
    last_name: Optional[str] = None
    email: Optional[str] = None
    phone: Optional[str] = None
    linkedin_url: Optional[str] = None
    title: Optional[str] = None
    notes: Optional[str] = None

    @field_validator("phone")
    @classmethod
    def validate_phone_format(cls, v):
        return validate_phone(v)


class ContactCreate(BaseModel):
    organization_id: Optional[int] = None
    title: Optional[str] = None
    department: Optional[str] = None
    role: Optional[str] = None
    email: Optional[str] = None
    phone: Optional[str] = None
    is_primary: bool = False
    notes: Optional[str] = None

    @field_validator("phone")
    @classmethod
    def validate_phone_format(cls, v):
        return validate_phone(v)


class ContactUpdate(BaseModel):
    organization_id: Optional[int] = None
    title: Optional[str] = None
    department: Optional[str] = None
    role: Optional[str] = None
    email: Optional[str] = None
    phone: Optional[str] = None
    is_primary: Optional[bool] = None
    notes: Optional[str] = None

    @field_validator("phone")
    @classmethod
    def validate_phone_format(cls, v):
        return validate_phone(v)


router = APIRouter(prefix="/individuals", tags=["individuals"])


def _contact_to_dict(contact):
    return {
        "id": contact.id,
        "individual_id": contact.individual_id,
        "organization_id": contact.organization_id,
        "title": contact.title,
        "department": contact.department,
        "role": contact.role,
        "email": contact.email,
        "phone": contact.phone,
        "is_primary": contact.is_primary,
        "notes": contact.notes,
        "created_at": contact.created_at.isoformat() if contact.created_at else None,
        "updated_at": contact.updated_at.isoformat() if contact.updated_at else None,
    }


def _ind_to_dict(ind, include_contacts=False, contacts=None):
    result = {
        "id": ind.id,
        "first_name": ind.first_name,
        "last_name": ind.last_name,
        "email": ind.email,
        "phone": ind.phone,
        "linkedin_url": ind.linkedin_url,
        "title": ind.title,
        "notes": ind.notes,
        "created_at": ind.created_at.isoformat() if ind.created_at else None,
        "updated_at": ind.updated_at.isoformat() if ind.updated_at else None,
    }
    if include_contacts and contacts is not None:
        result["contacts"] = [_contact_to_dict(c) for c in contacts]
    return result


@router.get("")
@log_errors
@require_auth
async def get_individuals_route(request: Request):
    """Get all individuals for the current tenant."""
    tenant_id = getattr(request.state, "tenant_id", None)
    individuals = await find_all_individuals(tenant_id=tenant_id)
    return [_ind_to_dict(ind) for ind in individuals]


@router.get("/search")
@log_errors
@require_auth
async def search_individuals_route(request: Request, q: Optional[str] = Query(None, min_length=1)):
    """Search individuals by name or email."""
    if not q:
        return []
    tenant_id = getattr(request.state, "tenant_id", None)
    individuals = await search_individuals(q, tenant_id=tenant_id)
    return [_ind_to_dict(ind) for ind in individuals]


@router.get("/{individual_id}")
@log_errors
@require_auth
async def get_individual_route(request: Request, individual_id: int):
    """Get a single individual by ID with their contacts."""
    individual = await find_individual_by_id(individual_id)
    if not individual:
        raise HTTPException(status_code=404, detail="Individual not found")
    contacts = await find_contacts_by_individual(individual_id)
    return _ind_to_dict(individual, include_contacts=True, contacts=contacts)


@router.post("")
@log_errors
@require_auth
async def create_individual_route(request: Request, data: IndividualCreate):
    """Create a new individual and auto-create them as primary contact."""
    ind_data = data.model_dump(exclude_none=True)
    individual = await create_individual(ind_data)

    # Auto-create the individual as their own primary contact
    await create_contact({
        "individual_id": individual.id,
        "email": individual.email,
        "phone": individual.phone,
        "title": individual.title,
        "is_primary": True,
    })

    contacts = await find_contacts_by_individual(individual.id)
    return _ind_to_dict(individual, include_contacts=True, contacts=contacts)


@router.put("/{individual_id}")
@log_errors
@require_auth
async def update_individual_route(request: Request, individual_id: int, data: IndividualUpdate):
    """Update an existing individual."""
    ind_data = data.model_dump(exclude_none=True)
    individual = await update_individual(individual_id, ind_data)
    if not individual:
        raise HTTPException(status_code=404, detail="Individual not found")
    return _ind_to_dict(individual)


@router.delete("/{individual_id}")
@log_errors
@require_auth
async def delete_individual_route(request: Request, individual_id: int):
    """Delete an individual."""
    success = await delete_individual(individual_id)
    if not success:
        raise HTTPException(status_code=404, detail="Individual not found")
    return {"success": True}


# Contact management endpoints

@router.get("/{individual_id}/contacts")
@log_errors
@require_auth
async def get_individual_contacts_route(request: Request, individual_id: int):
    """Get all contacts for an individual."""
    individual = await find_individual_by_id(individual_id)
    if not individual:
        raise HTTPException(status_code=404, detail="Individual not found")
    contacts = await find_contacts_by_individual(individual_id)
    return [_contact_to_dict(c) for c in contacts]


@router.post("/{individual_id}/contacts")
@log_errors
@require_auth
async def create_individual_contact_route(request: Request, individual_id: int, data: ContactCreate):
    """Create a new contact for an individual."""
    individual = await find_individual_by_id(individual_id)
    if not individual:
        raise HTTPException(status_code=404, detail="Individual not found")

    contact_data = data.model_dump(exclude_none=True)
    contact_data["individual_id"] = individual_id
    contact = await create_contact(contact_data)
    return _contact_to_dict(contact)


@router.put("/{individual_id}/contacts/{contact_id}")
@log_errors
@require_auth
async def update_individual_contact_route(request: Request, individual_id: int, contact_id: int, data: ContactUpdate):
    """Update a contact for an individual."""
    contact = await find_contact_by_id(contact_id)
    if not contact or contact.individual_id != individual_id:
        raise HTTPException(status_code=404, detail="Contact not found")

    contact_data = data.model_dump(exclude_none=True)
    updated = await update_contact(contact_id, contact_data)
    return _contact_to_dict(updated)


@router.delete("/{individual_id}/contacts/{contact_id}")
@log_errors
@require_auth
async def delete_individual_contact_route(request: Request, individual_id: int, contact_id: int):
    """Delete a contact for an individual."""
    contact = await find_contact_by_id(contact_id)
    if not contact or contact.individual_id != individual_id:
        raise HTTPException(status_code=404, detail="Contact not found")

    success = await delete_contact(contact_id)
    if not success:
        raise HTTPException(status_code=404, detail="Contact not found")
    return {"success": True}
