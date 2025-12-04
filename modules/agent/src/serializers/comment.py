"""Serializers for comment-related models."""


def comment_to_response(comment) -> dict:
    """Convert Comment model to dictionary."""
    creator = comment.creator
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
        "id": comment.id,
        "tenant_id": comment.tenant_id,
        "commentable_type": comment.commentable_type,
        "commentable_id": comment.commentable_id,
        "content": comment.content,
        "created_by": comment.created_by,
        "created_by_name": created_by_name,
        "created_by_email": created_by_email,
        "individual_id": individual_id,
        "parent_comment_id": comment.parent_comment_id,
        "created_at": comment.created_at.isoformat() if comment.created_at else None,
        "updated_at": comment.updated_at.isoformat() if comment.updated_at else None,
    }
