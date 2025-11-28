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
    first_name: Optional[str] = None
    last_name: Optional[str] = None
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
    first_name: Optional[str] = None
    last_name: Optional[str] = None
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
        # Extract unique related organizations from contacts
        seen_org_ids = set()
        related_orgs = []
        for contact in contacts:
            for org in (contact.organizations or []):
                if org.id not in seen_org_ids:
                    seen_org_ids.add(org.id)
                    related_orgs.append({"id": org.id, "name": org.name, "type": "organization"})
        result["relationships"] = related_orgs
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


def _normalize_linkedin_url(url: str | None) -> str | None:
    """Ensure LinkedIn URL has https:// prefix."""
    if not url:
        return url
    url = url.strip()
    if not url:
        return None
    if url.startswith("http://") or url.startswith("https://"):
        return url
    return f"https://{url}"


async def _check_duplicate_individual(
    first_name: str,
    last_name: str,
    email: str | None,
    linkedin_url: str | None,
    phone: str | None,
) -> bool:
    """Check if an individual with same name and (email, linkedin, or phone) already exists."""
    from lib.db import async_get_session
    from sqlalchemy import select, func, or_
    from models.Individual import Individual

    if not email and not linkedin_url and not phone:
        return False

    async with async_get_session() as session:
        conditions = [
            func.lower(Individual.first_name) == first_name.lower(),
            func.lower(Individual.last_name) == last_name.lower(),
        ]

        match_conditions = []
        if email:
            match_conditions.append(func.lower(Individual.email) == email.lower())
        if linkedin_url:
            match_conditions.append(func.lower(Individual.linkedin_url) == linkedin_url.lower())
        if phone:
            match_conditions.append(Individual.phone == phone)

        stmt = select(Individual).where(*conditions, or_(*match_conditions))
        result = await session.execute(stmt)
        return result.scalar_one_or_none() is not None


@router.post("")
@log_errors
@require_auth
async def create_individual_route(request: Request, data: IndividualCreate):
    """Create a new individual with an associated Account."""
    from lib.db import async_get_session
    from models.Account import Account

    tenant_id = getattr(request.state, "tenant_id", None)
    ind_data = data.model_dump(exclude_none=True)

    # Normalize LinkedIn URL
    if "linkedin_url" in ind_data:
        ind_data["linkedin_url"] = _normalize_linkedin_url(ind_data["linkedin_url"])

    # Check for duplicate individual
    is_duplicate = await _check_duplicate_individual(
        ind_data.get("first_name", ""),
        ind_data.get("last_name", ""),
        ind_data.get("email"),
        ind_data.get("linkedin_url"),
        ind_data.get("phone"),
    )
    if is_duplicate:
        raise HTTPException(status_code=400, detail="Individual already exists")

    # Create an Account for the Individual if not provided
    if not ind_data.get("account_id"):
        async with async_get_session() as session:
            account_name = f"{ind_data.get('first_name', '')} {ind_data.get('last_name', '')}".strip()
            account = Account(tenant_id=tenant_id, name=account_name)
            session.add(account)
            await session.commit()
            await session.refresh(account)
            ind_data["account_id"] = account.id

    individual = await create_individual(ind_data)
    return _ind_to_dict(individual, include_contacts=True, contacts=[])


@router.put("/{individual_id}")
@log_errors
@require_auth
async def update_individual_route(request: Request, individual_id: int, data: IndividualUpdate):
    """Update an existing individual."""
    ind_data = data.model_dump(exclude_unset=True)

    # Normalize LinkedIn URL (converts empty to None)
    if "linkedin_url" in ind_data:
        ind_data["linkedin_url"] = _normalize_linkedin_url(ind_data["linkedin_url"])

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
    """Create a new contact linked to an individual, optionally linked to an organization."""
    from lib.db import async_get_session
    from sqlalchemy import select
    from models.Contact import Contact, ContactIndividual, ContactOrganization

    individual = await find_individual_by_id(individual_id)
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

        # Link to individual
        ind_link = ContactIndividual(contact_id=contact.id, individual_id=individual_id)
        session.add(ind_link)

        # Link to organization if provided
        if data.organization_id:
            org_link = ContactOrganization(
                contact_id=contact.id,
                organization_id=data.organization_id,
                is_primary=data.is_primary
            )
            session.add(org_link)

        await session.commit()

        # Re-fetch to get relationships
        stmt = select(Contact).where(Contact.id == contact.id)
        result = await session.execute(stmt)
        contact = result.scalar_one()
        return _contact_to_dict(contact)


@router.put("/{individual_id}/contacts/{contact_id}")
@log_errors
@require_auth
async def update_individual_contact_route(request: Request, individual_id: int, contact_id: int, data: ContactUpdate):
    """Update a contact for an individual."""
    from lib.db import async_get_session
    from sqlalchemy import select
    from models.Contact import Contact, ContactIndividual

    async with async_get_session() as session:
        # Verify contact exists and is linked to this individual
        stmt = select(ContactIndividual).where(
            ContactIndividual.contact_id == contact_id,
            ContactIndividual.individual_id == individual_id
        )
        result = await session.execute(stmt)
        link = result.scalar_one_or_none()
        if not link:
            raise HTTPException(status_code=404, detail="Contact not found for this individual")

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
            # Unset all other contacts for this individual first
            from sqlalchemy import update
            stmt = select(ContactIndividual.contact_id).where(
                ContactIndividual.individual_id == individual_id,
                ContactIndividual.contact_id != contact_id
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


@router.delete("/{individual_id}/contacts/{contact_id}")
@log_errors
@require_auth
async def delete_individual_contact_route(request: Request, individual_id: int, contact_id: int):
    """Delete a contact for an individual."""
    from lib.db import async_get_session
    from sqlalchemy import select
    from models.Contact import ContactIndividual

    async with async_get_session() as session:
        # Verify contact is linked to this individual
        stmt = select(ContactIndividual).where(
            ContactIndividual.contact_id == contact_id,
            ContactIndividual.individual_id == individual_id
        )
        result = await session.execute(stmt)
        link = result.scalar_one_or_none()
        if not link:
            raise HTTPException(status_code=404, detail="Contact not found for this individual")

    success = await delete_contact(contact_id)
    if not success:
        raise HTTPException(status_code=404, detail="Contact not found")
    return {"success": True}
