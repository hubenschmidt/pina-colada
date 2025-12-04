"""Controller layer for technology routing to services."""

from typing import Optional, List

from lib.decorators import handle_http_exceptions
from serializers.base import model_to_dict
from schemas.technology import TechnologyCreate
from services.technology_service import (
    get_all_technologies as get_all_technologies_service,
    get_technology as get_technology_service,
    create_technology as create_technology_service,
)

# Re-export for routes
__all__ = ["TechnologyCreate"]


@handle_http_exceptions
async def get_all_technologies(category: Optional[str] = None) -> List[dict]:
    """Get all technologies."""
    technologies = await get_all_technologies_service(category=category)
    return [model_to_dict(t, include_relationships=False) for t in technologies]


@handle_http_exceptions
async def get_technology(tech_id: int) -> dict:
    """Get a single technology by ID."""
    tech = await get_technology_service(tech_id)
    return model_to_dict(tech, include_relationships=False)


@handle_http_exceptions
async def create_technology(data: TechnologyCreate) -> dict:
    """Create a new technology."""
    tech = await create_technology_service(name=data.name, category=data.category, vendor=data.vendor)
    return model_to_dict(tech, include_relationships=False)
