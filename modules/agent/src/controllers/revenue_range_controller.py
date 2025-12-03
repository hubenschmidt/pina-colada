"""Controller layer for revenue range routing to services."""

from typing import List

from lib.decorators import handle_http_exceptions
from lib.serialization import model_to_dict
from repositories.revenue_range_repository import find_all_revenue_ranges


@handle_http_exceptions
async def get_all_revenue_ranges() -> List[dict]:
    """Get all revenue ranges."""
    ranges = await find_all_revenue_ranges()
    return [model_to_dict(r, include_relationships=False) for r in ranges]
