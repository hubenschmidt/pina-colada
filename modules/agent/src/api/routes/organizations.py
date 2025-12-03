"""Routes for organizations API endpoints."""

from datetime import datetime
from typing import Optional, List
from fastapi import APIRouter, Request, Query, HTTPException
from pydantic import BaseModel, field_validator
from sqlalchemy import select, delete, update
from sqlalchemy.exc import IntegrityError
from lib.auth import require_auth
from lib.error_logging import log_errors
from lib.validators import validate_phone
from lib.db import async_get_session
from models.Account import Account
from models.Industry import Account_Industry
from models.AccountProject import AccountProject
from models.Contact import Contact, ContactOrganization
from repositories.organization_repository import (
    find_all_organizations,
    find_organization_by_id,
    create_organization,
    update_organization,
    delete_organization,
    search_organizations,
)
from repositories.contact_repository import find_contacts_by_organization, delete_contact
from repositories.technology_repository import (
    find_organization_technologies,
    add_organization_technology,
    remove_organization_technology,
)
from repositories.funding_round_repository import (
    find_funding_rounds_by_org,
    create_funding_round,
    delete_funding_round,
    find_funding_round_by_id,
)
from repositories.company_signal_repository import (
    find_signals_by_org,
    create_signal,
    delete_signal,
    find_signal_by_id,
)


class OrganizationCreate(BaseModel):
    name: str
    website: Optional[str] = None
    phone: Optional[str] = None
    employee_count: Optional[int] = None  # Legacy field
    employee_count_range_id: Optional[int] = None
    funding_stage_id: Optional[int] = None
    description: Optional[str] = None
    account_id: Optional[int] = None
    industry_ids: Optional[List[int]] = None
    project_ids: Optional[List[int]] = None  # Associate with multiple projects
    # Firmographic fields
    revenue_range_id: Optional[int] = None
    founding_year: Optional[int] = None
    headquarters_city: Optional[str] = None
    headquarters_state: Optional[str] = None
    headquarters_country: Optional[str] = None
    company_type: Optional[str] = None
    linkedin_url: Optional[str] = None
    crunchbase_url: Optional[str] = None

    @field_validator("phone")
    @classmethod
    def validate_phone_format(cls, v):
        return validate_phone(v)

    @field_validator("founding_year")
    @classmethod
    def validate_founding_year(cls, v):
        if v is None:
            return v
        current_year = datetime.now().year
        if v < 1800 or v > current_year:
            raise ValueError(f"founding_year must be between 1800 and {current_year}")
        return v


class OrganizationUpdate(BaseModel):
    name: Optional[str] = None
    website: Optional[str] = None
    phone: Optional[str] = None
    employee_count: Optional[int] = None  # Legacy field
    employee_count_range_id: Optional[int] = None
    funding_stage_id: Optional[int] = None
    description: Optional[str] = None
    industry_ids: Optional[List[int]] = None
    project_ids: Optional[List[int]] = None  # Associate with multiple projects
    # Firmographic fields
    revenue_range_id: Optional[int] = None
    founding_year: Optional[int] = None
    headquarters_city: Optional[str] = None
    headquarters_state: Optional[str] = None
    headquarters_country: Optional[str] = None
    company_type: Optional[str] = None
    linkedin_url: Optional[str] = None
    crunchbase_url: Optional[str] = None

    @field_validator("phone")
    @classmethod
    def validate_phone_format(cls, v):
        return validate_phone(v)

    @field_validator("founding_year")
    @classmethod
    def validate_founding_year(cls, v):
        if v is None:
            return v
        current_year = datetime.now().year
        if v < 1800 or v > current_year:
            raise ValueError(f"founding_year must be between 1800 and {current_year}")
        return v


class OrgContactCreate(BaseModel):
    individual_id: Optional[int] = None
    first_name: str
    last_name: str
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


def _technology_to_dict(org_tech):
    return {
        "organization_id": org_tech.organization_id,
        "technology_id": org_tech.technology_id,
        "technology": {
            "id": org_tech.technology.id,
            "name": org_tech.technology.name,
            "category": org_tech.technology.category,
            "vendor": org_tech.technology.vendor,
        } if org_tech.technology else None,
        "detected_at": org_tech.detected_at.isoformat() if org_tech.detected_at else None,
        "source": org_tech.source,
        "confidence": float(org_tech.confidence) if org_tech.confidence else None,
    }


def _funding_round_to_dict(fr):
    return {
        "id": fr.id,
        "organization_id": fr.organization_id,
        "round_type": fr.round_type,
        "amount": fr.amount,
        "announced_date": fr.announced_date.isoformat() if fr.announced_date else None,
        "lead_investor": fr.lead_investor,
        "source_url": fr.source_url,
        "created_at": fr.created_at.isoformat() if fr.created_at else None,
    }


def _signal_to_dict(s):
    return {
        "id": s.id,
        "organization_id": s.organization_id,
        "signal_type": s.signal_type,
        "headline": s.headline,
        "description": s.description,
        "signal_date": s.signal_date.isoformat() if s.signal_date else None,
        "source": s.source,
        "source_url": s.source_url,
        "sentiment": s.sentiment,
        "relevance_score": float(s.relevance_score) if s.relevance_score else None,
        "created_at": s.created_at.isoformat() if s.created_at else None,
    }


def _org_to_list_dict(org) -> dict:
    """Convert Organization model to dictionary - optimized for list/table view.

    Only returns fields needed for table columns:
    Name, Industry, Funding, Employees, Description, Website
    """
    industries = []
    if org.account and org.account.industries:
        industries = [ind.name for ind in org.account.industries]

    return {
        "id": org.id,
        "name": org.name,
        "website": org.website,
        "industries": industries,
        "employee_count_range": org.employee_count_range.label if org.employee_count_range else None,
        "funding_stage": org.funding_stage.label if org.funding_stage else None,
        "description": org.description,
    }


def _org_to_search_dict(org) -> dict:
    """Convert Organization model to dictionary - minimal for search/autocomplete.

    Only returns fields needed for dropdown selection.
    """
    industries = []
    if org.account and org.account.industries:
        industries = [ind.name for ind in org.account.industries]

    return {
        "id": org.id,
        "name": org.name,
        "industries": industries,
    }


def _org_to_dict(org, include_contacts=False, contacts=None, include_research=False):
    """Convert Organization model to dictionary - full detail view."""
    # Get industries from the account
    industries = []
    if org.account and org.account.industries:
        industries = [ind.name for ind in org.account.industries]
    # Get projects from the account
    projects = []
    if org.account and org.account.projects:
        projects = [{"id": p.id, "name": p.name} for p in org.account.projects]
    result = {
        "id": org.id,
        "name": org.name,
        "website": org.website,
        "phone": org.phone,
        "industries": industries,
        "projects": projects,
        "employee_count": org.employee_count,
        "employee_count_range_id": org.employee_count_range_id,
        "employee_count_range": org.employee_count_range.label if org.employee_count_range else None,
        "funding_stage_id": org.funding_stage_id,
        "funding_stage": org.funding_stage.label if org.funding_stage else None,
        "description": org.description,
        # Firmographic fields
        "revenue_range_id": org.revenue_range_id,
        "revenue_range": org.revenue_range.label if org.revenue_range else None,
        "founding_year": org.founding_year,
        "headquarters_city": org.headquarters_city,
        "headquarters_state": org.headquarters_state,
        "headquarters_country": org.headquarters_country,
        "company_type": org.company_type,
        "linkedin_url": org.linkedin_url,
        "crunchbase_url": org.crunchbase_url,
        "created_at": org.created_at.isoformat() if org.created_at else None,
        "updated_at": org.updated_at.isoformat() if org.updated_at else None,
    }
    if include_contacts and contacts is not None:
        result["contacts"] = [_contact_to_dict(c) for c in contacts]
        # Extract unique related individuals from contacts
        seen_ind_ids = set()
        related_individuals = []
        for contact in contacts:
            for ind in (contact.individuals or []):
                if ind.id not in seen_ind_ids:
                    seen_ind_ids.add(ind.id)
                    related_individuals.append({
                        "id": ind.id,
                        "name": f"{ind.first_name} {ind.last_name}".strip(),
                        "type": "individual",
                    })
        result["relationships"] = related_individuals
    if include_research:
        result["technologies"] = [_technology_to_dict(t) for t in (org.technologies or [])]
        result["funding_rounds"] = [_funding_round_to_dict(fr) for fr in (org.funding_rounds or [])]
        result["signals"] = [_signal_to_dict(s) for s in (org.signals or [])]
    return result


@router.get("")
@log_errors
@require_auth
async def get_organizations_route(
    request: Request,
    page: int = Query(1, ge=1),
    limit: int = Query(50, ge=1, le=100),
    order_by: str = Query("updated_at", alias="orderBy"),
    order: str = Query("DESC", regex="^(ASC|DESC)$"),
    search: Optional[str] = Query(None),
):
    """Get organizations for the current tenant with pagination, sorting, and search."""
    tenant_id = getattr(request.state, "tenant_id", None)
    organizations, total = await find_all_organizations(
        tenant_id=tenant_id,
        page=page,
        page_size=limit,
        order_by=order_by,
        order=order,
        search=search,
    )
    
    items = [_org_to_list_dict(org) for org in organizations]
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
async def search_organizations_route(request: Request, q: Optional[str] = Query(None, min_length=1)):
    """Search organizations by name."""
    if not q:
        return []
    tenant_id = getattr(request.state, "tenant_id", None)
    organizations = await search_organizations(q, tenant_id=tenant_id)
    return [_org_to_search_dict(org) for org in organizations]


@router.get("/{org_id}")
@log_errors
@require_auth
async def get_organization_route(request: Request, org_id: int):
    """Get a single organization by ID with contacts and research data."""
    org = await find_organization_by_id(org_id)
    if not org:
        raise HTTPException(status_code=404, detail="Organization not found")
    contacts = await find_contacts_by_organization(org_id)
    return _org_to_dict(org, include_contacts=True, contacts=contacts, include_research=True)


@router.post("")
@log_errors
@require_auth
async def create_organization_route(request: Request, data: OrganizationCreate):
    """Create a new organization with an associated Account."""
    tenant_id = getattr(request.state, "tenant_id", None)
    user_id = getattr(request.state, "user_id", None)
    org_data = data.model_dump(exclude_none=True)
    industry_ids = org_data.pop("industry_ids", None)
    project_ids = org_data.pop("project_ids", None)
    org_data["created_by"] = user_id
    org_data["updated_by"] = user_id

    # Create an Account for the Organization if not provided
    if not org_data.get("account_id"):
        async with async_get_session() as session:
            account = Account(tenant_id=tenant_id, name=org_data.get("name", ""), created_by=user_id, updated_by=user_id)
            session.add(account)
            await session.flush()

            # Link industries to the account
            if industry_ids:
                for industry_id in industry_ids:
                    stmt = Account_Industry.insert().values(
                        account_id=account.id,
                        industry_id=industry_id
                    )
                    await session.execute(stmt)

            # Link projects to the account
            if project_ids:
                for project_id in project_ids:
                    account_project = AccountProject(account_id=account.id, project_id=project_id)
                    session.add(account_project)

            await session.commit()
            await session.refresh(account)
            org_data["account_id"] = account.id

    try:
        org = await create_organization(org_data)
        return _org_to_dict(org)
    except IntegrityError as e:
        if "idx_organization_name_lower" in str(e):
            raise HTTPException(status_code=400, detail="An organization with this name already exists")
        raise


@router.put("/{org_id}")
@log_errors
@require_auth
async def update_organization_route(request: Request, org_id: int, data: OrganizationUpdate):
    """Update an existing organization."""
    user_id = getattr(request.state, "user_id", None)
    org_data = data.model_dump(exclude_unset=True)
    industry_ids = org_data.pop("industry_ids", None)
    project_ids = org_data.pop("project_ids", None)
    org_data["updated_by"] = user_id

    org = await update_organization(org_id, org_data)
    if not org:
        raise HTTPException(status_code=404, detail="Organization not found")

    # Update industries if provided
    if industry_ids is not None and org.account_id:
        async with async_get_session() as session:
            # Remove existing industries
            await session.execute(
                delete(Account_Industry).where(Account_Industry.c.account_id == org.account_id)
            )
            # Add new industries
            for industry_id in industry_ids:
                stmt = Account_Industry.insert().values(
                    account_id=org.account_id,
                    industry_id=industry_id
                )
                await session.execute(stmt)
            await session.commit()

    # Update projects if provided (project_ids in exclude_unset means it was explicitly set)
    if "project_ids" in data.model_dump(exclude_unset=True) and org.account_id:
        async with async_get_session() as session:
            # Remove existing project links
            await session.execute(
                delete(AccountProject).where(AccountProject.account_id == org.account_id)
            )
            # Add new project links
            if project_ids:
                for project_id in project_ids:
                    account_project = AccountProject(account_id=org.account_id, project_id=project_id)
                    session.add(account_project)
            await session.commit()

    # Re-fetch to get updated data
    org = await find_organization_by_id(org_id)

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
    org = await find_organization_by_id(org_id)
    if not org:
        raise HTTPException(status_code=404, detail="Organization not found")
    user_id = getattr(request.state, "user_id", None)

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
            created_by=user_id,
            updated_by=user_id,
        )
        session.add(contact)
        await session.flush()

        # Link to organization only - individual_id is for reference, not bidirectional linking
        org_link = ContactOrganization(contact_id=contact.id, organization_id=org_id, is_primary=data.is_primary)
        session.add(org_link)

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
    user_id = getattr(request.state, "user_id", None)
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
        contact.updated_by = user_id

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


# ==============================================
# Tech Stack endpoints
# ==============================================

class OrgTechnologyCreate(BaseModel):
    technology_id: int
    source: Optional[str] = None
    confidence: Optional[float] = None


@router.get("/{org_id}/technologies")
@log_errors
@require_auth
async def get_organization_technologies_route(request: Request, org_id: int):
    """Get all technologies for an organization."""
    org = await find_organization_by_id(org_id)
    if not org:
        raise HTTPException(status_code=404, detail="Organization not found")
    technologies = await find_organization_technologies(org_id)
    return {"technologies": [_technology_to_dict(t) for t in technologies]}


@router.post("/{org_id}/technologies")
@log_errors
@require_auth
async def add_organization_technology_route(request: Request, org_id: int, data: OrgTechnologyCreate):
    """Add a technology to an organization."""
    org = await find_organization_by_id(org_id)
    if not org:
        raise HTTPException(status_code=404, detail="Organization not found")
    org_tech = await add_organization_technology(
        org_id=org_id,
        tech_id=data.technology_id,
        source=data.source,
        confidence=data.confidence
    )
    return {"organization_technology": _technology_to_dict(org_tech)}


@router.delete("/{org_id}/technologies/{technology_id}")
@log_errors
@require_auth
async def remove_organization_technology_route(request: Request, org_id: int, technology_id: int):
    """Remove a technology from an organization."""
    success = await remove_organization_technology(org_id, technology_id)
    if not success:
        raise HTTPException(status_code=404, detail="Technology not found for this organization")
    return {"success": True}


# ==============================================
# Funding Round endpoints
# ==============================================

class FundingRoundCreate(BaseModel):
    round_type: str
    amount: Optional[int] = None
    announced_date: Optional[str] = None
    lead_investor: Optional[str] = None
    source_url: Optional[str] = None


@router.get("/{org_id}/funding-rounds")
@log_errors
@require_auth
async def get_organization_funding_rounds_route(request: Request, org_id: int):
    """Get all funding rounds for an organization."""
    org = await find_organization_by_id(org_id)
    if not org:
        raise HTTPException(status_code=404, detail="Organization not found")
    funding_rounds = await find_funding_rounds_by_org(org_id)
    return {"funding_rounds": [_funding_round_to_dict(fr) for fr in funding_rounds]}


@router.post("/{org_id}/funding-rounds")
@log_errors
@require_auth
async def create_organization_funding_round_route(request: Request, org_id: int, data: FundingRoundCreate):
    """Create a funding round for an organization."""
    org = await find_organization_by_id(org_id)
    if not org:
        raise HTTPException(status_code=404, detail="Organization not found")
    funding_round = await create_funding_round(
        org_id=org_id,
        round_type=data.round_type,
        amount=data.amount,
        announced_date=data.announced_date,
        lead_investor=data.lead_investor,
        source_url=data.source_url
    )
    return {"funding_round": _funding_round_to_dict(funding_round)}


@router.delete("/{org_id}/funding-rounds/{round_id}")
@log_errors
@require_auth
async def delete_organization_funding_round_route(request: Request, org_id: int, round_id: int):
    """Delete a funding round."""
    funding_round = await find_funding_round_by_id(round_id)
    if not funding_round or funding_round.organization_id != org_id:
        raise HTTPException(status_code=404, detail="Funding round not found")
    await delete_funding_round(round_id)
    return {"success": True}


# ==============================================
# Company Signal endpoints
# ==============================================

class SignalCreate(BaseModel):
    signal_type: str
    headline: str
    description: Optional[str] = None
    signal_date: Optional[str] = None
    source: Optional[str] = None
    source_url: Optional[str] = None
    sentiment: Optional[str] = None
    relevance_score: Optional[float] = None


@router.get("/{org_id}/signals")
@log_errors
@require_auth
async def get_organization_signals_route(
    request: Request,
    org_id: int,
    signal_type: Optional[str] = Query(None),
    limit: int = Query(20, le=100)
):
    """Get signals for an organization."""
    org = await find_organization_by_id(org_id)
    if not org:
        raise HTTPException(status_code=404, detail="Organization not found")
    signals = await find_signals_by_org(org_id, signal_type=signal_type, limit=limit)
    return {"signals": [_signal_to_dict(s) for s in signals]}


@router.post("/{org_id}/signals")
@log_errors
@require_auth
async def create_organization_signal_route(request: Request, org_id: int, data: SignalCreate):
    """Create a signal for an organization."""
    org = await find_organization_by_id(org_id)
    if not org:
        raise HTTPException(status_code=404, detail="Organization not found")
    signal = await create_signal(
        org_id=org_id,
        signal_type=data.signal_type,
        headline=data.headline,
        description=data.description,
        signal_date=data.signal_date,
        source=data.source,
        source_url=data.source_url,
        sentiment=data.sentiment,
        relevance_score=data.relevance_score
    )
    return {"signal": _signal_to_dict(signal)}


@router.delete("/{org_id}/signals/{signal_id}")
@log_errors
@require_auth
async def delete_organization_signal_route(request: Request, org_id: int, signal_id: int):
    """Delete a signal."""
    signal = await find_signal_by_id(signal_id)
    if not signal or signal.organization_id != org_id:
        raise HTTPException(status_code=404, detail="Signal not found")
    await delete_signal(signal_id)
    return {"success": True}
