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
    individual_id: Optional[int] = None
    first_name: Optional[str] = None
    last_name: Optional[str] = None
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
    first_name: Optional[str] = None
    last_name: Optional[str] = None
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
    # Use contact's own name fields, fallback to first linked Individual
    first_name = contact.first_name
    last_name = contact.last_name
    if not first_name and contact.individuals:
        first_name = contact.individuals[0].first_name
    if not last_name and contact.individuals:
        last_name = contact.individuals[0].last_name

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
        "individuals": [{"id": i.id, "first_name": i.first_name, "last_name": i.last_name} for i in (contact.individuals or [])],
        "organizations": [{"id": o.id, "name": o.name} for o in (contact.organizations or [])],
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
    """Create a new organization with an associated Account."""
    from lib.db import async_get_session
    from models.Account import Account

    tenant_id = getattr(request.state, "tenant_id", None)
    org_data = data.model_dump(exclude_none=True)

    # Create an Account for the Organization if not provided
    if not org_data.get("account_id"):
        async with async_get_session() as session:
            account = Account(tenant_id=tenant_id, name=org_data.get("name", ""))
            session.add(account)
            await session.commit()
            await session.refresh(account)
            org_data["account_id"] = account.id

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
    """Add a contact to an organization. Can link to an existing individual or be standalone."""
    from lib.db import async_get_session
    from sqlalchemy import select
    from models.Contact import Contact, ContactOrganization, ContactIndividual

    org = await find_organization_by_id(org_id)
    if not org:
        raise HTTPException(status_code=404, detail="Organization not found")

    # Verify individual exists if provided
    if data.individual_id:
        individual = await find_individual_by_id(data.individual_id)
        if not individual:
            raise HTTPException(status_code=404, detail="Individual not found")

    async with async_get_session() as session:
        # Create the contact
        contact = Contact(
            first_name=data.first_name,
            last_name=data.last_name,
            title=data.title,
            department=data.department,
            role=data.role,
            email=data.email,
            phone=data.phone,
            is_primary=data.is_primary,
            notes=data.notes,
        )
        session.add(contact)
        await session.flush()

        # Link to organization
        org_link = ContactOrganization(contact_id=contact.id, organization_id=org_id, is_primary=data.is_primary)
        session.add(org_link)

        # Link to individual if provided
        if data.individual_id:
            ind_link = ContactIndividual(contact_id=contact.id, individual_id=data.individual_id)
            session.add(ind_link)

        await session.commit()

        # Re-fetch to get relationships
        stmt = select(Contact).where(Contact.id == contact.id)
        result = await session.execute(stmt)
        contact = result.scalar_one()
        return _contact_to_dict(contact)


@router.put("/{org_id}/contacts/{contact_id}")
@log_errors
@require_auth
async def update_organization_contact_route(request: Request, org_id: int, contact_id: int, data: OrgContactUpdate):
    """Update a contact for an organization."""
    from lib.db import async_get_session
    from sqlalchemy import select
    from models.Contact import Contact, ContactOrganization

    async with async_get_session() as session:
        # Verify contact exists and is linked to this organization
        stmt = select(ContactOrganization).where(
            ContactOrganization.contact_id == contact_id,
            ContactOrganization.organization_id == org_id
        )
        result = await session.execute(stmt)
        link = result.scalar_one_or_none()
        if not link:
            raise HTTPException(status_code=404, detail="Contact not found for this organization")

        # Get and update the contact
        stmt = select(Contact).where(Contact.id == contact_id)
        result = await session.execute(stmt)
        contact = result.scalar_one()

        if data.first_name is not None:
            contact.first_name = data.first_name
        if data.last_name is not None:
            contact.last_name = data.last_name
        if data.title is not None:
            contact.title = data.title
        if data.department is not None:
            contact.department = data.department
        if data.role is not None:
            contact.role = data.role
        if data.email is not None:
            contact.email = data.email
        if data.phone is not None:
            contact.phone = data.phone
        if data.is_primary:
            # Unset all other contacts for this org first
            from sqlalchemy import update
            stmt = select(ContactOrganization.contact_id).where(
                ContactOrganization.organization_id == org_id,
                ContactOrganization.contact_id != contact_id
            )
            result = await session.execute(stmt)
            other_contact_ids = [row[0] for row in result.fetchall()]
            if other_contact_ids:
                await session.execute(
                    update(Contact).where(Contact.id.in_(other_contact_ids)).values(is_primary=False)
                )
            contact.is_primary = True
        if data.is_primary is not None:
            contact.is_primary = data.is_primary
        if data.notes is not None:
            contact.notes = data.notes

        await session.commit()

        # Re-fetch
        stmt = select(Contact).where(Contact.id == contact_id)
        result = await session.execute(stmt)
        contact = result.scalar_one()
        return _contact_to_dict(contact)


@router.delete("/{org_id}/contacts/{contact_id}")
@log_errors
@require_auth
async def delete_organization_contact_route(request: Request, org_id: int, contact_id: int):
    """Remove a contact from an organization (deletes the contact entirely)."""
    from lib.db import async_get_session
    from sqlalchemy import select
    from models.Contact import ContactOrganization

    async with async_get_session() as session:
        # Verify contact is linked to this organization
        stmt = select(ContactOrganization).where(
            ContactOrganization.contact_id == contact_id,
            ContactOrganization.organization_id == org_id
        )
        result = await session.execute(stmt)
        link = result.scalar_one_or_none()
        if not link:
            raise HTTPException(status_code=404, detail="Contact not found for this organization")

    success = await delete_contact(contact_id)
    if not success:
        raise HTTPException(status_code=404, detail="Contact not found")
    return {"success": True}
