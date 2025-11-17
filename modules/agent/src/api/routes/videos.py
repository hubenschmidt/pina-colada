"""
Screenshot serving endpoint for Playwright MCP
"""
from fastapi import APIRouter, HTTPException
from fastapi.responses import FileResponse
import os
import logging

logger = logging.getLogger(__name__)

router = APIRouter(prefix="/screenshots", tags=["screenshots"])

SCREENSHOT_DIR = "/tmp/playwright-mcp-output"


@router.get("/{path:path}")
async def serve_screenshot(path: str):
    """Serve a screenshot file from the playwright-mcp-output directory"""
    # Security: prevent absolute paths
    if path.startswith("/") or ".." in path:
        raise HTTPException(status_code=400, detail="Invalid path")

    file_path = os.path.join(SCREENSHOT_DIR, path)

    if not os.path.exists(file_path):
        logger.error(f"Screenshot not found: {file_path}")
        raise HTTPException(status_code=404, detail="Screenshot not found")

    logger.info(f"Serving screenshot: {file_path}")
    return FileResponse(
        file_path,
        media_type="image/png",
        headers={"Content-Disposition": f"inline; filename={os.path.basename(path)}"}
    )
