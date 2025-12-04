"""Routes for documents API endpoints."""

from typing import Optional

from fastapi import APIRouter, Request, UploadFile, File, Form, Query
from fastapi.responses import RedirectResponse, StreamingResponse

from controllers.document_controller import (
    check_filename,
    create_document_version,
    delete_document,
    download_document,
    get_document,
    get_document_versions,
    get_documents,
    link_document,
    set_current_version,
    unlink_document,
    update_document,
    upload_document,
)
from schemas.document import DocumentUpdate, EntityLink
from lib.auth import require_auth
from lib.error_logging import log_errors


router = APIRouter(prefix="/assets/documents", tags=["documents"])


@router.get("")
@require_auth
@log_errors
async def list_documents_route(
    request: Request,
    page: int = Query(1, ge=1),
    limit: int = Query(50, ge=1, le=100),
    order_by: str = Query("updated_at", alias="orderBy"),
    order: str = Query("DESC", regex="^(ASC|DESC)$"),
    entity_type: Optional[str] = None,
    entity_id: Optional[int] = None,
    search: Optional[str] = Query(None),
    tags: Optional[str] = Query(None),
):
    """List documents with pagination, filtering by entity, tags, and search."""
    return await get_documents(request, page, limit, order_by, order, entity_type, entity_id, search, tags)


@router.get("/check-filename")
@require_auth
@log_errors
async def check_filename_route(
    request: Request,
    filename: str = Query(...),
    entity_type: str = Query(...),
    entity_id: int = Query(...),
):
    """Check if a filename already exists on an entity."""
    return await check_filename(request, filename, entity_type, entity_id)


@router.get("/{document_id}")
@require_auth
@log_errors
async def get_document_route(request: Request, document_id: int):
    """Get a single document by ID."""
    return await get_document(request, document_id)


@router.post("")
@require_auth
@log_errors
async def upload_document_route(
    request: Request,
    file: UploadFile = File(...),
    description: Optional[str] = Form(None),
    entity_type: Optional[str] = Form(None),
    entity_id: Optional[int] = Form(None),
):
    """Upload a new document."""
    content = await file.read()
    content_type = file.content_type or "application/octet-stream"
    return await upload_document(
        request, file.filename, content, content_type, description, entity_type, entity_id
    )


@router.get("/{document_id}/download")
@require_auth
@log_errors
async def download_document_route(request: Request, document_id: int):
    """Download a document file."""
    document, redirect_url, content = await download_document(request, document_id)

    if redirect_url:
        return RedirectResponse(url=redirect_url)

    return StreamingResponse(
        iter([content]),
        media_type=document.content_type,
        headers={"Content-Disposition": f'attachment; filename="{document.filename}"'},
    )


@router.put("/{document_id}")
@require_auth
@log_errors
async def update_document_route(request: Request, document_id: int, data: DocumentUpdate):
    """Update a document's metadata."""
    return await update_document(request, document_id, data)


@router.delete("/{document_id}")
@require_auth
@log_errors
async def delete_document_route(request: Request, document_id: int):
    """Delete a document."""
    return await delete_document(request, document_id)


@router.post("/{document_id}/link")
@require_auth
@log_errors
async def link_document_route(request: Request, document_id: int, data: EntityLink):
    """Link a document to an entity."""
    return await link_document(request, document_id, data)


@router.delete("/{document_id}/link")
@require_auth
@log_errors
async def unlink_document_route(
    request: Request,
    document_id: int,
    entity_type: str,
    entity_id: int,
):
    """Unlink a document from an entity."""
    return await unlink_document(request, document_id, entity_type, entity_id)


# Version endpoints

@router.get("/{document_id}/versions")
@require_auth
@log_errors
async def get_document_versions_route(request: Request, document_id: int):
    """Get all versions of a document."""
    return await get_document_versions(request, document_id)


@router.post("/{document_id}/versions")
@require_auth
@log_errors
async def create_document_version_route(
    request: Request,
    document_id: int,
    file: UploadFile = File(...),
):
    """Create a new version of an existing document."""
    content = await file.read()
    content_type = file.content_type or "application/octet-stream"
    return await create_document_version(request, document_id, file.filename, content, content_type)


@router.patch("/{document_id}/set-current")
@require_auth
@log_errors
async def set_current_version_route(request: Request, document_id: int):
    """Set a specific version as the current version."""
    return await set_current_version(request, document_id)
