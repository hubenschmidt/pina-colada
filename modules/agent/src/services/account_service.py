"""Service layer for account business logic."""

import logging
from typing import Optional, List

from fastapi import HTTPException
from sqlalchemy import select
from sqlalchemy.orm import selectinload

from lib.db import async_get_session
from models.Account import Account
from models.AccountRelationship import AccountRelationship

logger = logging.getLogger(__name__)


async def search_accounts(query: str, tenant_id: Optional[int]) -> List:
    """Search accounts by name."""
    async with async_get_session() as session:
        search_pattern = f"%{query}%"
        stmt = (
            select(Account)
            .options(selectinload(Account.organizations), selectinload(Account.individuals))
            .where(Account.name.ilike(search_pattern))
            .order_by(Account.name)
            .limit(20)
        )
        if tenant_id:
            stmt = stmt.where(Account.tenant_id == tenant_id)
        result = await session.execute(stmt)
        return list(result.scalars().all())


async def create_account_relationship(
    from_account_id: int,
    to_account_id: int,
    user_id: int,
    relationship_type: Optional[str] = None,
    notes: Optional[str] = None,
) -> AccountRelationship:
    """Create a relationship between two accounts."""
    if from_account_id == to_account_id:
        raise HTTPException(status_code=400, detail="Cannot create relationship to self")

    async with async_get_session() as session:
        # Check if relationship already exists
        stmt = select(AccountRelationship).where(
            AccountRelationship.from_account_id == from_account_id,
            AccountRelationship.to_account_id == to_account_id,
        )
        existing = (await session.execute(stmt)).scalar_one_or_none()
        if existing:
            raise HTTPException(status_code=400, detail="Relationship already exists")

        relationship = AccountRelationship(
            from_account_id=from_account_id,
            to_account_id=to_account_id,
            relationship_type=relationship_type,
            notes=notes,
            created_by=user_id,
            updated_by=user_id,
        )
        session.add(relationship)
        await session.commit()
        await session.refresh(relationship)
        return relationship


async def delete_account_relationship(from_account_id: int, relationship_id: int) -> None:
    """Delete a relationship from an account."""
    async with async_get_session() as session:
        stmt = select(AccountRelationship).where(
            AccountRelationship.id == relationship_id,
            (AccountRelationship.from_account_id == from_account_id) |
            (AccountRelationship.to_account_id == from_account_id),
        )
        relationship = (await session.execute(stmt)).scalar_one_or_none()
        if not relationship:
            raise HTTPException(status_code=404, detail="Relationship not found")
        await session.delete(relationship)
        await session.commit()


async def get_account_relationships(account_id: int) -> List[AccountRelationship]:
    """Get all relationships for an account (both outgoing and incoming)."""
    async with async_get_session() as session:
        stmt = (
            select(AccountRelationship)
            .options(
                selectinload(AccountRelationship.from_account).selectinload(Account.organizations),
                selectinload(AccountRelationship.from_account).selectinload(Account.individuals),
                selectinload(AccountRelationship.to_account).selectinload(Account.organizations),
                selectinload(AccountRelationship.to_account).selectinload(Account.individuals),
            )
            .where(
                (AccountRelationship.from_account_id == account_id) |
                (AccountRelationship.to_account_id == account_id)
            )
        )
        result = await session.execute(stmt)
        return list(result.scalars().all())
