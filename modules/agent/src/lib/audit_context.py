"""Audit context for automatic created_by/updated_by population.

Uses contextvars to store current user_id and SQLAlchemy event listeners
to automatically populate audit columns on any model that has them.
"""

import logging
from contextvars import ContextVar
from typing import Optional

from sqlalchemy import event
from sqlalchemy.orm import Mapper

logger = logging.getLogger(__name__)

# Context variable to store current user_id for the request
_current_user_id: ContextVar[Optional[int]] = ContextVar("current_user_id", default=None)


def set_current_user_id(user_id: Optional[int]) -> None:
    """Set the current user_id for audit tracking."""
    _current_user_id.set(user_id)


def get_current_user_id() -> Optional[int]:
    """Get the current user_id for audit tracking."""
    return _current_user_id.get()


def clear_current_user_id() -> None:
    """Clear the current user_id."""
    _current_user_id.set(None)


def _set_audit_columns_on_insert(mapper: Mapper, connection, target) -> None:
    """Event listener to set created_by and updated_by on insert."""
    user_id = get_current_user_id()
    if user_id is None:
        return

    if hasattr(target, "created_by") and target.created_by is None:
        target.created_by = user_id
    if hasattr(target, "updated_by") and target.updated_by is None:
        target.updated_by = user_id


def _set_audit_columns_on_update(mapper: Mapper, connection, target) -> None:
    """Event listener to set updated_by on update."""
    user_id = get_current_user_id()
    if user_id is None:
        return

    if hasattr(target, "updated_by"):
        target.updated_by = user_id


_listeners_registered = False


def register_audit_listeners(base_class) -> None:
    """Register SQLAlchemy event listeners for audit columns.

    Call this once after all models are imported, passing the Base class.
    """
    global _listeners_registered
    if _listeners_registered:
        return

    event.listen(base_class, "before_insert", _set_audit_columns_on_insert, propagate=True)
    event.listen(base_class, "before_update", _set_audit_columns_on_update, propagate=True)
    _listeners_registered = True
    logger.info("Audit event listeners registered")
