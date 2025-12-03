"""Repository layer for individual data access."""

import logging
from typing import List, Optional, Dict, Any, Tuple

from pydantic import BaseModel, field_validator
from sqlalchemy import select, func, or_
from sqlalchemy.orm import selectinload

from lib.db import async_get_session
from lib.validators import validate_phone
from models.Account import Account
from models.Individual import Individual

logger = logging.getLogger(__name__)


# Pydantic models

class IndividualCreate(BaseModel):
    first_name: str
    last_name: str
    email: Optional[str] = None
    phone: Optional[str] = None
    linkedin_url: Optional[str] = None
    title: Optional[str] = None
    description: Optional[str] = None
    account_id: Optional[int] = None
    industry_ids: Optional[List[int]] = None
    project_ids: Optional[List[int]] = None
    twitter_url: Optional[str] = None
    github_url: Optional[str] = None
    bio: Optional[str] = None
    seniority_level: Optional[str] = None
    department: Optional[str] = None
    is_decision_maker: Optional[bool] = None
    reports_to_id: Optional[int] = None

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
    description: Optional[str] = None
    industry_ids: Optional[List[int]] = None
    project_ids: Optional[List[int]] = None
    twitter_url: Optional[str] = None
    github_url: Optional[str] = None
    bio: Optional[str] = None
    seniority_level: Optional[str] = None
    department: Optional[str] = None
    is_decision_maker: Optional[bool] = None
    reports_to_id: Optional[int] = None

    @field_validator("phone")
    @classmethod
    def validate_phone_format(cls, v):
        return validate_phone(v)


class IndContactCreate(BaseModel):
    first_name: str
    last_name: str
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


class IndContactUpdate(BaseModel):
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


async def find_all_individuals(tenant_id: Optional[int] = None) -> List[Individual]:
    """Find all individuals, optionally filtered by tenant (through Account)."""
    async with async_get_session() as session:
        stmt = (
            select(Individual)
            .options(
                selectinload(Individual.account).selectinload(Account.industries),
                selectinload(Individual.account).selectinload(Account.projects),
            )
            .order_by(Individual.updated_at.desc())
        )
        if tenant_id is not None:
            stmt = stmt.join(Account, Individual.account_id == Account.id).where(Account.tenant_id == tenant_id)
        result = await session.execute(stmt)
        return list(result.scalars().all())


async def find_all_individuals_paginated(
    page: int = 1,
    page_size: int = 50,
    search: Optional[str] = None,
    order_by: str = "updated_at",
    order: str = "DESC",
    tenant_id: Optional[int] = None,
) -> Tuple[List[Individual], int]:
    """Find individuals with pagination, sorting, and search at DB level.

    Optimized for list view - only loads industries (not projects).
    """
    async with async_get_session() as session:
        stmt = (
            select(Individual)
            .options(selectinload(Individual.account).selectinload(Account.industries))
            .outerjoin(Account, Individual.account_id == Account.id)
        )

        if tenant_id is not None:
            stmt = stmt.where(Account.tenant_id == tenant_id)

        if search and search.strip():
            search_lower = search.strip().lower()
            stmt = stmt.where(
                or_(
                    func.lower(Individual.first_name).contains(search_lower),
                    func.lower(Individual.last_name).contains(search_lower),
                    func.lower(Individual.email).contains(search_lower),
                    func.lower(Individual.title).contains(search_lower),
                )
            )

        count_stmt = select(func.count()).select_from(stmt.subquery())
        count_result = await session.execute(count_stmt)
        total_count = count_result.scalar() or 0

        sort_map = {
            "updated_at": Individual.updated_at,
            "first_name": Individual.first_name,
            "last_name": Individual.last_name,
            "email": Individual.email,
            "phone": Individual.phone,
            "title": Individual.title,
        }
        sort_column = sort_map.get(order_by, Individual.updated_at)
        stmt = stmt.order_by(sort_column.desc() if order.upper() == "DESC" else sort_column.asc())

        offset = (page - 1) * page_size
        stmt = stmt.limit(page_size).offset(offset)

        result = await session.execute(stmt)
        individuals = list(result.scalars().all())

        return individuals, total_count


async def find_individual_by_id(individual_id: int) -> Optional[Individual]:
    """Find individual by ID with all relationships for detail view."""
    async with async_get_session() as session:
        stmt = (
            select(Individual)
            .options(
                selectinload(Individual.account).selectinload(Account.industries),
                selectinload(Individual.account).selectinload(Account.projects),
                selectinload(Individual.reports_to),
                selectinload(Individual.direct_reports),
            )
            .where(Individual.id == individual_id)
        )
        result = await session.execute(stmt)
        return result.scalar_one_or_none()


async def find_individual_by_email(email: str, tenant_id: Optional[int] = None) -> Optional[Individual]:
    """Find individual by email (case-insensitive), optionally scoped to tenant (through Account)."""
    async with async_get_session() as session:
        stmt = select(Individual).where(func.lower(Individual.email) == func.lower(email))
        if tenant_id is not None:
            stmt = stmt.join(Account, Individual.account_id == Account.id).where(Account.tenant_id == tenant_id)
        result = await session.execute(stmt)
        return result.scalar_one_or_none()


async def find_individual_by_name(first_name: str, last_name: str, tenant_id: Optional[int] = None) -> Optional[Individual]:
    """Find individual by first and last name (case-insensitive), optionally scoped to tenant."""
    async with async_get_session() as session:
        stmt = select(Individual).where(
            func.lower(Individual.first_name) == func.lower(first_name),
            func.lower(Individual.last_name) == func.lower(last_name)
        )
        if tenant_id is not None:
            stmt = stmt.join(Account, Individual.account_id == Account.id).where(Account.tenant_id == tenant_id)
        result = await session.execute(stmt)
        return result.scalar_one_or_none()


async def get_or_create_individual(first_name: str, last_name: str, tenant_id: Optional[int] = None) -> Individual:
    """Get or create individual by name."""
    existing = await find_individual_by_name(first_name, last_name, tenant_id)
    if existing:
        return existing

    # Create account for the individual with tenant_id
    async with async_get_session() as session:
        try:
            account = Account(
                tenant_id=tenant_id,
                name=f"{first_name} {last_name}"
            )
            session.add(account)
            await session.flush()

            individual = Individual(
                account_id=account.id,
                first_name=first_name,
                last_name=last_name
            )
            session.add(individual)
            await session.commit()
            await session.refresh(individual)
            return individual
        except Exception as e:
            await session.rollback()
            logger.error(f"Failed to create individual: {e}")
            raise


async def create_individual(data: Dict[str, Any]) -> Individual:
    """Create a new individual."""
    async with async_get_session() as session:
        try:
            individual = Individual(**data)
            session.add(individual)
            await session.commit()
            await session.refresh(individual)
            return individual
        except Exception as e:
            await session.rollback()
            logger.error(f"Failed to create individual: {e}")
            raise


async def update_individual(individual_id: int, data: Dict[str, Any]) -> Optional[Individual]:
    """Update an existing individual."""
    async with async_get_session() as session:
        try:
            individual = await session.get(Individual, individual_id)
            if not individual:
                return None

            for key, value in data.items():
                if hasattr(individual, key):
                    setattr(individual, key, value)

            await session.commit()
            await session.refresh(individual)
            return individual
        except Exception as e:
            await session.rollback()
            logger.error(f"Failed to update individual: {e}")
            raise


async def delete_individual(individual_id: int) -> bool:
    """Delete an individual by ID."""
    async with async_get_session() as session:
        try:
            individual = await session.get(Individual, individual_id)
            if not individual:
                return False
            await session.delete(individual)
            await session.commit()
            return True
        except Exception as e:
            await session.rollback()
            logger.error(f"Failed to delete individual: {e}")
            raise


async def search_individuals(query: str, tenant_id: Optional[int] = None) -> List[Individual]:
    """Search individuals by name or email (case-insensitive partial match), filtered by tenant (through Account)."""
    async with async_get_session() as session:
        search_pattern = func.lower(f"%{query}%")
        stmt = (
            select(Individual)
            .options(
                selectinload(Individual.account).selectinload(Account.industries),
                selectinload(Individual.account).selectinload(Account.projects),
            )
            .where(
                (func.lower(Individual.first_name).like(search_pattern)) |
                (func.lower(Individual.last_name).like(search_pattern)) |
                (func.lower(Individual.email).like(search_pattern))
            )
            .order_by(Individual.updated_at.desc())
        )
        if tenant_id is not None:
            stmt = stmt.join(Account, Individual.account_id == Account.id).where(Account.tenant_id == tenant_id)
        result = await session.execute(stmt)
        return list(result.scalars().all())
