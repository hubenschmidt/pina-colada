"""Repository layer for task data access."""

import logging
from typing import Any, Dict, List, Optional, Tuple
from sqlalchemy import and_, func, or_, select
from sqlalchemy.orm import selectinload
from lib.db import async_get_session
from models.Account import Account
from models.AccountProject import AccountProject
from models.Deal import Deal
from models.Individual import Individual
from models.Job import Job
from models.Lead import Lead
from models.LeadProject import LeadProject
from models.Opportunity import Opportunity
from models.Organization import Organization
from models.Partnership import Partnership
from models.Project import Project
from models.Status import Status
from models.Task import Task
from schemas.task import TaskCreate, TaskUpdate

__all__ = ["TaskCreate", "TaskUpdate"]

logger = logging.getLogger(__name__)


async def find_all_tasks(
    tenant_id: int,
    project_id: Optional[int] = None,
    page: int = 1,
    page_size: int = 20,
    order_by: str = "created_at",
    order: str = "DESC",
    search: Optional[str] = None,
) -> Tuple[List[Task], int]:
    """Find all tasks with optional project scope filtering.

    When project_id is provided, returns tasks for:
    - The project itself (taskable_type='Project', taskable_id=project_id)
    - Deals belonging to the project
    - Leads linked to the project (via LeadProject)
    - Accounts linked to the project (via AccountProject)
    """
    async with async_get_session() as session:
        # Base query with eager loading
        base_query = (
            select(Task)
            .options(
                selectinload(Task.current_status),
                selectinload(Task.priority),
            )
        )

        # Apply search filter on task title only
        if search and search.strip():
            search_lower = search.strip().lower()
            base_query = base_query.where(func.lower(Task.title).contains(search_lower))

        if project_id:
            # Get Deal IDs for this project
            deal_stmt = select(Deal.id).where(Deal.project_id == project_id)
            deal_result = await session.execute(deal_stmt)
            deal_ids = [row[0] for row in deal_result.fetchall()]

            # Get Lead IDs linked to this project
            lead_stmt = select(LeadProject.lead_id).where(LeadProject.project_id == project_id)
            lead_result = await session.execute(lead_stmt)
            lead_ids = [row[0] for row in lead_result.fetchall()]

            # Get Account IDs linked to this project
            account_stmt = select(AccountProject.account_id).where(AccountProject.project_id == project_id)
            account_result = await session.execute(account_stmt)
            account_ids = [row[0] for row in account_result.fetchall()]

            # Build conditions for project scope
            conditions = [
                and_(Task.taskable_type == "Project", Task.taskable_id == project_id),
            ]
            if deal_ids:
                conditions.append(
                    and_(Task.taskable_type == "Deal", Task.taskable_id.in_(deal_ids))
                )
            if lead_ids:
                conditions.append(
                    and_(Task.taskable_type == "Lead", Task.taskable_id.in_(lead_ids))
                )
            if account_ids:
                conditions.append(
                    and_(Task.taskable_type == "Account", Task.taskable_id.in_(account_ids))
                )

            base_query = base_query.where(or_(*conditions))

        # Apply ordering
        order_column = getattr(Task, order_by, Task.created_at)
        order_func = order_column.desc() if order.upper() == "DESC" else order_column.asc()
        base_query = base_query.order_by(order_func)

        # Get total count
        count_query = select(func.count()).select_from(base_query.subquery())
        count_result = await session.execute(count_query)
        total = count_result.scalar() or 0

        # Apply pagination
        offset = (page - 1) * page_size
        base_query = base_query.offset(offset).limit(page_size)

        result = await session.execute(base_query)
        tasks = list(result.scalars().all())

        return tasks, total


async def find_task_by_id(task_id: int) -> Optional[Task]:
    """Find a task by ID."""
    async with async_get_session() as session:
        stmt = (
            select(Task)
            .options(
                selectinload(Task.current_status),
                selectinload(Task.priority),
            )
            .where(Task.id == task_id)
        )
        result = await session.execute(stmt)
        return result.scalar_one_or_none()


async def create_task(data: Dict[str, Any]) -> Task:
    """Create a new task."""
    async with async_get_session() as session:
        try:
            task = Task(**data)
            session.add(task)
            await session.commit()
            await session.refresh(task, ["current_status", "priority", "created_at", "updated_at"])
            return task
        except Exception as e:
            await session.rollback()
            logger.error(f"Failed to create task: {e}")
            raise


async def update_task(task_id: int, data: Dict[str, Any]) -> Optional[Task]:
    """Update a task."""
    async with async_get_session() as session:
        try:
            stmt = select(Task).where(Task.id == task_id)
            result = await session.execute(stmt)
            task = result.scalar_one_or_none()
            if not task:
                return None
            for key, value in data.items():
                if hasattr(task, key):
                    setattr(task, key, value)
            await session.commit()
            await session.refresh(task, ["current_status", "priority", "created_at", "updated_at"])
            return task
        except Exception as e:
            await session.rollback()
            logger.error(f"Failed to update task: {e}")
            raise


async def delete_task(task_id: int) -> bool:
    """Delete a task."""
    async with async_get_session() as session:
        try:
            stmt = select(Task).where(Task.id == task_id)
            result = await session.execute(stmt)
            task = result.scalar_one_or_none()
            if not task:
                return False
            await session.delete(task)
            await session.commit()
            return True
        except Exception as e:
            await session.rollback()
            logger.error(f"Failed to delete task: {e}")
            raise


async def find_tasks_by_entity(entity_type: str, entity_id: int) -> List[Task]:
    """Find all tasks for a specific entity."""
    async with async_get_session() as session:
        stmt = (
            select(Task)
            .options(
                selectinload(Task.current_status),
                selectinload(Task.priority),
            )
            .where(Task.taskable_type == entity_type)
            .where(Task.taskable_id == entity_id)
            .order_by(Task.created_at.desc())
        )
        result = await session.execute(stmt)
        return list(result.scalars().all())


async def get_lead_display_name(session, lead_id: int) -> Optional[str]:
    """Get display name for a lead from its subtype table."""
    # Try Job first
    result = await session.execute(select(Job.job_title).where(Job.id == lead_id))
    name = result.scalar_one_or_none()
    if name:
        return name
    # Try Opportunity
    result = await session.execute(select(Opportunity.opportunity_name).where(Opportunity.id == lead_id))
    name = result.scalar_one_or_none()
    if name:
        return name
    # Try Partnership
    result = await session.execute(select(Partnership.partnership_name).where(Partnership.id == lead_id))
    return result.scalar_one_or_none()


async def get_entity_display_name(entity_type: str, entity_id: int) -> Optional[str]:
    """Get display name for an entity."""
    if entity_type not in ("Project", "Deal", "Lead", "Account"):
        return None

    async with async_get_session() as session:
        if entity_type == "Lead":
            return await get_lead_display_name(session, entity_id)

        stmt_map = {
            "Project": select(Project.name).where(Project.id == entity_id),
            "Deal": select(Deal.name).where(Deal.id == entity_id),
            "Account": select(Account.name).where(Account.id == entity_id),
        }
        result = await session.execute(stmt_map[entity_type])
        return result.scalar_one_or_none()


async def get_lead_type(lead_id: int) -> Optional[str]:
    """Get the type of a lead (Job, Opportunity, Partnership)."""
    async with async_get_session() as session:
        result = await session.execute(
            select(Lead.type).where(Lead.id == lead_id)
        )
        return result.scalar_one_or_none()


async def get_account_entity(account_id: int) -> Optional[tuple[str, int]]:
    """Get the Individual or Organization linked to an Account. Returns (type, id) or None."""
    async with async_get_session() as session:
        # Check if there's an individual linked to this account
        ind_result = await session.execute(
            select(Individual.id).where(Individual.account_id == account_id).limit(1)
        )
        ind_id = ind_result.scalar_one_or_none()
        if ind_id:
            return ("Individual", ind_id)

        # Check if there's an organization linked to this account
        org_result = await session.execute(
            select(Organization.id).where(Organization.account_id == account_id).limit(1)
        )
        org_id = org_result.scalar_one_or_none()
        if org_id:
            return ("Organization", org_id)

        return None


async def find_task_statuses() -> List[Status]:
    """Find all task statuses."""
    async with async_get_session() as session:
        stmt = (
            select(Status)
            .where(Status.category == "task_status")
            .order_by(Status.id)
        )
        result = await session.execute(stmt)
        return list(result.scalars().all())


async def find_task_priorities() -> List[Status]:
    """Find all task priorities."""
    async with async_get_session() as session:
        stmt = (
            select(Status)
            .where(Status.category == "task_priority")
            .order_by(Status.id)
        )
        result = await session.execute(stmt)
        return list(result.scalars().all())


async def batch_get_entity_display_names(entities: List[Tuple[str, int]]) -> Dict[Tuple[str, int], Optional[str]]:
    """Batch fetch display names for multiple entities. Returns dict mapping (type, id) -> name."""
    if not entities:
        return {}
    
    async with async_get_session() as session:
        result_map: Dict[Tuple[str, int], Optional[str]] = {}
        
        # Group by entity type
        by_type: Dict[str, List[int]] = {}
        for entity_type, entity_id in entities:
            if entity_type not in ("Project", "Deal", "Lead", "Account"):
                result_map[(entity_type, entity_id)] = None
                continue
            if entity_type not in by_type:
                by_type[entity_type] = []
            by_type[entity_type].append(entity_id)
        
        # Batch fetch by type
        if "Project" in by_type:
            stmt = select(Project.id, Project.name).where(Project.id.in_(by_type["Project"]))
            result = await session.execute(stmt)
            for row in result.fetchall():
                result_map[("Project", row[0])] = row[1]
        
        if "Deal" in by_type:
            stmt = select(Deal.id, Deal.name).where(Deal.id.in_(by_type["Deal"]))
            result = await session.execute(stmt)
            for row in result.fetchall():
                result_map[("Deal", row[0])] = row[1]
        
        if "Lead" in by_type:
            lead_ids = by_type["Lead"]
            # Get names from Job
            result = await session.execute(select(Job.id, Job.job_title).where(Job.id.in_(lead_ids)))
            for row in result.fetchall():
                result_map[("Lead", row[0])] = row[1]
            # Get names from Opportunity
            result = await session.execute(select(Opportunity.id, Opportunity.opportunity_name).where(Opportunity.id.in_(lead_ids)))
            for row in result.fetchall():
                result_map[("Lead", row[0])] = row[1]
            # Get names from Partnership
            result = await session.execute(select(Partnership.id, Partnership.partnership_name).where(Partnership.id.in_(lead_ids)))
            for row in result.fetchall():
                result_map[("Lead", row[0])] = row[1]
        
        if "Account" in by_type:
            stmt = select(Account.id, Account.name).where(Account.id.in_(by_type["Account"]))
            result = await session.execute(stmt)
            for row in result.fetchall():
                result_map[("Account", row[0])] = row[1]
        
        # Set None for any entities not found
        for entity_type, entity_id in entities:
            if (entity_type, entity_id) not in result_map:
                result_map[(entity_type, entity_id)] = None
        
        return result_map


async def batch_get_lead_types(lead_ids: List[int]) -> Dict[int, Optional[str]]:
    """Batch fetch lead types for multiple leads. Returns dict mapping lead_id -> type."""
    if not lead_ids:
        return {}
    
    async with async_get_session() as session:
        stmt = select(Lead.id, Lead.type).where(Lead.id.in_(lead_ids))
        result = await session.execute(stmt)
        return {row[0]: row[1] for row in result.fetchall()}


async def batch_get_account_entities(account_ids: List[int]) -> Dict[int, Optional[Tuple[str, int]]]:
    """Batch fetch Individual/Organization for multiple accounts. Returns dict mapping account_id -> (type, id) or None."""
    if not account_ids:
        return {}
    
    async with async_get_session() as session:
        result_map: Dict[int, Optional[Tuple[str, int]]] = {}
        
        # Batch fetch individuals
        ind_stmt = select(Individual.id, Individual.account_id).where(Individual.account_id.in_(account_ids))
        ind_result = await session.execute(ind_stmt)
        for row in ind_result.fetchall():
            result_map[row[1]] = ("Individual", row[0])
        
        # Batch fetch organizations (only for accounts not already found as individuals)
        remaining_account_ids = [aid for aid in account_ids if aid not in result_map]
        if remaining_account_ids:
            org_stmt = select(Organization.id, Organization.account_id).where(Organization.account_id.in_(remaining_account_ids))
            org_result = await session.execute(org_stmt)
            for row in org_result.fetchall():
                result_map[row[1]] = ("Organization", row[0])
        
        # Set None for accounts not found
        for account_id in account_ids:
            if account_id not in result_map:
                result_map[account_id] = None
        
        return result_map
