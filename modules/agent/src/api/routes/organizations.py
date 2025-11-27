"""Routes for organizations API endpoints."""

from typing import Optional, List
from fastapi import APIRouter, Request, Query, HTTPException
from pydantic import BaseModel, field_validator
from lib.auth import require_auth
from lib.error_logging import log_errors
from lib.validators import validate_phone
from repositories.organization_repository import (
    find_all_organizations,
    find_organization_by_id,
    create_organization,
    update_organization,
    delete_organization,
    search_organizations,
)
from repositories.contact_repository import (
    find_contacts_by_organization,
    create_contact,
    update_contact,
    delete_contact,
    find_contact_by_id,
)
from repositories.individual_repository import find_individual_by_id


class OrganizationCreate(BaseModel):
    name: str
    website: Optional[str] = None
    phone: Optional[str] = None
    employee_count: Optional[int] = None
    description: Optional[str] = None
    account_id: Optional[int] = None

    @field_validator("phone")
    @classmethod
    def validate_phone_format(cls, v):
        return validate_phone(v)


class OrganizationUpdate(BaseModel):
    name: Optional[str] = None
    website: Optional[str] = None
    phone: Optional[str] = None
    employee_count: Optional[int] = None
    description: Optional[str] = None

    @field_validator("phone")
    @classmethod
    def validate_phone_format(cls, v):
        return validate_phone(v)


class OrgContactCreate(BaseModel):
    individual_id: int
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


class OrgContactUpdate(BaseModel):
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


router = APIRouter(prefix="/organizations", tags=["organizations"])


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


def _org_to_dict(org, include_contacts=False, contacts=None):
    result = {
        "id": org.id,
        "name": org.name,
        "website": org.website,
        "phone": org.phone,
        "industries": [ind.name for ind in org.industries] if org.industries else [],
        "employee_count": org.employee_count,
        "description": org.description,
        "created_at": org.created_at.isoformat() if org.created_at else None,
        "updated_at": org.updated_at.isoformat() if org.updated_at else None,
    }
    if include_contacts and contacts is not None:
        result["contacts"] = [_contact_to_dict(c) for c in contacts]
    return result


@router.get("")
@log_errors
@require_auth
async def get_organizations_route(request: Request):
    """Get all organizations for the current tenant."""
    tenant_id = getattr(request.state, "tenant_id", None)
    organizations = await find_all_organizations(tenant_id=tenant_id)
    return [_org_to_dict(org) for org in organizations]


@router.get("/search")
@log_errors
@require_auth
async def search_organizations_route(request: Request, q: Optional[str] = Query(None, min_length=1)):
    """Search organizations by name."""
    if not q:
        return []
    tenant_id = getattr(request.state, "tenant_id", None)
    organizations = await search_organizations(q, tenant_id=tenant_id)
    return [_org_to_dict(org) for org in organizations]


@router.get("/{org_id}")
@log_errors
@require_auth
async def get_organization_route(request: Request, org_id: int):
    """Get a single organization by ID with contacts."""
    org = await find_organization_by_id(org_id)
    if not org:
        raise HTTPException(status_code=404, detail="Organization not found")
    contacts = await find_contacts_by_organization(org_id)
    return _org_to_dict(org, include_contacts=True, contacts=contacts)


@router.post("")
@log_errors
@require_auth
async def create_organization_route(request: Request, data: OrganizationCreate):
    """Create a new organization."""
    org_data = data.model_dump(exclude_none=True)
    org = await create_organization(org_data)
    return _org_to_dict(org)


@router.put("/{org_id}")
@log_errors
@require_auth
async def update_organization_route(request: Request, org_id: int, data: OrganizationUpdate):
    """Update an existing organization."""
    org_data = data.model_dump(exclude_none=True)
    org = await update_organization(org_id, org_data)
    if not org:
        raise HTTPException(status_code=404, detail="Organization not found")
    return _org_to_dict(org)


@router.delete("/{org_id}")
@log_errors
@require_auth
async def delete_organization_route(request: Request, org_id: int):
    """Delete an organization."""
    success = await delete_organization(org_id)
    if not success:
        raise HTTPException(status_code=404, detail="Organization not found")
    return {"success": True}


# Contact management endpoints

@router.get("/{org_id}/contacts")
@log_errors
@require_auth
async def get_organization_contacts_route(request: Request, org_id: int):
    """Get all contacts for an organization."""
    org = await find_organization_by_id(org_id)
    if not org:
        raise HTTPException(status_code=404, detail="Organization not found")
    contacts = await find_contacts_by_organization(org_id)
    return [_contact_to_dict(c) for c in contacts]


@router.post("/{org_id}/contacts")
@log_errors
@require_auth
async def create_organization_contact_route(request: Request, org_id: int, data: OrgContactCreate):
    """Add a contact (existing individual) to an organization."""
    org = await find_organization_by_id(org_id)
    if not org:
        raise HTTPException(status_code=404, detail="Organization not found")

    # Verify individual exists
    individual = await find_individual_by_id(data.individual_id)
    if not individual:
        raise HTTPException(status_code=404, detail="Individual not found")

    contact_data = data.model_dump(exclude_none=True)
    contact_data["organization_id"] = org_id
    contact = await create_contact(contact_data)
    return _contact_to_dict(contact)


@router.put("/{org_id}/contacts/{contact_id}")
@log_errors
@require_auth
async def update_organization_contact_route(request: Request, org_id: int, contact_id: int, data: OrgContactUpdate):
    """Update a contact for an organization."""
    contact = await find_contact_by_id(contact_id)
    if not contact or contact.organization_id != org_id:
        raise HTTPException(status_code=404, detail="Contact not found")

    contact_data = data.model_dump(exclude_none=True)
    updated = await update_contact(contact_id, contact_data)
    return _contact_to_dict(updated)


@router.delete("/{org_id}/contacts/{contact_id}")
@log_errors
@require_auth
async def delete_organization_contact_route(request: Request, org_id: int, contact_id: int):
    """Remove a contact from an organization."""
    contact = await find_contact_by_id(contact_id)
    if not contact or contact.organization_id != org_id:
        raise HTTPException(status_code=404, detail="Contact not found")

    success = await delete_contact(contact_id)
    if not success:
        raise HTTPException(status_code=404, detail="Contact not found")
    return {"success": True}
