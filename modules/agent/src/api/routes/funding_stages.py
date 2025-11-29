"""Routes for funding stages API endpoints."""

from fastapi import APIRouter, Request
from lib.auth import require_auth
from lib.error_logging import log_errors
from lib.serialization import model_to_dict
from repositories.funding_stage_repository import find_all_funding_stages

router = APIRouter(prefix="/funding-stages", tags=["funding-stages"])


@router.get("")
@log_errors
@require_auth
async def get_funding_stages(request: Request):
    """Get all funding stages."""
    stages = await find_all_funding_stages()
    return [model_to_dict(s, include_relationships=False) for s in stages]
