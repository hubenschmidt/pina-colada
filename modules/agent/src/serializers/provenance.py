"""Serializers for provenance-related models."""


def provenance_to_response(p) -> dict:
    """Convert provenance model to response dict."""
    return {
        "id": p.id,
        "entity_type": p.entity_type,
        "entity_id": p.entity_id,
        "field_name": p.field_name,
        "source": p.source,
        "source_url": p.source_url,
        "confidence": float(p.confidence) if p.confidence else None,
        "verified_at": p.verified_at.isoformat() if p.verified_at else None,
        "verified_by": p.verified_by,
        "raw_value": p.raw_value,
        "created_at": p.created_at.isoformat() if p.created_at else None,
        "updated_at": p.updated_at.isoformat() if p.updated_at else None,
    }
