"""Controller layer for account routing to services."""

import logging
from typing import Optional, List

from fastapi import Request

from lib.decorators import handle_http_exceptions
from services.account_service import search_accounts as search_accounts_service

logger = logging.getLogger(__name__)


def _get_account_type_and_entity_id(account) -> tuple:
    """Determine account type and entity ID from linked records."""
    if account.organizations:
        return "organization", account.organizations[0].id
    if account.individuals:
        return "individual", account.individuals[0].id
    return "unknown", account.id


def _account_to_response(account) -> dict:
    """Convert account to dict with type info."""
    account_type, entity_id = _get_account_type_and_entity_id(account)
    return {
        "id": entity_id,
        "account_id": account.id,
        "name": account.name,
        "type": account_type,
    }


@handle_http_exceptions
async def search_accounts(request: Request, query: str) -> List[dict]:
    """Search accounts by name."""
    tenant_id = request.state.tenant_id
    accounts = await search_accounts_service(query, tenant_id)
    return [_account_to_response(a) for a in accounts]
