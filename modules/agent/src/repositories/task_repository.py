"""Repository layer for task data access."""

import logging
from typing import List, Optional, Dict, Any, Tuple
from sqlalchemy import select, and_, or_, func
from sqlalchemy.orm import selectinload
from models.Task import Task
from models.Project import Project
from models.Deal import Deal
from models.Lead import Lead
from models.Account import Account
from models.LeadProject import LeadProject
from models.AccountProject import AccountProject
from models.Status import Status
from lib.db import async_get_session

logger = logging.getLogger(__name__)


async def find_all_tasks(
    tenant_id: int,
    project_id: Optional[int] = None,
    page: int = 1,
    page_size: int = 20,
    order_by: str = "created_at",
    order: str = "DESC",
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
            await session.refresh(task, ["current_status", "priority"])
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
            await session.refresh(task, ["current_status", "priority"])
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


async def get_entity_display_name(entity_type: str, entity_id: int) -> Optional[str]:
    """Get display name for an entity."""
    if entity_type not in ("Project", "Deal", "Lead", "Account"):
        return None

    async with async_get_session() as session:
        stmt_map = {
            "Project": select(Project.name).where(Project.id == entity_id),
            "Deal": select(Deal.name).where(Deal.id == entity_id),
            "Lead": select(Lead.title).where(Lead.id == entity_id),
            "Account": select(Account.name).where(Account.id == entity_id),
        }
        result = await session.execute(stmt_map[entity_type])
        return result.scalar_one_or_none()


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
