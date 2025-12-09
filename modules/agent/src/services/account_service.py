"""Service layer for account business logic."""

import logging
from typing import Optional, List

from fastapi import HTTPException

from repositories.account_repository import (
    search_accounts as search_accounts_repo,
    find_account_relationship,
    create_account_relationship as create_account_relationship_repo,
    find_account_relationship_by_id,
    delete_account_relationship as delete_account_relationship_repo,
    get_account_relationships as get_account_relationships_repo,
)

logger = logging.getLogger(__name__)


async def search_accounts(query: str, tenant_id: Optional[int]) -> List:
    """Search accounts by name."""
    return await search_accounts_repo(query, tenant_id)


async def create_account_relationship(
    from_account_id: int,
    to_account_id: int,
    user_id: int,
    relationship_type: Optional[str] = None,
    notes: Optional[str] = None,
):
    """Create a relationship between two accounts."""
    if from_account_id == to_account_id:
        raise HTTPException(status_code=400, detail="Cannot create relationship to self")

    existing = await find_account_relationship(from_account_id, to_account_id)
    if existing:
        raise HTTPException(status_code=400, detail="Relationship already exists")

    return await create_account_relationship_repo(
        from_account_id=from_account_id,
        to_account_id=to_account_id,
        user_id=user_id,
        relationship_type=relationship_type,
        notes=notes,
    )


async def delete_account_relationship(from_account_id: int, relationship_id: int) -> None:
    """Delete a relationship from an account."""
    relationship = await find_account_relationship_by_id(relationship_id, from_account_id)
    if not relationship:
        raise HTTPException(status_code=404, detail="Relationship not found")
    await delete_account_relationship_repo(relationship)


async def get_account_relationships(account_id: int) -> List:
    """Get all relationships for an account (both outgoing and incoming)."""
    return await get_account_relationships_repo(account_id)
