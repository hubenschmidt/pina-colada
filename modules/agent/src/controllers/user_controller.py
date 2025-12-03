"""Controller layer for user routing to services."""

from lib.decorators import handle_http_exceptions
from services.user_service import get_user_tenant as get_user_tenant_service


def _tenant_to_response(tenant, individual) -> dict:
    """Convert tenant and individual to response dict."""
    return {
        "id": tenant.id,
        "tenant": {
            "name": tenant.name,
            "slug": tenant.slug,
            "plan": tenant.plan,
            "industry": tenant.industry,
            "website": tenant.website,
            "employee_count": tenant.employee_count,
        },
        "individual": {
            "first_name": individual.first_name if individual else "",
            "last_name": individual.last_name if individual else "",
            "phone": individual.phone if individual else "",
            "linkedin_url": individual.linkedin_url if individual else "",
            "title": individual.title if individual else "",
        } if individual else None,
        "created_at": tenant.created_at.isoformat() if tenant.created_at else None,
    }


@handle_http_exceptions
async def get_user_tenant(email: str) -> dict:
    """Get tenant info for a user."""
    result = await get_user_tenant_service(email)
    return _tenant_to_response(result["tenant"], result["individual"])
