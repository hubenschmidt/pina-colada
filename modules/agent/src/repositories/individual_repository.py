"""Repository layer for individual data access."""

import logging
from typing import List, Optional, Dict, Any
from sqlalchemy import select, func
from sqlalchemy.orm import selectinload
from models.Individual import Individual
from models.Account import Account
from lib.db import async_get_session

logger = logging.getLogger(__name__)


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
