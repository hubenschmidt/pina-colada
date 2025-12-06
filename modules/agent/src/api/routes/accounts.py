"""Routes for accounts API endpoints."""

from typing import Optional

from fastapi import APIRouter, Request, Query

from controllers.account_controller import (
    search_accounts,
    create_account_relationship,
    delete_account_relationship,
)
from schemas.account import AccountRelationshipCreate


router = APIRouter(prefix="/accounts", tags=["accounts"])


@router.get("/search")
async def search_accounts_route(request: Request, q: Optional[str] = Query(None, min_length=1)):
    """Search accounts by name."""
    if not q:
        return []
    return await search_accounts(request, q)


@router.post("/{account_id}/relationships")
async def create_account_relationship_route(request: Request, account_id: int, data: AccountRelationshipCreate):
    """Create a relationship to another account."""
    return await create_account_relationship(
        request=request,
        from_account_id=account_id,
        to_account_id=data.to_account_id,
        relationship_type=data.relationship_type,
        notes=data.notes,
    )


@router.delete("/{account_id}/relationships/{relationship_id}")
async def delete_account_relationship_route(request: Request, account_id: int, relationship_id: int):
    """Delete a relationship."""
    return await delete_account_relationship(account_id, relationship_id)
