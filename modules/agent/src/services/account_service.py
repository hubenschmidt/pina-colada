"""Service layer for account business logic."""

import logging
from typing import Optional, List

from sqlalchemy import select
from sqlalchemy.orm import selectinload

from lib.db import async_get_session
from models.Account import Account

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
