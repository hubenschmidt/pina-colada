"""Repository layer for deal data access."""

import logging
from typing import List, Optional
from sqlalchemy import select
from sqlalchemy.orm import joinedload
from models.Deal import Deal, DealCreateData, DealUpdateData, orm_to_dict, dict_to_orm, update_orm_from_dict
from lib.db import get_session

logger = logging.getLogger(__name__)


def find_all_deals(tenant_id: Optional[int] = None) -> List[Deal]:
    """Find all deals, optionally filtered by tenant."""
    session = get_session()
    try:
        stmt = select(Deal).options(joinedload(Deal.current_status)).order_by(Deal.created_at.desc())
        if tenant_id is not None:
            stmt = stmt.where(Deal.tenant_id == tenant_id)
        return list(session.execute(stmt).unique().scalars().all())
    finally:
        session.close()


def find_deal_by_id(deal_id: int) -> Optional[Deal]:
    """Find deal by ID."""
    session = get_session()
    try:
        stmt = select(Deal).options(joinedload(Deal.current_status)).where(Deal.id == deal_id)
        return session.execute(stmt).unique().scalar_one_or_none()
    finally:
        session.close()


def find_deal_by_name(name: str, tenant_id: Optional[int] = None) -> Optional[Deal]:
    """Find deal by name, optionally scoped to tenant."""
    session = get_session()
    try:
        stmt = select(Deal).where(Deal.name == name)
        if tenant_id is not None:
            stmt = stmt.where(Deal.tenant_id == tenant_id)
        return session.execute(stmt).scalar_one_or_none()
    finally:
        session.close()


def create_deal(data: DealCreateData) -> Deal:
    """Create a new deal."""
    session = get_session()
    try:
        deal = dict_to_orm(data)
        session.add(deal)
        session.commit()
        session.refresh(deal)
        return deal
    except Exception as e:
        session.rollback()
        logger.error(f"Failed to create deal: {e}")
        raise
    finally:
        session.close()


def get_or_create_deal(name: str, tenant_id: Optional[int] = None) -> Deal:
    """Get or create deal by name."""
    existing = find_deal_by_name(name, tenant_id)
    if existing:
        return existing

    data: DealCreateData = {
        "name": name,
        "tenant_id": tenant_id
    }
    return create_deal(data)


def update_deal(deal_id: int, data: DealUpdateData) -> Optional[Deal]:
    """Update an existing deal."""
    session = get_session()
    try:
        deal = session.get(Deal, deal_id)
        if not deal:
            return None

        update_orm_from_dict(deal, data)
        session.commit()
        session.refresh(deal)
        return deal
    except Exception as e:
        session.rollback()
        logger.error(f"Failed to update deal: {e}")
        raise
    finally:
        session.close()


def delete_deal(deal_id: int) -> bool:
    """Delete a deal by ID."""
    session = get_session()
    try:
        deal = session.get(Deal, deal_id)
        if not deal:
            return False
        session.delete(deal)
        session.commit()
        return True
    except Exception as e:
        session.rollback()
        logger.error(f"Failed to delete deal: {e}")
        raise
    finally:
        session.close()
