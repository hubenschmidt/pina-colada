"""Repository layer for individual data access."""

import logging
from typing import Any, Dict, List, Optional, Tuple
from sqlalchemy import and_, delete, func, or_, select
from sqlalchemy.orm import selectinload
from lib.db import async_get_session
from models.Account import Account
from models.AccountProject import AccountProject
from models.AccountRelationship import AccountRelationship
from models.Individual import Individual
from models.IndividualRelationship import IndividualRelationship
from models.Industry import AccountIndustry
from schemas.individual import (

    IndContactCreate,
    IndContactUpdate,
    IndividualCreate,
    IndividualUpdate,
)

__all__ = ["IndividualCreate", "IndividualUpdate", "IndContactCreate", "IndContactUpdate"]

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
                selectinload(Individual.account).selectinload(Account.outgoing_relationships).selectinload(AccountRelationship.to_account).selectinload(Account.organizations),
                selectinload(Individual.account).selectinload(Account.outgoing_relationships).selectinload(AccountRelationship.to_account).selectinload(Account.individuals),
                selectinload(Individual.account).selectinload(Account.incoming_relationships).selectinload(AccountRelationship.from_account).selectinload(Account.organizations),
                selectinload(Individual.account).selectinload(Account.incoming_relationships).selectinload(AccountRelationship.from_account).selectinload(Account.individuals),
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
        # Split query into words and match any word against any field
        words = query.lower().split()
        conditions = []
        for word in words:
            pattern = f"%{word}%"
            conditions.append(
                (func.lower(Individual.first_name).like(pattern)) |
                (func.lower(Individual.last_name).like(pattern)) |
                (func.lower(Individual.email).like(pattern))
            )
        # All words must match (AND logic) - each word can match different fields
        stmt = (
            select(Individual)
            .options(
                selectinload(Individual.account).selectinload(Account.industries),
                selectinload(Individual.account).selectinload(Account.projects),
            )
            .where(and_(*conditions) if conditions else True)
            .order_by(Individual.updated_at.desc())
        )
        if tenant_id is not None:
            stmt = stmt.join(Account, Individual.account_id == Account.id).where(Account.tenant_id == tenant_id)
        result = await session.execute(stmt)
        return list(result.scalars().all())


async def get_individual_account_id(individual_id: int) -> Optional[int]:
    """Get the account_id for an individual."""
    individual = await find_individual_by_id(individual_id)
    return individual.account_id if individual else None


async def check_duplicate_individual(
    first_name: str,
    last_name: str,
    email: Optional[str],
    linkedin_url: Optional[str],
    phone: Optional[str],
) -> bool:
    """Check if an individual with same name and (email, linkedin, or phone) already exists."""
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


async def create_account_for_individual(
    tenant_id: Optional[int],
    account_name: str,
    created_by: Optional[int],
    updated_by: Optional[int],
    industry_ids: Optional[List[int]] = None,
    project_ids: Optional[List[int]] = None,
) -> int:
    """Create an account for an individual with industries and projects."""
    async with async_get_session() as session:
        account = Account(
            tenant_id=tenant_id,
            name=account_name,
            created_by=created_by,
            updated_by=updated_by,
        )
        session.add(account)
        await session.flush()

        if industry_ids:
            for industry_id in industry_ids:
                session.add(AccountIndustry(account_id=account.id, industry_id=industry_id))

        if project_ids:
            for project_id in project_ids:
                session.add(AccountProject(account_id=account.id, project_id=project_id))

        await session.commit()
        return account.id


async def update_account_industries(account_id: int, industry_ids: List[int]) -> None:
    """Replace industries for an account."""
    async with async_get_session() as session:
        await session.execute(delete(AccountIndustry).where(AccountIndustry.account_id == account_id))
        for industry_id in industry_ids:
            session.add(AccountIndustry(account_id=account_id, industry_id=industry_id))
        await session.commit()


async def update_account_projects(account_id: int, project_ids: Optional[List[int]]) -> None:
    """Replace projects for an account."""
    async with async_get_session() as session:
        await session.execute(delete(AccountProject).where(AccountProject.account_id == account_id))
        if project_ids:
            for project_id in project_ids:
                session.add(AccountProject(account_id=account_id, project_id=project_id))
        await session.commit()


async def find_individual_relationships(individual_id: int) -> Tuple[List, List]:
    """Find all relationships for an individual (outgoing and incoming)."""
    async with async_get_session() as session:
        # Get outgoing relationships
        stmt = (
            select(IndividualRelationship)
            .options(selectinload(IndividualRelationship.to_individual))
            .where(IndividualRelationship.from_individual_id == individual_id)
        )
        result = await session.execute(stmt)
        outgoing = list(result.scalars().all())

        # Get incoming relationships
        stmt = (
            select(IndividualRelationship)
            .options(selectinload(IndividualRelationship.from_individual))
            .where(IndividualRelationship.to_individual_id == individual_id)
        )
        result = await session.execute(stmt)
        incoming = list(result.scalars().all())

        return outgoing, incoming


async def find_individual_relationship(from_id: int, to_id: int) -> bool:
    """Check if a relationship exists between two individuals."""
    async with async_get_session() as session:
        stmt = select(IndividualRelationship).where(
            IndividualRelationship.from_individual_id == from_id,
            IndividualRelationship.to_individual_id == to_id,
        )
        result = await session.execute(stmt)
        return result.scalar_one_or_none() is not None


async def create_individual_relationship(
    from_individual_id: int,
    to_individual_id: int,
    relationship_type: Optional[str] = None,
    notes: Optional[str] = None,
):
    """Create a relationship between two individuals."""
    async with async_get_session() as session:
        relationship = IndividualRelationship(
            from_individual_id=from_individual_id,
            to_individual_id=to_individual_id,
            relationship_type=relationship_type,
            notes=notes,
        )
        session.add(relationship)
        await session.commit()
        await session.refresh(relationship)
        return relationship


async def find_individual_relationship_by_id(relationship_id: int, individual_id: int):
    """Find relationship by ID where individual is either from or to."""
    async with async_get_session() as session:
        stmt = select(IndividualRelationship).where(
            IndividualRelationship.id == relationship_id,
            or_(
                IndividualRelationship.from_individual_id == individual_id,
                IndividualRelationship.to_individual_id == individual_id,
            )
        )
        result = await session.execute(stmt)
        return result.scalar_one_or_none()


async def delete_individual_relationship_by_id(relationship_id: int) -> bool:
    """Delete an individual relationship by ID."""
    async with async_get_session() as session:
        rel = await session.get(IndividualRelationship, relationship_id)
        if not rel:
            return False
        await session.delete(rel)
        await session.commit()
        return True
