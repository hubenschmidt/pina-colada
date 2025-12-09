"""Repository layer for account data access."""

import logging
from typing import List, Optional

from sqlalchemy import select
from sqlalchemy.orm import selectinload

from lib.db import async_get_session
from models.Account import Account
from models.AccountRelationship import AccountRelationship

logger = logging.getLogger(__name__)


async def search_accounts(query: str, tenant_id: Optional[int]) -> List[Account]:
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


async def find_account_relationship(from_account_id: int, to_account_id: int) -> Optional[AccountRelationship]:
    """Find existing relationship between two accounts."""
    async with async_get_session() as session:
        stmt = select(AccountRelationship).where(
            AccountRelationship.from_account_id == from_account_id,
            AccountRelationship.to_account_id == to_account_id,
        )
        result = await session.execute(stmt)
        return result.scalar_one_or_none()


async def create_account_relationship(
    from_account_id: int,
    to_account_id: int,
    user_id: int,
    relationship_type: Optional[str] = None,
    notes: Optional[str] = None,
) -> AccountRelationship:
    """Create a new account relationship."""
    async with async_get_session() as session:
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


async def find_account_relationship_by_id(
    relationship_id: int, account_id: int
) -> Optional[AccountRelationship]:
    """Find relationship by ID where account is either from or to."""
    async with async_get_session() as session:
        stmt = select(AccountRelationship).where(
            AccountRelationship.id == relationship_id,
            (AccountRelationship.from_account_id == account_id) |
            (AccountRelationship.to_account_id == account_id),
        )
        result = await session.execute(stmt)
        return result.scalar_one_or_none()


async def delete_account_relationship(relationship: AccountRelationship) -> None:
    """Delete an account relationship."""
    async with async_get_session() as session:
        # Re-fetch in this session
        rel = await session.get(AccountRelationship, relationship.id)
        if rel:
            await session.delete(rel)
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
