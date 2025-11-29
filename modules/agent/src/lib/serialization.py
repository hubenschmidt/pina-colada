"""Serialization utilities for converting SQLAlchemy models to dictionaries."""

from datetime import datetime, date
from typing import Any, Dict, Optional
from sqlalchemy.orm.exc import DetachedInstanceError
from sqlalchemy.inspection import inspect


def _serialize_value(value: Any) -> Any:
    """Serialize a single value, handling datetime types."""
    if isinstance(value, (datetime, date)) and value:
        return value.isoformat()
    return value


def _serialize_columns(model: Any, mapper) -> Dict[str, Any]:
    """Extract and serialize column values from a model."""
    return {
        column.key: _serialize_value(getattr(model, column.key))
        for column in mapper.columns
    }


def _serialize_relationship(
    rel_value: Any,
    max_depth: int,
    current_depth: int,
    seen: set
) -> Any:
    """Serialize a relationship value (single or list)."""
    if rel_value is None:
        return None

    if isinstance(rel_value, list):
        return [
            model_to_dict(item, True, max_depth, current_depth + 1, seen.copy())
            for item in rel_value
        ]

    return model_to_dict(rel_value, True, max_depth, current_depth + 1, seen.copy())


def _get_relationship_value(model: Any, key: str) -> tuple:
    """Safely get a relationship value, handling detached instances."""
    try:
        return True, getattr(model, key, None)
    except DetachedInstanceError:
        return False, None


def model_to_dict(
    model: Any,
    include_relationships: bool = False,
    max_depth: int = 3,
    _current_depth: int = 0,
    _seen: Optional[set] = None
) -> Dict[str, Any]:
    """Convert a SQLAlchemy ORM model to a dictionary."""
    if model is None:
        return {}

    seen = _seen if _seen is not None else set()
    model_id = id(model)
    if model_id in seen:
        return {}
    seen.add(model_id)

    mapper = inspect(model.__class__)
    result = _serialize_columns(model, mapper)

    if not include_relationships or _current_depth >= max_depth:
        return result

    for relationship in mapper.relationships:
        loaded, rel_value = _get_relationship_value(model, relationship.key)
        if loaded:
            result[relationship.key] = _serialize_relationship(
                rel_value, max_depth, _current_depth, seen
            )

    return result

