"""Repository layer for organization data access."""

import logging
from datetime import datetime
from typing import List, Optional, Dict, Any

from pydantic import BaseModel, field_validator
from sqlalchemy import select, func
from sqlalchemy.orm import selectinload

from lib.db import async_get_session
from lib.validators import validate_phone
from models.Account import Account
from models.Industry import Account_Industry
from models.Organization import Organization
from models.OrganizationTechnology import OrganizationTechnology

logger = logging.getLogger(__name__)


# Pydantic models

class OrganizationCreate(BaseModel):
    name: str
    website: Optional[str] = None
    phone: Optional[str] = None
    employee_count: Optional[int] = None
    employee_count_range_id: Optional[int] = None
    funding_stage_id: Optional[int] = None
    description: Optional[str] = None
    account_id: Optional[int] = None
    industry_ids: Optional[List[int]] = None
    project_ids: Optional[List[int]] = None
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
    employee_count: Optional[int] = None
    employee_count_range_id: Optional[int] = None
    funding_stage_id: Optional[int] = None
    description: Optional[str] = None
    industry_ids: Optional[List[int]] = None
    project_ids: Optional[List[int]] = None
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


class OrgTechnologyCreate(BaseModel):
    technology_id: int
    source: Optional[str] = None
    confidence: Optional[float] = None


class FundingRoundCreate(BaseModel):
    round_type: str
    amount: Optional[int] = None
    announced_date: Optional[str] = None
    lead_investor: Optional[str] = None
    source_url: Optional[str] = None


class SignalCreate(BaseModel):
    signal_type: str
    headline: str
    description: Optional[str] = None
    signal_date: Optional[str] = None
    source: Optional[str] = None
    source_url: Optional[str] = None
    sentiment: Optional[str] = None
    relevance_score: Optional[float] = None


async def find_all_organizations(
    tenant_id: Optional[int] = None,
    page: int = 1,
    page_size: int = 50,
    order_by: str = "updated_at",
    order: str = "DESC",
    search: Optional[str] = None,
) -> tuple[List[Organization], int]:
    """Find organizations with pagination, filtering, and sorting at database level.
    
    Optimized for list view - only loads essential relationships:
    - Account industries (for industry display)
    - Employee count range (for employees display)
    - Funding stage (for funding display)
    """
    async with async_get_session() as session:
        # Base query with minimal relationships needed for list view
        stmt = select(Organization).options(
            selectinload(Organization.account).selectinload(Account.industries),
            selectinload(Organization.employee_count_range),
            selectinload(Organization.funding_stage),
        )
        
        if tenant_id is not None:
            stmt = stmt.join(Account, Organization.account_id == Account.id).where(Account.tenant_id == tenant_id)

        # Apply search filter at DB level
        if search and search.strip():
            search_lower = f"%{search.strip().lower()}%"
            stmt = stmt.where(
                func.lower(Organization.name).like(search_lower)
            )

        # Get total count before pagination
        count_stmt = select(func.count()).select_from(stmt.subquery())
        total = (await session.execute(count_stmt)).scalar() or 0

        # Apply sorting at DB level
        sort_map = {
            "name": Organization.name,
            "updated_at": Organization.updated_at,
            "created_at": Organization.created_at,
            "description": Organization.description,
            "website": Organization.website,
            "funding_stage": Organization.funding_stage_id,
            "employee_count_range": Organization.employee_count_range_id,
            "industries": Organization.account_id,  # Sort by account_id as proxy for industries
        }
        sort_column = sort_map.get(order_by, Organization.updated_at)
        if order.upper() == "ASC":
            stmt = stmt.order_by(sort_column.asc())
        else:
            stmt = stmt.order_by(sort_column.desc())

        # Apply pagination
        offset = (page - 1) * page_size
        stmt = stmt.offset(offset).limit(page_size)

        result = await session.execute(stmt)
        return list(result.scalars().all()), total


async def find_organization_by_id(org_id: int) -> Optional[Organization]:
    """Find organization by ID with all relationships for detail view."""
    async with async_get_session() as session:
        stmt = (
            select(Organization)
            .options(
                selectinload(Organization.account).selectinload(Account.industries),
                selectinload(Organization.account).selectinload(Account.projects),
                selectinload(Organization.employee_count_range),
                selectinload(Organization.funding_stage),
                selectinload(Organization.revenue_range),
                selectinload(Organization.technologies).selectinload(OrganizationTechnology.technology),
                selectinload(Organization.funding_rounds),
                selectinload(Organization.signals),
            )
            .where(Organization.id == org_id)
        )
        result = await session.execute(stmt)
        return result.scalar_one_or_none()


async def find_organization_by_name(name: str, tenant_id: Optional[int] = None) -> Optional[Organization]:
    """Find organization by name (case-insensitive), optionally scoped to tenant (through Account)."""
    name = name.strip() if name else name
    async with async_get_session() as session:
        stmt = select(Organization).where(func.lower(Organization.name) == func.lower(name))
        if tenant_id is not None:
            stmt = stmt.join(Account, Organization.account_id == Account.id).where(Account.tenant_id == tenant_id)
        result = await session.execute(stmt)
        return result.scalar_one_or_none()


async def create_organization(data: Dict[str, Any]) -> Organization:
    """Create a new organization."""
    async with async_get_session() as session:
        try:
            org = Organization(**data)
            session.add(org)
            await session.commit()
            # Re-fetch with relationships loaded
            stmt = (
                select(Organization)
                .options(
                    selectinload(Organization.account).selectinload(Account.industries),
                    selectinload(Organization.employee_count_range),
                    selectinload(Organization.funding_stage),
                    selectinload(Organization.revenue_range),
                )
                .where(Organization.id == org.id)
            )
            result = await session.execute(stmt)
            return result.scalar_one()
        except Exception as e:
            await session.rollback()
            logger.error(f"Failed to create organization: {e}")
            raise


async def get_or_create_organization(
    name: str,
    tenant_id: Optional[int] = None,
    industry_ids: Optional[List[int]] = None
) -> Organization:
    """Get or create organization by name, optionally linking industries to the account."""
    name = name.strip() if name else name
    existing = await find_organization_by_name(name, tenant_id)
    if existing:
        return existing

    # Create account for the organization with tenant_id
    async with async_get_session() as session:
        try:
            account = Account(
                tenant_id=tenant_id,
                name=name
            )
            session.add(account)
            await session.flush()

            org = Organization(
                account_id=account.id,
                name=name
            )
            session.add(org)
            await session.flush()

            # Link industries to the account if provided
            if industry_ids:
                for industry_id in industry_ids:
                    stmt = Account_Industry.insert().values(
                        account_id=account.id,
                        industry_id=industry_id
                    )
                    await session.execute(stmt)

            await session.commit()
            await session.refresh(org)
            return org
        except Exception as e:
            await session.rollback()
            logger.error(f"Failed to create organization: {e}")
            raise


async def update_organization(org_id: int, data: Dict[str, Any]) -> Optional[Organization]:
    """Update an existing organization."""
    async with async_get_session() as session:
        try:
            org = await session.get(Organization, org_id)
            if not org:
                return None

            for key, value in data.items():
                if hasattr(org, key):
                    setattr(org, key, value)

            await session.commit()
            # Re-fetch with relationships loaded
            stmt = (
                select(Organization)
                .options(
                    selectinload(Organization.account).selectinload(Account.industries),
                    selectinload(Organization.employee_count_range),
                    selectinload(Organization.funding_stage),
                    selectinload(Organization.revenue_range),
                )
                .where(Organization.id == org_id)
            )
            result = await session.execute(stmt)
            return result.scalar_one()
        except Exception as e:
            await session.rollback()
            logger.error(f"Failed to update organization: {e}")
            raise


async def delete_organization(org_id: int) -> bool:
    """Delete an organization by ID."""
    async with async_get_session() as session:
        try:
            org = await session.get(Organization, org_id)
            if not org:
                return False
            await session.delete(org)
            await session.commit()
            return True
        except Exception as e:
            await session.rollback()
            logger.error(f"Failed to delete organization: {e}")
            raise


async def search_organizations(query: str, tenant_id: Optional[int] = None) -> List[Organization]:
    """Search organizations by name (case-insensitive partial match), filtered by tenant (through Account)."""
    async with async_get_session() as session:
        stmt = select(Organization).options(
            selectinload(Organization.account).selectinload(Account.industries),
            selectinload(Organization.employee_count_range),
            selectinload(Organization.funding_stage),
            selectinload(Organization.revenue_range),
        ).where(
            func.lower(Organization.name).contains(func.lower(query))
        ).order_by(Organization.name)
        if tenant_id is not None:
            stmt = stmt.join(Account, Organization.account_id == Account.id).where(Account.tenant_id == tenant_id)
        result = await session.execute(stmt)
        return list(result.scalars().all())
