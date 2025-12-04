"""Repository layer for technology data access."""

from typing import List, Optional
from sqlalchemy import select
from lib.db import async_get_session
from models.OrganizationTechnology import OrganizationTechnology
from models.Technology import Technology
from schemas.technology import TechnologyCreate

__all__ = ["TechnologyCreate"]


async def find_all_technologies(category: Optional[str] = None) -> List[Technology]:
    """Find all technologies, optionally filtered by category."""
    async with async_get_session() as session:
        stmt = select(Technology).order_by(Technology.category, Technology.name)
        if category:
            stmt = stmt.where(Technology.category == category)
        result = await session.execute(stmt)
        return list(result.scalars().all())


async def find_technology_by_id(tech_id: int) -> Optional[Technology]:
    """Find technology by ID."""
    async with async_get_session() as session:
        return await session.get(Technology, tech_id)


async def create_technology(name: str, category: str, vendor: Optional[str] = None) -> Technology:
    """Create a new technology."""
    async with async_get_session() as session:
        tech = Technology(name=name, category=category, vendor=vendor)
        session.add(tech)
        await session.commit()
        await session.refresh(tech)
        return tech


async def find_organization_technologies(org_id: int) -> List[OrganizationTechnology]:
    """Find all technologies for an organization."""
    async with async_get_session() as session:
        stmt = (
            select(OrganizationTechnology)
            .where(OrganizationTechnology.organization_id == org_id)
            .order_by(OrganizationTechnology.detected_at.desc())
        )
        result = await session.execute(stmt)
        return list(result.scalars().all())


async def add_organization_technology(
    org_id: int,
    tech_id: int,
    source: Optional[str] = None,
    confidence: Optional[float] = None
) -> OrganizationTechnology:
    """Add a technology to an organization."""
    async with async_get_session() as session:
        org_tech = OrganizationTechnology(
            organization_id=org_id,
            technology_id=tech_id,
            source=source,
            confidence=confidence
        )
        session.add(org_tech)
        await session.commit()
        await session.refresh(org_tech)
        return org_tech


async def remove_organization_technology(org_id: int, tech_id: int) -> bool:
    """Remove a technology from an organization."""
    async with async_get_session() as session:
        stmt = select(OrganizationTechnology).where(
            OrganizationTechnology.organization_id == org_id,
            OrganizationTechnology.technology_id == tech_id
        )
        result = await session.execute(stmt)
        org_tech = result.scalar_one_or_none()
        if not org_tech:
            return False
        await session.delete(org_tech)
        await session.commit()
        return True
