"""Serializers for contact-related models."""


def contact_to_list_response(contact) -> dict:
    """Convert Contact model to dict - optimized for list/table view."""
    organizations = [
        {"id": org.id, "name": org.name}
        for org in (contact.organizations or [])
    ]

    return {
        "id": contact.id,
        "first_name": contact.first_name or "",
        "last_name": contact.last_name or "",
        "title": contact.title,
        "email": contact.email,
        "phone": contact.phone,
        "organizations": organizations,
        "updated_at": contact.updated_at.isoformat() if contact.updated_at else None,
    }


def contact_to_detail_response(contact) -> dict:
    """Convert Contact model to dict with linked individuals and organizations."""
    first_name = contact.first_name
    last_name = contact.last_name

    if not first_name and contact.individuals:
        first_name = contact.individuals[0].first_name
    if not last_name and contact.individuals:
        last_name = contact.individuals[0].last_name

    individuals = [
        {"id": ind.id, "first_name": ind.first_name, "last_name": ind.last_name}
        for ind in (contact.individuals or [])
    ]

    organizations = [
        {"id": org.id, "name": org.name}
        for org in (contact.organizations or [])
    ]

    return {
        "id": contact.id,
        "first_name": first_name or "",
        "last_name": last_name or "",
        "title": contact.title,
        "department": contact.department,
        "role": contact.role,
        "email": contact.email,
        "phone": contact.phone,
        "is_primary": contact.is_primary,
        "notes": contact.notes,
        "individuals": individuals,
        "organizations": organizations,
        "created_at": contact.created_at.isoformat() if contact.created_at else None,
        "updated_at": contact.updated_at.isoformat() if contact.updated_at else None,
    }
