"""Repository layer for lead data access."""

import logging
from typing import List, Optional, Dict, Any
from sqlalchemy import select
from sqlalchemy.orm import joinedload
from models.Lead import Lead
from models.LeadProject import LeadProject
from lib.db import async_get_session

logger = logging.getLogger(__name__)


async def find_all_leads(deal_id: Optional[int] = None, lead_type: Optional[str] = None) -> List[Lead]:
    """Find all leads, optionally filtered by deal or type."""
    async with async_get_session() as session:
        stmt = select(Lead).options(
            joinedload(Lead.current_status),
            joinedload(Lead.deal)
        ).order_by(Lead.created_at.desc())

        if deal_id is not None:
            stmt = stmt.where(Lead.deal_id == deal_id)
        if lead_type is not None:
            stmt = stmt.where(Lead.type == lead_type)

        result = await session.execute(stmt)
        return list(result.unique().scalars().all())


async def find_lead_by_id(lead_id: int) -> Optional[Lead]:
    """Find lead by ID."""
    async with async_get_session() as session:
        stmt = select(Lead).options(
            joinedload(Lead.current_status),
            joinedload(Lead.deal)
        ).where(Lead.id == lead_id)
        result = await session.execute(stmt)
        return result.unique().scalar_one_or_none()


async def create_lead(data: Dict[str, Any], project_ids: List[int]) -> Lead:
    """Create a new lead with required project associations.

    Args:
        data: Lead field data
        project_ids: List of project IDs to associate (required, must have at least one)

    Raises:
        ValueError: If project_ids is empty
    """
    if not project_ids:
        raise ValueError("Lead must be associated with at least one project")

    async with async_get_session() as session:
        try:
            lead = Lead(**data)
            session.add(lead)
            await session.flush()  # Get the lead ID

            # Create project associations
            for project_id in project_ids:
                link = LeadProject(lead_id=lead.id, project_id=project_id)
                session.add(link)

            await session.commit()
            await session.refresh(lead)
            return lead
        except Exception as e:
            await session.rollback()
            logger.error(f"Failed to create lead: {e}")
            raise


async def update_lead(lead_id: int, data: Dict[str, Any]) -> Optional[Lead]:
    """Update an existing lead."""
    async with async_get_session() as session:
        try:
            lead = await session.get(Lead, lead_id)
            if not lead:
                return None

            for key, value in data.items():
                if hasattr(lead, key):
                    setattr(lead, key, value)
            await session.commit()
            await session.refresh(lead)
            return lead
        except Exception as e:
            await session.rollback()
            logger.error(f"Failed to update lead: {e}")
            raise


async def delete_lead(lead_id: int) -> bool:
    """Delete a lead by ID."""
    async with async_get_session() as session:
        try:
            lead = await session.get(Lead, lead_id)
            if not lead:
                return False
            await session.delete(lead)
            await session.commit()
            return True
        except Exception as e:
            await session.rollback()
            logger.error(f"Failed to delete lead: {e}")
            raise


async def find_leads_by_status(status_id: int) -> List[Lead]:
    """Find all leads with a specific status."""
    async with async_get_session() as session:
        stmt = select(Lead).options(
            joinedload(Lead.current_status),
            joinedload(Lead.deal)
        ).where(Lead.current_status_id == status_id).order_by(Lead.created_at.desc())
        result = await session.execute(stmt)
        return list(result.unique().scalars().all())
