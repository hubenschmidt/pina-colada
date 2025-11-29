"""Routes for accounts API endpoints."""

from typing import Optional
from fastapi import APIRouter, Request, Query
from sqlalchemy import select, or_
from sqlalchemy.orm import selectinload
from lib.auth import require_auth
from lib.error_logging import log_errors
from lib.db import async_get_session
from models.Account import Account


router = APIRouter(prefix="/accounts", tags=["accounts"])


def _get_account_type_and_entity_id(account) -> tuple:
    """Determine account type and entity ID from linked records."""
    if account.organizations:
        return "organization", account.organizations[0].id
    if account.individuals:
        return "individual", account.individuals[0].id
    return "unknown", account.id


def _account_to_dict(account):
    """Convert account to dict with type info."""
    account_type, entity_id = _get_account_type_and_entity_id(account)
    return {
        "id": entity_id,
        "account_id": account.id,
        "name": account.name,
        "type": account_type,
    }


@router.get("/search")
@log_errors
@require_auth
async def search_accounts_route(request: Request, q: Optional[str] = Query(None, min_length=1)):
    """Search accounts by name."""
    if not q:
        return []

    tenant_id = getattr(request.state, "tenant_id", None)

    async with async_get_session() as session:
        search_pattern = f"%{q}%"

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
        accounts = list(result.scalars().all())

        return [_account_to_dict(a) for a in accounts]
