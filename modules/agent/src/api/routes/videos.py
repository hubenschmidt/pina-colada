"""
Video and screenshot serving endpoint for Playwright recordings
"""
from fastapi import APIRouter, HTTPException
from fastapi.responses import FileResponse
import os
import logging

logger = logging.getLogger(__name__)

router = APIRouter(prefix="/videos", tags=["videos"])

VIDEO_DIR = "/tmp/videos"
SCREENSHOT_DIR = "/tmp/playwright-mcp-output"


@router.get("/{filename}")
async def serve_video(filename: str):
    """Serve a video file from the videos directory"""
    # Security: prevent path traversal
    if ".." in filename or "/" in filename:
        raise HTTPException(status_code=400, detail="Invalid filename")

    file_path = os.path.join(VIDEO_DIR, filename)

    if not os.path.exists(file_path):
        logger.error(f"Video not found: {file_path}")
        raise HTTPException(status_code=404, detail="Video not found")

    logger.info(f"Serving video: {file_path}")
    return FileResponse(
        file_path,
        media_type="video/webm",
        headers={"Content-Disposition": f"inline; filename={filename}"}
    )


@router.get("/screenshots/{path:path}")
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
