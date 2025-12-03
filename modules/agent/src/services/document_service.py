"""Service layer for document business logic."""

import logging
import time
from typing import Optional, List, Dict, Any, Tuple

from fastapi import HTTPException

from lib.storage import get_storage
from repositories.document_repository import (
    find_documents_by_tenant,
    find_document_by_id,
    create_document as create_document_repo,
    update_document as update_document_repo,
    delete_document as delete_document_repo,
    link_document_to_entity,
    unlink_document_from_entity,
    get_document_entities,
    get_document_tags,
    get_documents_entities_batch,
    get_documents_tags_batch,
    find_document_versions,
    get_version_count,
    find_existing_document_by_filename,
    create_new_version as create_new_version_repo,
    set_current_version as set_current_version_repo,
    DocumentUpdate,
    EntityLink,
)

# Re-export Pydantic models for controllers
__all__ = ["DocumentUpdate", "EntityLink"]

logger = logging.getLogger(__name__)

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


async def get_documents_paginated(
    tenant_id: int,
    page: int,
    limit: int,
    order_by: str,
    order: str,
    entity_type: Optional[str] = None,
    entity_id: Optional[int] = None,
    search: Optional[str] = None,
    tags: Optional[str] = None,
) -> Tuple[List, int, Dict[int, List], Dict[int, List]]:
    """Get documents with pagination and batch-load related data."""
    normalized_type = _normalize_entity_type(entity_type) if entity_type else None
    tag_list = tags.split(",") if tags else None

    documents, total = await find_documents_by_tenant(
        tenant_id, normalized_type, entity_id, page, limit, order_by, order, search, tag_list
    )

    doc_ids = [doc.id for doc in documents]
    entities_map = await get_documents_entities_batch(doc_ids)
    tags_map = await get_documents_tags_batch(doc_ids)

    return documents, total, entities_map, tags_map


async def check_filename_exists(
    tenant_id: int,
    filename: str,
    entity_type: str,
    entity_id: int,
):
    """Check if a filename already exists on an entity."""
    normalized_type = _normalize_entity_type(entity_type)
    existing = await find_existing_document_by_filename(
        tenant_id, filename, normalized_type, entity_id
    )

    if existing:
        entities = await get_document_entities(existing.id)
        tags = await get_document_tags(existing.id)
        version_count = await get_version_count(existing.id, tenant_id)
        return existing, entities, tags, version_count

    return None, None, None, None


async def get_document(document_id: int, tenant_id: int):
    """Get document by ID with entities and tags."""
    document = await find_document_by_id(document_id, tenant_id)
    if not document:
        raise HTTPException(status_code=404, detail="Document not found")

    entities = await get_document_entities(document_id)
    tags = await get_document_tags(document_id)
    return document, entities, tags


async def upload_document(
    tenant_id: int,
    user_id: Optional[int],
    filename: str,
    content: bytes,
    content_type: str,
    description: Optional[str] = None,
    entity_type: Optional[str] = None,
    entity_id: Optional[int] = None,
):
    """Upload a new document to storage and create DB record."""
    if len(content) > MAX_FILE_SIZE:
        raise HTTPException(status_code=400, detail="File exceeds 10MB limit")

    storage = get_storage()
    temp_id = int(time.time() * 1000)
    storage_path = f"{tenant_id}/{temp_id}/{filename}"

    await storage.upload(storage_path, content, content_type)

    document_data = {
        "tenant_id": tenant_id,
        "user_id": user_id,
        "filename": filename,
        "content_type": content_type,
        "storage_path": storage_path,
        "file_size": len(content),
        "description": description,
        "created_by": user_id,
        "updated_by": user_id,
    }

    document = await create_document_repo(document_data)

    if entity_type and entity_id:
        normalized_type = _normalize_entity_type(entity_type)
        await link_document_to_entity(document.id, normalized_type, entity_id)

    entities = await get_document_entities(document.id)
    tags = await get_document_tags(document.id)
    return document, entities, tags


async def download_document(document_id: int, tenant_id: int):
    """Get document for download with storage URL or content."""
    document = await find_document_by_id(document_id, tenant_id)
    if not document:
        raise HTTPException(status_code=404, detail="Document not found")

    storage = get_storage()
    url = storage.get_url(document.storage_path)

    if not url.startswith("file://"):
        return document, url, None

    content = await storage.download(document.storage_path)
    return document, None, content


async def update_document(
    document_id: int,
    tenant_id: int,
    user_id: Optional[int],
    description: Optional[str] = None,
):
    """Update a document's metadata."""
    update_data = {"updated_by": user_id}
    if description is not None:
        update_data["description"] = description

    document = await update_document_repo(document_id, tenant_id, update_data)
    if not document:
        raise HTTPException(status_code=404, detail="Document not found")

    entities = await get_document_entities(document_id)
    tags = await get_document_tags(document_id)
    return document, entities, tags


async def delete_document(document_id: int, tenant_id: int):
    """Delete a document from DB and storage."""
    storage_path = await delete_document_repo(document_id, tenant_id)
    if not storage_path:
        raise HTTPException(status_code=404, detail="Document not found")

    storage = get_storage()
    try:
        await storage.delete(storage_path)
    except Exception as e:
        logger.error(f"Failed to delete file from storage: {e}")

    return True


async def link_document(
    document_id: int,
    tenant_id: int,
    entity_type: str,
    entity_id: int,
):
    """Link a document to an entity."""
    document = await find_document_by_id(document_id, tenant_id)
    if not document:
        raise HTTPException(status_code=404, detail="Document not found")

    normalized_type = _normalize_entity_type(entity_type)
    await link_document_to_entity(document_id, normalized_type, entity_id)

    entities = await get_document_entities(document_id)
    tags = await get_document_tags(document_id)
    return document, entities, tags


async def unlink_document(
    document_id: int,
    tenant_id: int,
    entity_type: str,
    entity_id: int,
):
    """Unlink a document from an entity."""
    document = await find_document_by_id(document_id, tenant_id)
    if not document:
        raise HTTPException(status_code=404, detail="Document not found")

    normalized_type = _normalize_entity_type(entity_type)
    unlinked = await unlink_document_from_entity(document_id, normalized_type, entity_id)
    if not unlinked:
        raise HTTPException(status_code=404, detail="Link not found")

    entities = await get_document_entities(document_id)
    tags = await get_document_tags(document_id)
    return document, entities, tags


async def get_document_versions(document_id: int, tenant_id: int):
    """Get all versions of a document with entities and tags."""
    versions = await find_document_versions(document_id, tenant_id)
    if not versions:
        raise HTTPException(status_code=404, detail="Document not found")

    result = []
    for doc in versions:
        entities = await get_document_entities(doc.id)
        tags = await get_document_tags(doc.id)
        result.append((doc, entities, tags))

    return result


async def create_document_version(
    document_id: int,
    tenant_id: int,
    user_id: Optional[int],
    filename: str,
    content: bytes,
    content_type: str,
):
    """Create a new version of an existing document."""
    parent_doc = await find_document_by_id(document_id, tenant_id)
    if not parent_doc:
        raise HTTPException(status_code=404, detail="Document not found")

    if len(content) > MAX_FILE_SIZE:
        raise HTTPException(status_code=400, detail="File exceeds 10MB limit")

    storage = get_storage()
    temp_id = int(time.time() * 1000)
    storage_path = f"{tenant_id}/{temp_id}/{filename}"

    await storage.upload(storage_path, content, content_type)

    new_version = await create_new_version_repo(
        parent_id=document_id,
        tenant_id=tenant_id,
        user_id=user_id,
        storage_path=storage_path,
        file_size=len(content),
        content_type=content_type,
    )

    if not new_version:
        raise HTTPException(status_code=500, detail="Failed to create version")

    entities = await get_document_entities(new_version.id)
    tags = await get_document_tags(new_version.id)
    version_count = await get_version_count(new_version.id, tenant_id)
    return new_version, entities, tags, version_count


async def set_document_current_version(document_id: int, tenant_id: int):
    """Set a specific version as the current version."""
    document = await set_current_version_repo(document_id, tenant_id)
    if not document:
        raise HTTPException(status_code=404, detail="Document not found")

    entities = await get_document_entities(document_id)
    tags = await get_document_tags(document_id)
    version_count = await get_version_count(document_id, tenant_id)
    return document, entities, tags, version_count
