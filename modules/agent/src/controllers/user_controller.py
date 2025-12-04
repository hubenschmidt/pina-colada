"""Controller layer for user routing to services."""

from lib.decorators import handle_http_exceptions
from serializers.user import tenant_to_response
from services.user_service import get_user_tenant as get_user_tenant_service


@handle_http_exceptions
async def get_user_tenant(email: str) -> dict:
    """Get tenant info for a user."""
    result = await get_user_tenant_service(email)
    return tenant_to_response(result["tenant"], result["individual"])
