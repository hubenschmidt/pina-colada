"""Repository layer for partnership data access."""

import logging
from typing import Dict, Any, List, Optional

from pydantic import BaseModel
from sqlalchemy import select, func as sql_func, or_
from sqlalchemy.orm import joinedload

from lib.db import async_get_session
from models.Account import Account
from models.Deal import Deal
from models.Individual import Individual
from models.Lead import Lead
from models.LeadProject import LeadProject
from models.Organization import Organization
from models.Partnership import Partnership
from models.Status import Status

logger = logging.getLogger(__name__)


# Pydantic models

class PartnershipCreate(BaseModel):
    account_type: str = "Organization"
    account: Optional[str] = None
    contacts: Optional[List[dict]] = None
    industry: Optional[List[str]] = None
    industry_ids: Optional[List[int]] = None
    title: str
    partnership_name: str
    partnership_type: Optional[str] = None
    start_date: Optional[str] = None
    end_date: Optional[str] = None
    description: Optional[str] = None
    status: str = "Exploring"
    source: str = "manual"
    project_ids: Optional[List[int]] = None


class PartnershipUpdate(BaseModel):
    account: Optional[str] = None
    contacts: Optional[List[dict]] = None
    title: Optional[str] = None
    partnership_name: Optional[str] = None
    partnership_type: Optional[str] = None
    start_date: Optional[str] = None
    end_date: Optional[str] = None
    description: Optional[str] = None
    status: Optional[str] = None
    source: Optional[str] = None
    project_ids: Optional[List[int]] = None


async def _load_partnership_with_relationships(session, partnership_id: int) -> Partnership:
    """Load partnership with all relationships eagerly."""
    stmt = select(Partnership).options(
        joinedload(Partnership.lead).joinedload(Lead.current_status),
        joinedload(Partnership.lead).joinedload(Lead.tenant),
        joinedload(Partnership.lead).joinedload(Lead.deal).joinedload(Deal.tenant),
        joinedload(Partnership.lead).joinedload(Lead.deal).joinedload(Deal.current_status),
        joinedload(Partnership.lead).joinedload(Lead.account).joinedload(Account.organizations).joinedload(Organization.contacts),
        joinedload(Partnership.lead).joinedload(Lead.account).joinedload(Account.individuals).joinedload(Individual.contacts),
        joinedload(Partnership.lead).joinedload(Lead.account).joinedload(Account.industries),
        joinedload(Partnership.lead).joinedload(Lead.projects),
    ).where(Partnership.id == partnership_id)
    result = await session.execute(stmt)
    return result.unique().scalar_one()


async def find_all_partnerships(
    page: int = 1,
    page_size: int = 50,
    search: Optional[str] = None,
    order_by: str = "updated_at",
    order: str = "DESC",
    tenant_id: Optional[int] = None,
    project_id: Optional[int] = None
) -> tuple[List[Partnership], int]:
    """Find partnerships with pagination, filtering, and sorting."""
    async with async_get_session() as session:
        stmt = select(Partnership).options(
            joinedload(Partnership.lead).joinedload(Lead.current_status),
            joinedload(Partnership.lead).joinedload(Lead.tenant),
            joinedload(Partnership.lead).joinedload(Lead.deal).joinedload(Deal.tenant),
            joinedload(Partnership.lead).joinedload(Lead.deal).joinedload(Deal.current_status),
            joinedload(Partnership.lead).joinedload(Lead.account).joinedload(Account.organizations).joinedload(Organization.contacts),
            joinedload(Partnership.lead).joinedload(Lead.account).joinedload(Account.individuals).joinedload(Individual.contacts),
            joinedload(Partnership.lead).joinedload(Lead.account).joinedload(Account.industries),
            joinedload(Partnership.lead).joinedload(Lead.projects),
        ).join(Lead).outerjoin(Account, Lead.account_id == Account.id).outerjoin(Organization, Account.id == Organization.account_id)

        if tenant_id is not None:
            stmt = stmt.where(Lead.tenant_id == tenant_id)

        if project_id is not None:
            stmt = stmt.join(LeadProject, Lead.id == LeadProject.lead_id).where(LeadProject.project_id == project_id)

        if search and search.strip():
            search_lower = search.strip().lower()
            stmt = stmt.where(
                or_(
                    sql_func.lower(Organization.name).contains(search_lower),
                    sql_func.lower(Partnership.partnership_name).contains(search_lower),
                    sql_func.lower(Lead.title).contains(search_lower)
                )
            )

        count_stmt = select(sql_func.count()).select_from(stmt.alias())
        count_result = await session.execute(count_stmt)
        total_count = count_result.scalar() or 0

        sort_map = {
            "updated_at": Partnership.updated_at,
            "created_at": Lead.created_at,
            "partnership_name": Partnership.partnership_name,
            "start_date": Partnership.start_date,
            "end_date": Partnership.end_date,
        }
        sort_column = sort_map.get(order_by, Partnership.updated_at)
        stmt = stmt.order_by(sort_column.desc() if order.upper() == "DESC" else sort_column.asc())

        offset = (page - 1) * page_size
        stmt = stmt.limit(page_size).offset(offset)

        result = await session.execute(stmt)
        partnerships = list(result.unique().scalars().all())

        return partnerships, total_count


async def find_partnership_by_id(partnership_id: int) -> Optional[Partnership]:
    """Find partnership by ID with related data."""
    async with async_get_session() as session:
        stmt = select(Partnership).options(
            joinedload(Partnership.lead).joinedload(Lead.current_status),
            joinedload(Partnership.lead).joinedload(Lead.tenant),
            joinedload(Partnership.lead).joinedload(Lead.deal).joinedload(Deal.tenant),
            joinedload(Partnership.lead).joinedload(Lead.deal).joinedload(Deal.current_status),
            joinedload(Partnership.lead).joinedload(Lead.account).joinedload(Account.organizations).joinedload(Organization.contacts),
            joinedload(Partnership.lead).joinedload(Lead.account).joinedload(Account.individuals).joinedload(Individual.contacts),
            joinedload(Partnership.lead).joinedload(Lead.account).joinedload(Account.industries),
            joinedload(Partnership.lead).joinedload(Lead.projects),
        ).where(Partnership.id == partnership_id)
        result = await session.execute(stmt)
        return result.unique().scalar_one_or_none()


async def create_partnership(data: Dict[str, Any]) -> Partnership:
    """Create a new partnership (with Lead parent)."""
    async with async_get_session() as session:
        try:
            account_id = data.get("account_id")
            account_name = data.get("account_name", "Unknown")
            tenant_id = data.get("tenant_id")
            deal_id = data.get("deal_id")
            status_id = data.get("current_status_id")

            if not account_id:
                raise ValueError("account_id is required")
            if not deal_id:
                raise ValueError("deal_id is required")

            title = data.get("title") or f"{account_name} - {data.get('partnership_name', 'Partnership')}"

            lead_data: Dict[str, Any] = {
                "deal_id": deal_id,
                "type": "Partnership",
                "title": title,
                "description": data.get("description"),
                "source": data.get("source", "manual"),
                "current_status_id": status_id,
                "account_id": account_id,
                "tenant_id": tenant_id,
                "created_by": data.get("created_by"),
                "updated_by": data.get("updated_by"),
            }

            lead = Lead(**lead_data)
            session.add(lead)
            await session.flush()

            project_ids = data.get("project_ids") or []
            for pid in project_ids:
                lead_project = LeadProject(lead_id=lead.id, project_id=pid)
                session.add(lead_project)

            partnership = Partnership(
                id=lead.id,
                partnership_name=data.get("partnership_name", ""),
                partnership_type=data.get("partnership_type"),
                start_date=data.get("start_date"),
                end_date=data.get("end_date"),
                description=data.get("description"),
                created_by=data.get("created_by"),
                updated_by=data.get("updated_by"),
            )

            session.add(partnership)
            await session.commit()
            return await _load_partnership_with_relationships(session, partnership.id)
        except Exception as e:
            await session.rollback()
            logger.error(f"Failed to create partnership: {e}")
            raise


async def update_partnership(partnership_id: int, data: Dict[str, Any]) -> Optional[Partnership]:
    """Update an existing partnership."""
    async with async_get_session() as session:
        try:
            stmt = select(Partnership).options(
                joinedload(Partnership.lead).joinedload(Lead.account).joinedload(Account.organizations),
                joinedload(Partnership.lead).joinedload(Lead.account).joinedload(Account.individuals)
            ).where(Partnership.id == partnership_id)
            result = await session.execute(stmt)
            partnership = result.unique().scalar_one_or_none()

            if not partnership:
                return None

            if "partnership_name" in data and data["partnership_name"] is not None:
                partnership.partnership_name = data["partnership_name"]
            if "partnership_type" in data:
                partnership.partnership_type = data["partnership_type"]
            if "start_date" in data:
                partnership.start_date = data["start_date"]
            if "end_date" in data:
                partnership.end_date = data["end_date"]
            if "description" in data:
                partnership.description = data["description"]
                if partnership.lead:
                    partnership.lead.description = data["description"]

            if partnership.lead:
                if "title" in data and data["title"] is not None:
                    partnership.lead.title = data["title"]
                if "current_status_id" in data and data["current_status_id"] is not None:
                    partnership.lead.current_status_id = data["current_status_id"]
                if "source" in data and data["source"] is not None:
                    partnership.lead.source = data["source"]

                if "project_ids" in data:
                    from models.LeadProject import LeadProject
                    from sqlalchemy import delete
                    await session.execute(
                        delete(LeadProject).where(LeadProject.lead_id == partnership.lead.id)
                    )
                    for pid in (data["project_ids"] or []):
                        lead_project = LeadProject(lead_id=partnership.lead.id, project_id=pid)
                        session.add(lead_project)

            await session.commit()
            return await _load_partnership_with_relationships(session, partnership.id)
        except Exception as e:
            await session.rollback()
            logger.error(f"Failed to update partnership: {e}")
            raise


async def delete_partnership(partnership_id: int) -> bool:
    """Delete a partnership by ID."""
    async with async_get_session() as session:
        try:
            stmt = select(Partnership).where(Partnership.id == partnership_id)
            result = await session.execute(stmt)
            partnership = result.scalar_one_or_none()
            if not partnership:
                return False

            await session.delete(partnership)
            await session.commit()
            return True
        except Exception as e:
            await session.rollback()
            logger.error(f"Failed to delete partnership: {e}")
            raise
