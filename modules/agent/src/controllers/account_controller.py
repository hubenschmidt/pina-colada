"""Controller layer for account routing to services."""

from typing import Optional, List

from fastapi import Request

from lib.decorators import handle_http_exceptions
from serializers.account import account_to_response
from services.account_service import (
    search_accounts as search_accounts_service,
    create_account_relationship as create_relationship_service,
    delete_account_relationship as delete_relationship_service,
)


@handle_http_exceptions
async def search_accounts(request: Request, query: str) -> List[dict]:
    """Search accounts by name."""
    tenant_id = request.state.tenant_id
    accounts = await search_accounts_service(query, tenant_id)
    return [account_to_response(a) for a in accounts]


@handle_http_exceptions
async def create_account_relationship(
    request: Request,
    from_account_id: int,
    to_account_id: int,
    relationship_type: Optional[str] = None,
    notes: Optional[str] = None,
) -> dict:
    """Create a relationship between two accounts."""
    user_id = request.state.user_id
    relationship = await create_relationship_service(
        from_account_id=from_account_id,
        to_account_id=to_account_id,
        user_id=user_id,
        relationship_type=relationship_type,
        notes=notes,
    )
    return {"id": relationship.id, "to_account_id": relationship.to_account_id}


@handle_http_exceptions
async def delete_account_relationship(from_account_id: int, relationship_id: int) -> dict:
    """Delete a relationship."""
    await delete_relationship_service(from_account_id, relationship_id)
    return {"success": True}
