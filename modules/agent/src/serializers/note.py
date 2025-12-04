"""Serializers for note-related models."""


def note_to_response(note) -> dict:
    """Convert Note model to dictionary."""
    creator = note.creator
    created_by_name = None
    created_by_email = None
    individual_id = None
    if creator:
        first = creator.first_name or ""
        last = creator.last_name or ""
        created_by_name = f"{first} {last}".strip() or None
        created_by_email = creator.email
        individual_id = creator.individual_id

    return {
        "id": note.id,
        "tenant_id": note.tenant_id,
        "entity_type": note.entity_type,
        "entity_id": note.entity_id,
        "content": note.content,
        "created_by": note.created_by,
        "created_by_name": created_by_name,
        "created_by_email": created_by_email,
        "individual_id": individual_id,
        "created_at": note.created_at.isoformat() if note.created_at else None,
        "updated_at": note.updated_at.isoformat() if note.updated_at else None,
    }
