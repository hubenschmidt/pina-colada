"""Serialization utilities for converting SQLAlchemy models to dictionaries."""

from datetime import datetime, date
from typing import Any, Dict, Optional
from sqlalchemy.orm import RelationshipProperty
from sqlalchemy.orm.exc import DetachedInstanceError
from sqlalchemy.inspection import inspect


def model_to_dict(
    model: Any,
    include_relationships: bool = False,
    max_depth: int = 3,
    _current_depth: int = 0,
    _seen: Optional[set] = None
) -> Dict[str, Any]:
    """
    Convert a SQLAlchemy ORM model to a dictionary.

    Args:
        model: SQLAlchemy model instance
        include_relationships: If True, include related objects as nested dicts
        max_depth: Maximum depth for nested relationships (default 3)
        _current_depth: Internal counter for current depth
        _seen: Internal set to track visited objects and prevent cycles

    Returns:
        Dictionary representation of the model
    """
    if model is None:
        return {}

    if _seen is None:
        _seen = set()

    model_id = id(model)
    if model_id in _seen:
        return {}
    _seen.add(model_id)

    result = {}
    mapper = inspect(model.__class__)

    for column in mapper.columns:
        value = getattr(model, column.key)

        # Handle datetime objects
        if isinstance(value, (datetime, date)):
            result[column.key] = value.isoformat() if value else None
        else:
            result[column.key] = value

    if include_relationships and _current_depth < max_depth:
        def process_relationship(relationship):
            try:
                rel_value = getattr(model, relationship.key, None)
            except DetachedInstanceError:
                # Relationship not loaded and session is closed - skip it
                return
            
            if rel_value is None:
                result[relationship.key] = None
                return
            
            if isinstance(rel_value, list):
                # One-to-many or many-to-many relationship
                result[relationship.key] = [
                    model_to_dict(
                        item,
                        include_relationships=True,
                        max_depth=max_depth,
                        _current_depth=_current_depth + 1,
                        _seen=_seen.copy()
                    )
                    for item in rel_value
                ]
                return
            
            # Many-to-one or one-to-one relationship
            result[relationship.key] = model_to_dict(
                rel_value,
                include_relationships=True,
                max_depth=max_depth,
                _current_depth=_current_depth + 1,
                _seen=_seen.copy()
            )
        
        for relationship in mapper.relationships:
            process_relationship(relationship)

    return result

