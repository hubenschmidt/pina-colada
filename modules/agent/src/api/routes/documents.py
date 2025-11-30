"""Routes for documents API endpoints."""

from typing import Optional

from fastapi import APIRouter, Request, HTTPException, UploadFile, File, Form, Query
from fastapi.responses import RedirectResponse, StreamingResponse
from pydantic import BaseModel

from lib.auth import require_auth
from lib.error_logging import log_errors
from lib.storage import get_storage
from repositories.document_repository import (
    find_documents_by_tenant,
    find_document_by_id,
    create_document,
    update_document,
    delete_document,
    link_document_to_entity,
    unlink_document_from_entity,
    get_document_entities,
    get_document_tags,
    find_document_versions,
    get_version_count,
    find_existing_document_by_filename,
    create_new_version,
    set_current_version,
)


MAX_FILE_SIZE = 10 * 1024 * 1024  # 10MB

ENTITY_TYPE_MAP = {
    "project": "Project",
    "lead": "Lead",
    "account": "Account",
    "organization": "Organization",
    "individual": "Individual",
    "contact": "Contact",
}


def _normalize_entity_type(entity_type: str) -> str:
    """Normalize entity_type to PascalCase."""
    return ENTITY_TYPE_MAP.get(entity_type.lower(), entity_type)


class DocumentUpdate(BaseModel):
    description: Optional[str] = None


class EntityLink(BaseModel):
    entity_type: str
    entity_id: int


def _document_to_dict(document, entities=None, tags=None, version_count=None) -> dict:
    """Convert Document model to dictionary."""
    return {
        "id": document.id,
        "tenant_id": document.tenant_id,
        "user_id": document.user_id,
        "filename": document.filename,
        "content_type": document.content_type,
        "file_size": document.file_size,
        "description": document.description,
        "created_at": document.created_at.isoformat() if document.created_at else None,
        "updated_at": document.updated_at.isoformat() if document.updated_at else None,
        "entities": entities or [],
        "tags": tags or [],
        "parent_id": document.parent_id,
        "version_number": document.version_number,
        "is_current_version": document.is_current_version,
        "version_count": version_count,
    }


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
    """List documents for the tenant with pagination and sorting, optionally filtered by entity, tags, and search."""
    tenant_id = request.state.tenant_id

    normalized_type = _normalize_entity_type(entity_type) if entity_type else None
    tag_list = tags.split(",") if tags else None
    documents, total = await find_documents_by_tenant(
        tenant_id, normalized_type, entity_id, page, limit, order_by, order, search, tag_list
    )

    result = []
    for doc in documents:
        entities = await get_document_entities(doc.id)
        tags = await get_document_tags(doc.id)
        result.append(_document_to_dict(doc, entities, tags))

    total_pages = (total + limit - 1) // limit if limit > 0 else 1
    
    return {
        "items": result,
        "currentPage": page,
        "totalPages": total_pages,
        "total": total,
        "pageSize": limit,
    }


@router.get("/check-filename")
@require_auth
@log_errors
async def check_filename_route(
    request: Request,
    filename: str = Query(...),
    entity_type: str = Query(...),
    entity_id: int = Query(...),
):
    """Check if a filename already exists on an entity.

    Returns the existing document if found, for version creation.
    """
    tenant_id = request.state.tenant_id
    normalized_type = _normalize_entity_type(entity_type)

    existing = await find_existing_document_by_filename(
        tenant_id, filename, normalized_type, entity_id
    )

    if existing:
        entities = await get_document_entities(existing.id)
        tags = await get_document_tags(existing.id)
        version_count = await get_version_count(existing.id, tenant_id)
        return {
            "exists": True,
            "document": _document_to_dict(existing, entities, tags, version_count),
        }

    return {"exists": False, "document": None}


@router.get("/{document_id}")
@require_auth
@log_errors
async def get_document_route(request: Request, document_id: int):
    """Get a single document by ID."""
    tenant_id = request.state.tenant_id
    document = await find_document_by_id(document_id, tenant_id)
    if not document:
        raise HTTPException(status_code=404, detail="Document not found")

    entities = await get_document_entities(document_id)
    tags = await get_document_tags(document_id)
    return _document_to_dict(document, entities, tags)


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
    tenant_id = request.state.tenant_id
    user_id = getattr(request.state, "user_id", None)

    # Read and validate file
    content = await file.read()
    if len(content) > MAX_FILE_SIZE:
        raise HTTPException(status_code=400, detail="File exceeds 10MB limit")

    # Generate storage path
    storage = get_storage()
    # We'll use a simple incrementing ID approach - document ID will be assigned by DB
    # For now, use timestamp-based path
    import time
    temp_id = int(time.time() * 1000)
    storage_path = f"{tenant_id}/{temp_id}/{file.filename}"

    # Upload to storage
    await storage.upload(storage_path, content, file.content_type or "application/octet-stream")

    # Create document record
    document_data = {
        "tenant_id": tenant_id,
        "user_id": user_id,
        "filename": file.filename,
        "content_type": file.content_type or "application/octet-stream",
        "storage_path": storage_path,
        "file_size": len(content),
        "description": description,
    }

    document = await create_document(document_data)

    # Link to entity if provided
    if entity_type and entity_id:
        normalized_type = _normalize_entity_type(entity_type)
        await link_document_to_entity(document.id, normalized_type, entity_id)

    entities = await get_document_entities(document.id)
    tags = await get_document_tags(document.id)
    return _document_to_dict(document, entities, tags)


@router.get("/{document_id}/download")
@require_auth
@log_errors
async def download_document_route(request: Request, document_id: int):
    """Download a document file."""
    tenant_id = request.state.tenant_id
    document = await find_document_by_id(document_id, tenant_id)
    if not document:
        raise HTTPException(status_code=404, detail="Document not found")

    storage = get_storage()

    # For R2, redirect to presigned URL
    # For local, stream the file
    url = storage.get_url(document.storage_path)
    if url.startswith("file://"):
        # Local storage - stream the file
        content = await storage.download(document.storage_path)
        return StreamingResponse(
            iter([content]),
            media_type=document.content_type,
            headers={
                "Content-Disposition": f'attachment; filename="{document.filename}"'
            },
        )
    else:
        # R2 - redirect to presigned URL
        return RedirectResponse(url=url)


@router.put("/{document_id}")
@require_auth
@log_errors
async def update_document_route(
    request: Request, document_id: int, data: DocumentUpdate
):
    """Update a document's metadata."""
    tenant_id = request.state.tenant_id

    update_data = {}
    if data.description is not None:
        update_data["description"] = data.description

    document = await update_document(document_id, tenant_id, update_data)
    if not document:
        raise HTTPException(status_code=404, detail="Document not found")

    entities = await get_document_entities(document_id)
    tags = await get_document_tags(document_id)
    return _document_to_dict(document, entities, tags)


@router.delete("/{document_id}")
@require_auth
@log_errors
async def delete_document_route(request: Request, document_id: int):
    """Delete a document."""
    tenant_id = request.state.tenant_id

    storage_path = await delete_document(document_id, tenant_id)
    if not storage_path:
        raise HTTPException(status_code=404, detail="Document not found")

    # Delete from storage
    storage = get_storage()
    try:
        await storage.delete(storage_path)
    except Exception as e:
        # Log but don't fail - DB record is already deleted
        import logging
        logging.error(f"Failed to delete file from storage: {e}")

    return {"status": "deleted"}


@router.post("/{document_id}/link")
@require_auth
@log_errors
async def link_document_route(request: Request, document_id: int, data: EntityLink):
    """Link a document to an entity."""
    tenant_id = request.state.tenant_id

    # Verify document exists and belongs to tenant
    document = await find_document_by_id(document_id, tenant_id)
    if not document:
        raise HTTPException(status_code=404, detail="Document not found")

    normalized_type = _normalize_entity_type(data.entity_type)
    await link_document_to_entity(document_id, normalized_type, data.entity_id)

    entities = await get_document_entities(document_id)
    tags = await get_document_tags(document_id)
    return _document_to_dict(document, entities, tags)


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
    tenant_id = request.state.tenant_id

    # Verify document exists and belongs to tenant
    document = await find_document_by_id(document_id, tenant_id)
    if not document:
        raise HTTPException(status_code=404, detail="Document not found")

    normalized_type = _normalize_entity_type(entity_type)
    unlinked = await unlink_document_from_entity(document_id, normalized_type, entity_id)
    if not unlinked:
        raise HTTPException(status_code=404, detail="Link not found")

    entities = await get_document_entities(document_id)
    tags = await get_document_tags(document_id)
    return _document_to_dict(document, entities, tags)


# ============== Version Endpoints ==============


@router.get("/{document_id}/versions")
@require_auth
@log_errors
async def get_document_versions_route(request: Request, document_id: int):
    """Get all versions of a document."""
    tenant_id = request.state.tenant_id

    versions = await find_document_versions(document_id, tenant_id)
    if not versions:
        raise HTTPException(status_code=404, detail="Document not found")

    result = []
    for doc in versions:
        entities = await get_document_entities(doc.id)
        tags = await get_document_tags(doc.id)
        result.append(_document_to_dict(doc, entities, tags))

    return {"versions": result}


@router.post("/{document_id}/versions")
@require_auth
@log_errors
async def create_document_version_route(
    request: Request,
    document_id: int,
    file: UploadFile = File(...),
):
    """Create a new version of an existing document."""
    tenant_id = request.state.tenant_id
    user_id = getattr(request.state, "user_id", None)

    # Verify parent document exists
    parent_doc = await find_document_by_id(document_id, tenant_id)
    if not parent_doc:
        raise HTTPException(status_code=404, detail="Document not found")

    # Read and validate file
    content = await file.read()
    if len(content) > MAX_FILE_SIZE:
        raise HTTPException(status_code=400, detail="File exceeds 10MB limit")

    # Generate storage path
    storage = get_storage()
    import time
    temp_id = int(time.time() * 1000)
    storage_path = f"{tenant_id}/{temp_id}/{file.filename}"

    # Upload to storage
    await storage.upload(storage_path, content, file.content_type or "application/octet-stream")

    # Create new version
    new_version = await create_new_version(
        parent_id=document_id,
        tenant_id=tenant_id,
        user_id=user_id,
        storage_path=storage_path,
        file_size=len(content),
        content_type=file.content_type or "application/octet-stream",
    )

    if not new_version:
        raise HTTPException(status_code=500, detail="Failed to create version")

    entities = await get_document_entities(new_version.id)
    tags = await get_document_tags(new_version.id)
    version_count = await get_version_count(new_version.id, tenant_id)
    return _document_to_dict(new_version, entities, tags, version_count)


@router.patch("/{document_id}/set-current")
@require_auth
@log_errors
async def set_current_version_route(request: Request, document_id: int):
    """Set a specific version as the current version."""
    tenant_id = request.state.tenant_id

    document = await set_current_version(document_id, tenant_id)
    if not document:
        raise HTTPException(status_code=404, detail="Document not found")

    entities = await get_document_entities(document_id)
    tags = await get_document_tags(document_id)
    version_count = await get_version_count(document_id, tenant_id)
    return _document_to_dict(document, entities, tags, version_count)
