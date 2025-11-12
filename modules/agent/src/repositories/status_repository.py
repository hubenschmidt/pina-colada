"""Repository layer for status data access."""

import logging
from typing import List, Optional
from sqlalchemy import select
from models.Status import Status, StatusCreateData, StatusUpdateData, orm_to_dict, dict_to_orm, update_orm_from_dict
from lib.db import get_session

logger = logging.getLogger(__name__)


def find_all_statuses() -> List[Status]:
    """Find all statuses."""
    session = get_session()
    try:
        stmt = select(Status).order_by(Status.category, Status.name)
        return list(session.execute(stmt).scalars().all())
    finally:
        session.close()


def find_statuses_by_category(category: str) -> List[Status]:
    """Find statuses by category."""
    session = get_session()
    try:
        stmt = select(Status).where(Status.category == category).order_by(Status.name)
        return list(session.execute(stmt).scalars().all())
    finally:
        session.close()


def find_status_by_id(status_id: int) -> Optional[Status]:
    """Find status by ID."""
    session = get_session()
    try:
        return session.get(Status, status_id)
    finally:
        session.close()


def find_status_by_name(name: str, category: Optional[str] = None) -> Optional[Status]:
    """Find status by name and optional category."""
    session = get_session()
    try:
        stmt = select(Status).where(Status.name == name)
        if category:
            stmt = stmt.where(Status.category == category)
        return session.execute(stmt).scalar_one_or_none()
    finally:
        session.close()


def create_status(data: StatusCreateData) -> Status:
    """Create a new status."""
    session = get_session()
    try:
        status = dict_to_orm(data)
        session.add(status)
        session.commit()
        session.refresh(status)
        return status
    except Exception as e:
        session.rollback()
        logger.error(f"Failed to create status: {e}")
        raise
    finally:
        session.close()


def update_status(status_id: int, data: StatusUpdateData) -> Optional[Status]:
    """Update an existing status."""
    session = get_session()
    try:
        status = session.get(Status, status_id)
        if not status:
            return None

        update_orm_from_dict(status, data)
        session.commit()
        session.refresh(status)
        return status
    except Exception as e:
        session.rollback()
        logger.error(f"Failed to update status: {e}")
        raise
    finally:
        session.close()


def delete_status(status_id: int) -> bool:
    """Delete a status by ID."""
    session = get_session()
    try:
        status = session.get(Status, status_id)
        if not status:
            return False
        session.delete(status)
        session.commit()
        return True
    except Exception as e:
        session.rollback()
        logger.error(f"Failed to delete status: {e}")
        raise
    finally:
        session.close()
