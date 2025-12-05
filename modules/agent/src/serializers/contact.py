"""Serializers for contact-related models."""

from serializers.account import get_account_type_and_entity_id


def contact_to_list_response(contact) -> dict:
    """Convert Contact model to dict - optimized for list/table view."""
    accounts = [
        {"id": acc.id, "name": acc.name}
        for acc in (contact.accounts or [])
    ]

    return {
        "id": contact.id,
        "first_name": contact.first_name or "",
        "last_name": contact.last_name or "",
        "title": contact.title,
        "email": contact.email,
        "phone": contact.phone,
        "accounts": accounts,
        "updated_at": contact.updated_at.isoformat() if contact.updated_at else None,
    }


def contact_to_detail_response(contact) -> dict:
    """Convert Contact model to dict with linked accounts."""
    accounts = []
    for acc in (contact.accounts or []):
        account_type, entity_id = get_account_type_and_entity_id(acc)
        accounts.append({
            "id": acc.id,
            "name": acc.name,
            "type": account_type,
        })

    return {
        "id": contact.id,
        "first_name": contact.first_name or "",
        "last_name": contact.last_name or "",
        "title": contact.title,
        "department": contact.department,
        "role": contact.role,
        "email": contact.email,
        "phone": contact.phone,
        "is_primary": contact.is_primary,
        "notes": contact.notes,
        "accounts": accounts,
        "created_at": contact.created_at.isoformat() if contact.created_at else None,
        "updated_at": contact.updated_at.isoformat() if contact.updated_at else None,
    }
