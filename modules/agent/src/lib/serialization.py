"""Serialization utilities for converting SQLAlchemy models to dictionaries."""

from datetime import datetime, date
from typing import Any, Dict, Optional
from sqlalchemy.orm import RelationshipProperty
from sqlalchemy.inspection import inspect


def model_to_dict(model: Any, include_relationships: bool = False) -> Dict[str, Any]:
    """
    Convert a SQLAlchemy ORM model to a dictionary.
    
    Args:
        model: SQLAlchemy model instance
        include_relationships: If True, include related objects as nested dicts
        
    Returns:
        Dictionary representation of the model
    """
    if model is None:
        return {}
    
    result = {}
    mapper = inspect(model.__class__)
    
    for column in mapper.columns:
        value = getattr(model, column.key)
        
        # Handle datetime objects
        if isinstance(value, (datetime, date)):
            result[column.key] = value.isoformat() if value else None
        else:
            result[column.key] = value
    
    if include_relationships:
        for relationship in mapper.relationships:
            rel_value = getattr(model, relationship.key, None)
            
            if rel_value is None:
                result[relationship.key] = None
            elif isinstance(rel_value, list):
                # One-to-many or many-to-many relationship
                result[relationship.key] = [
                    model_to_dict(item, include_relationships=False)
                    for item in rel_value
                ]
            else:
                # Many-to-one or one-to-one relationship
                result[relationship.key] = model_to_dict(
                    rel_value, include_relationships=False
                )
    
    return result

