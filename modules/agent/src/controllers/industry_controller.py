"""Controller layer for industry routing to services."""

from typing import List

from lib.decorators import handle_http_exceptions
from lib.serialization import model_to_dict
from services.industry_service import (
    IndustryCreate,
    get_all_industries as get_all_industries_service,
    create_industry as create_industry_service,
)

# Re-export for routes
__all__ = ["IndustryCreate"]


@handle_http_exceptions
async def get_all_industries() -> List[dict]:
    """Get all industries."""
    industries = await get_all_industries_service()
    return [model_to_dict(ind, include_relationships=False) for ind in industries]


@handle_http_exceptions
async def create_industry(data: IndustryCreate) -> dict:
    """Create a new industry."""
    industry = await create_industry_service(data.name)
    return model_to_dict(industry, include_relationships=False)
