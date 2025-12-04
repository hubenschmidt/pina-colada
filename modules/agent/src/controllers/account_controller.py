"""Controller layer for account routing to services."""

import logging
from typing import Optional, List

from fastapi import Request

from lib.decorators import handle_http_exceptions
from serializers.account import account_to_response
from services.account_service import search_accounts as search_accounts_service


@handle_http_exceptions
async def search_accounts(request: Request, query: str) -> List[dict]:
    """Search accounts by name."""
    tenant_id = request.state.tenant_id
    accounts = await search_accounts_service(query, tenant_id)
    return [account_to_response(a) for a in accounts]
