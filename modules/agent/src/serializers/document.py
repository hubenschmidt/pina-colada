"""Serializers for document-related models."""

from typing import List, Optional


def document_to_list_response(document, entities=None, tags=None) -> dict:
    """Convert Document model to dict - optimized for list/table view."""
    return {
        "id": document.id,
        "filename": document.filename,
        "file_size": document.file_size,
        "description": document.description,
        "created_at": document.created_at.isoformat() if document.created_at else None,
        "entities": entities or [],
        "tags": tags or [],
        "version_number": document.version_number,
    }


def document_to_detail_response(document, entities=None, tags=None, version_count=None) -> dict:
    """Convert Document model to dictionary - full detail view."""
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
        "created_by": document.created_by,
        "updated_by": document.updated_by,
        "entities": entities or [],
        "tags": tags or [],
        "parent_id": document.parent_id,
        "version_number": document.version_number,
        "is_current_version": document.is_current_version,
        "version_count": version_count,
    }
