"""Serializers for user-related models."""


def tenant_to_response(tenant, individual) -> dict:
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
