"""Controller layer for funding stage routing to services."""

from typing import List

from lib.decorators import handle_http_exceptions
from serializers.base import model_to_dict
from services.funding_stage_service import get_all_funding_stages as find_all_funding_stages


@handle_http_exceptions
async def get_all_funding_stages() -> List[dict]:
    """Get all funding stages."""
    stages = await find_all_funding_stages()
    return [model_to_dict(s, include_relationships=False) for s in stages]
