"""Routes for metrics API endpoints (stub for Go backend compatibility)."""

from fastapi import APIRouter

router = APIRouter(prefix="/metrics", tags=["metrics"])


@router.get("/recording/active")
async def get_active_recording():
    """Check if metrics recording is active (stub - actual impl in Go backend)."""
    return {"active": False}
