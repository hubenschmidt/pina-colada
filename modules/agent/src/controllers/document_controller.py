"""Controller layer for document routing to services."""

from typing import Optional

from fastapi import Request

from lib.decorators import handle_http_exceptions
from serializers.common import to_paged_response
from serializers.document import document_to_list_response, document_to_detail_response
from schemas.document import DocumentUpdate
from services.document_service import (
    EntityLink,
    get_documents_paginated,
    check_filename_exists,
    get_document as get_document_service,
    upload_document as upload_document_service,
    download_document as download_document_service,
    update_document as update_document_service,
    delete_document as delete_document_service,
    link_document as link_document_service,
    unlink_document as unlink_document_service,
    get_document_versions as get_versions_service,
    create_document_version as create_version_service,
    set_document_current_version as set_current_version_service,
)

# Re-export for routes
__all__ = ["DocumentUpdate", "EntityLink"]


# Document CRUD

@handle_http_exceptions
async def get_documents(
    request: Request,
    page: int,
    limit: int,
    order_by: str,
    order: str,
    entity_type: Optional[str] = None,
    entity_id: Optional[int] = None,
    search: Optional[str] = None,
    tags: Optional[str] = None,
) -> dict:
    """Get documents with pagination."""
    tenant_id = request.state.tenant_id
    documents, total, entities_map, tags_map = await get_documents_paginated(
        tenant_id, page, limit, order_by, order, entity_type, entity_id, search, tags
    )
    items = [
        document_to_list_response(doc, entities_map.get(doc.id, []), tags_map.get(doc.id, []))
        for doc in documents
    ]
    return to_paged_response(total, page, limit, items)


@handle_http_exceptions
async def check_filename(
    request: Request,
    filename: str,
    entity_type: str,
    entity_id: int,
) -> dict:
    """Check if a filename already exists on an entity."""
    tenant_id = request.state.tenant_id
    existing, entities, tags, version_count = await check_filename_exists(
        tenant_id, filename, entity_type, entity_id
    )

    if existing:
        return {
            "exists": True,
            "document": document_to_detail_response(existing, entities, tags, version_count),
        }

    return {"exists": False, "document": None}


@handle_http_exceptions
async def get_document(request: Request, document_id: int) -> dict:
    """Get document by ID."""
    tenant_id = request.state.tenant_id
    document, entities, tags = await get_document_service(document_id, tenant_id)
    return document_to_detail_response(document, entities, tags)


@handle_http_exceptions
async def upload_document(
    request: Request,
    filename: str,
    content: bytes,
    content_type: str,
    description: Optional[str] = None,
    tags: Optional[str] = None,
    entity_type: Optional[str] = None,
    entity_id: Optional[int] = None,
) -> dict:
    """Upload a new document."""
    tenant_id = request.state.tenant_id
    user_id = request.state.user_id
    document, entities, doc_tags = await upload_document_service(
        tenant_id, user_id, filename, content, content_type, description, tags, entity_type, entity_id
    )
    return document_to_detail_response(document, entities, doc_tags)


@handle_http_exceptions
async def download_document(request: Request, document_id: int):
    """Download a document file. Returns (document, redirect_url, content)."""
    tenant_id = request.state.tenant_id
    return await download_document_service(document_id, tenant_id)


@handle_http_exceptions
async def update_document(request: Request, document_id: int, data: DocumentUpdate) -> dict:
    """Update a document's metadata."""
    tenant_id = request.state.tenant_id
    user_id = request.state.user_id
    document, entities, tags = await update_document_service(
        document_id, tenant_id, user_id, data.description
    )
    return document_to_detail_response(document, entities, tags)


@handle_http_exceptions
async def delete_document(request: Request, document_id: int) -> dict:
    """Delete a document."""
    tenant_id = request.state.tenant_id
    await delete_document_service(document_id, tenant_id)
    return {"status": "deleted"}


@handle_http_exceptions
async def link_document(request: Request, document_id: int, data: EntityLink) -> dict:
    """Link a document to an entity."""
    tenant_id = request.state.tenant_id
    document, entities, tags = await link_document_service(
        document_id, tenant_id, data.entity_type, data.entity_id
    )
    return document_to_detail_response(document, entities, tags)


@handle_http_exceptions
async def unlink_document(
    request: Request,
    document_id: int,
    entity_type: str,
    entity_id: int,
) -> dict:
    """Unlink a document from an entity."""
    tenant_id = request.state.tenant_id
    document, entities, tags = await unlink_document_service(
        document_id, tenant_id, entity_type, entity_id
    )
    return document_to_detail_response(document, entities, tags)


# Version management

@handle_http_exceptions
async def get_document_versions(request: Request, document_id: int) -> dict:
    """Get all versions of a document."""
    tenant_id = request.state.tenant_id
    versions = await get_versions_service(document_id, tenant_id)
    result = [document_to_detail_response(doc, entities, tags) for doc, entities, tags in versions]
    return {"versions": result}


@handle_http_exceptions
async def create_document_version(
    request: Request,
    document_id: int,
    filename: str,
    content: bytes,
    content_type: str,
) -> dict:
    """Create a new version of an existing document."""
    tenant_id = request.state.tenant_id
    user_id = request.state.user_id
    new_version, entities, tags, version_count = await create_version_service(
        document_id, tenant_id, user_id, filename, content, content_type
    )
    return document_to_detail_response(new_version, entities, tags, version_count)


@handle_http_exceptions
async def set_current_version(request: Request, document_id: int) -> dict:
    """Set a specific version as the current version."""
    tenant_id = request.state.tenant_id
    document, entities, tags, version_count = await set_current_version_service(document_id, tenant_id)
    return document_to_detail_response(document, entities, tags, version_count)
