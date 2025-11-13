"""Repository layer for lead data access."""

import logging
from typing import List, Optional, Dict, Any
from sqlalchemy import select
from sqlalchemy.orm import joinedload
from models.Lead import Lead
from lib.db import get_session

logger = logging.getLogger(__name__)


def find_all_leads(deal_id: Optional[int] = None, lead_type: Optional[str] = None) -> List[Lead]:
    """Find all leads, optionally filtered by deal or type."""
    session = get_session()
    try:
        stmt = select(Lead).options(
            joinedload(Lead.current_status),
            joinedload(Lead.deal)
        ).order_by(Lead.created_at.desc())

        if deal_id is not None:
            stmt = stmt.where(Lead.deal_id == deal_id)
        if lead_type is not None:
            stmt = stmt.where(Lead.type == lead_type)

        return list(session.execute(stmt).unique().scalars().all())
    finally:
        session.close()


def find_lead_by_id(lead_id: int) -> Optional[Lead]:
    """Find lead by ID."""
    session = get_session()
    try:
        stmt = select(Lead).options(
            joinedload(Lead.current_status),
            joinedload(Lead.deal)
        ).where(Lead.id == lead_id)
        return session.execute(stmt).unique().scalar_one_or_none()
    finally:
        session.close()


def create_lead(data: Dict[str, Any]) -> Lead:
    """Create a new lead."""
    session = get_session()
    try:
        lead = Lead(**data)
        session.add(lead)
        session.commit()
        session.refresh(lead)
        return lead
    except Exception as e:
        session.rollback()
        logger.error(f"Failed to create lead: {e}")
        raise
    finally:
        session.close()


def update_lead(lead_id: int, data: Dict[str, Any]) -> Optional[Lead]:
    """Update an existing lead."""
    session = get_session()
    try:
        lead = session.get(Lead, lead_id)
        if not lead:
            return None

        for key, value in data.items():
            if hasattr(lead, key):
                setattr(lead, key, value)
        session.commit()
        session.refresh(lead)
        return lead
    except Exception as e:
        session.rollback()
        logger.error(f"Failed to update lead: {e}")
        raise
    finally:
        session.close()


def delete_lead(lead_id: int) -> bool:
    """Delete a lead by ID."""
    session = get_session()
    try:
        lead = session.get(Lead, lead_id)
        if not lead:
            return False
        session.delete(lead)
        session.commit()
        return True
    except Exception as e:
        session.rollback()
        logger.error(f"Failed to delete lead: {e}")
        raise
    finally:
        session.close()


def find_leads_by_status(status_id: int) -> List[Lead]:
    """Find all leads with a specific status."""
    session = get_session()
    try:
        stmt = select(Lead).options(
            joinedload(Lead.current_status),
            joinedload(Lead.deal)
        ).where(Lead.current_status_id == status_id).order_by(Lead.created_at.desc())
        return list(session.execute(stmt).unique().scalars().all())
    finally:
        session.close()
