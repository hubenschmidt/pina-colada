"""Repository layer for opportunity data access."""

import logging
from typing import Any, Dict, List, Optional
from sqlalchemy import func as sql_func, or_, select
from sqlalchemy.orm import joinedload
from lib.db import async_get_session
from models.Account import Account
from models.Deal import Deal
from models.Individual import Individual
from models.Lead import Lead
from models.LeadProject import LeadProject
from models.Opportunity import Opportunity
from models.Organization import Organization
from models.Status import Status
from schemas.opportunity import OpportunityCreate, OpportunityUpdate

__all__ = ["OpportunityCreate", "OpportunityUpdate"]

logger = logging.getLogger(__name__)


async def _load_opportunity_with_relationships(session, opp_id: int) -> Opportunity:
    """Load opportunity with all relationships eagerly."""
    stmt = select(Opportunity).options(
        joinedload(Opportunity.lead).joinedload(Lead.current_status),
        joinedload(Opportunity.lead).joinedload(Lead.tenant),
        joinedload(Opportunity.lead).joinedload(Lead.deal).joinedload(Deal.tenant),
        joinedload(Opportunity.lead).joinedload(Lead.deal).joinedload(Deal.current_status),
        joinedload(Opportunity.lead).joinedload(Lead.account).joinedload(Account.organizations).joinedload(Organization.contacts),
        joinedload(Opportunity.lead).joinedload(Lead.account).joinedload(Account.individuals).joinedload(Individual.contacts),
        joinedload(Opportunity.lead).joinedload(Lead.account).joinedload(Account.industries),
        joinedload(Opportunity.lead).joinedload(Lead.projects),
    ).where(Opportunity.id == opp_id)
    result = await session.execute(stmt)
    return result.unique().scalar_one()


async def find_all_opportunities(
    page: int = 1,
    page_size: int = 50,
    search: Optional[str] = None,
    order_by: str = "updated_at",
    order: str = "DESC",
    tenant_id: Optional[int] = None,
    project_id: Optional[int] = None
) -> tuple[List[Opportunity], int]:
    """Find opportunities with pagination, filtering, and sorting."""
    async with async_get_session() as session:
        stmt = select(Opportunity).options(
            joinedload(Opportunity.lead).joinedload(Lead.current_status),
            joinedload(Opportunity.lead).joinedload(Lead.tenant),
            joinedload(Opportunity.lead).joinedload(Lead.deal).joinedload(Deal.tenant),
            joinedload(Opportunity.lead).joinedload(Lead.deal).joinedload(Deal.current_status),
            joinedload(Opportunity.lead).joinedload(Lead.account).joinedload(Account.organizations).joinedload(Organization.contacts),
            joinedload(Opportunity.lead).joinedload(Lead.account).joinedload(Account.individuals).joinedload(Individual.contacts),
            joinedload(Opportunity.lead).joinedload(Lead.account).joinedload(Account.industries),
            joinedload(Opportunity.lead).joinedload(Lead.projects),
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
                    sql_func.lower(Opportunity.opportunity_name).contains(search_lower),
                    sql_func.lower(Lead.title).contains(search_lower)
                )
            )

        count_stmt = select(sql_func.count()).select_from(stmt.alias())
        count_result = await session.execute(count_stmt)
        total_count = count_result.scalar() or 0

        sort_map = {
            "updated_at": Lead.updated_at,
            "created_at": Lead.created_at,
            "opportunity_name": Opportunity.opportunity_name,
            "estimated_value": Opportunity.estimated_value,
            "expected_close_date": Opportunity.expected_close_date,
        }
        sort_column = sort_map.get(order_by, Lead.updated_at)
        stmt = stmt.order_by(sort_column.desc() if order.upper() == "DESC" else sort_column.asc())

        offset = (page - 1) * page_size
        stmt = stmt.limit(page_size).offset(offset)

        result = await session.execute(stmt)
        opportunities = list(result.unique().scalars().all())

        return opportunities, total_count


async def find_opportunity_by_id(opp_id: int) -> Optional[Opportunity]:
    """Find opportunity by ID with related data."""
    async with async_get_session() as session:
        stmt = select(Opportunity).options(
            joinedload(Opportunity.lead).joinedload(Lead.current_status),
            joinedload(Opportunity.lead).joinedload(Lead.tenant),
            joinedload(Opportunity.lead).joinedload(Lead.deal).joinedload(Deal.tenant),
            joinedload(Opportunity.lead).joinedload(Lead.deal).joinedload(Deal.current_status),
            joinedload(Opportunity.lead).joinedload(Lead.account).joinedload(Account.organizations).joinedload(Organization.contacts),
            joinedload(Opportunity.lead).joinedload(Lead.account).joinedload(Account.individuals).joinedload(Individual.contacts),
            joinedload(Opportunity.lead).joinedload(Lead.account).joinedload(Account.industries),
            joinedload(Opportunity.lead).joinedload(Lead.projects),
        ).where(Opportunity.id == opp_id)
        result = await session.execute(stmt)
        return result.unique().scalar_one_or_none()


async def create_opportunity(data: Dict[str, Any]) -> Opportunity:
    """Create a new opportunity (with Lead parent)."""
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

            title = data.get("title") or f"{account_name} - {data.get('opportunity_name', 'Opportunity')}"

            lead_data: Dict[str, Any] = {
                "deal_id": deal_id,
                "type": "Opportunity",
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

            opp = Opportunity(
                id=lead.id,
                opportunity_name=data.get("opportunity_name", ""),
                estimated_value=data.get("estimated_value"),
                probability=data.get("probability"),
                expected_close_date=data.get("expected_close_date"),
                description=data.get("description"),
            )

            session.add(opp)
            await session.commit()
            return await _load_opportunity_with_relationships(session, opp.id)
        except Exception as e:
            await session.rollback()
            logger.error(f"Failed to create opportunity: {e}")
            raise


async def update_opportunity(opp_id: int, data: Dict[str, Any]) -> Optional[Opportunity]:
    """Update an existing opportunity."""
    async with async_get_session() as session:
        try:
            stmt = select(Opportunity).options(
                joinedload(Opportunity.lead).joinedload(Lead.account).joinedload(Account.organizations),
                joinedload(Opportunity.lead).joinedload(Lead.account).joinedload(Account.individuals)
            ).where(Opportunity.id == opp_id)
            result = await session.execute(stmt)
            opp = result.unique().scalar_one_or_none()

            if not opp:
                return None

            if "opportunity_name" in data and data["opportunity_name"] is not None:
                opp.opportunity_name = data["opportunity_name"]
            if "estimated_value" in data:
                opp.estimated_value = data["estimated_value"]
            if "probability" in data:
                opp.probability = data["probability"]
            if "expected_close_date" in data:
                opp.expected_close_date = data["expected_close_date"]
            if "description" in data:
                opp.description = data["description"]
                if opp.lead:
                    opp.lead.description = data["description"]

            if opp.lead:
                if "title" in data and data["title"] is not None:
                    opp.lead.title = data["title"]
                if "current_status_id" in data and data["current_status_id"] is not None:
                    opp.lead.current_status_id = data["current_status_id"]
                if "source" in data and data["source"] is not None:
                    opp.lead.source = data["source"]

                if "project_ids" in data:
                    from models.LeadProject import LeadProject
                    from sqlalchemy import delete
                    await session.execute(
                        delete(LeadProject).where(LeadProject.lead_id == opp.lead.id)
                    )
                    for pid in (data["project_ids"] or []):
                        lead_project = LeadProject(lead_id=opp.lead.id, project_id=pid)
                        session.add(lead_project)

            await session.commit()
            return await _load_opportunity_with_relationships(session, opp.id)
        except Exception as e:
            await session.rollback()
            logger.error(f"Failed to update opportunity: {e}")
            raise


async def delete_opportunity(opp_id: int) -> bool:
    """Delete an opportunity by ID."""
    async with async_get_session() as session:
        try:
            stmt = select(Opportunity).where(Opportunity.id == opp_id)
            result = await session.execute(stmt)
            opp = result.scalar_one_or_none()
            if not opp:
                return False

            await session.delete(opp)
            await session.commit()
            return True
        except Exception as e:
            await session.rollback()
            logger.error(f"Failed to delete opportunity: {e}")
            raise
