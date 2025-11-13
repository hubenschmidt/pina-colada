"""Data models for tasks.

SQLAlchemy model for database persistence (unavoidable OOP requirement).
Functional TypedDict models for business logic.
"""

from datetime import datetime, date
from typing import TypedDict, Optional
from sqlalchemy import Column, Text, DateTime, BigInteger, ForeignKey, Date, func
from sqlalchemy.orm import relationship

from models import Base


# SQLAlchemy model (OOP required for ORM)
class Task(Base):
    """Task SQLAlchemy model."""

    __tablename__ = "Task"

    id = Column(BigInteger, primary_key=True, autoincrement=True)
    taskable_type = Column(Text, nullable=True)  # 'Deal', 'Lead', 'Job', 'Project', 'Organization', 'Individual', 'Contact'
    taskable_id = Column(BigInteger, nullable=True)
    title = Column(Text, nullable=False)
    description = Column(Text, nullable=True)
    current_status_id = Column(BigInteger, ForeignKey("Status.id", ondelete="SET NULL"), nullable=True)
    priority_id = Column(BigInteger, ForeignKey("Status.id", ondelete="SET NULL"), nullable=True)
    due_date = Column(Date, nullable=True)
    completed_at = Column(DateTime(timezone=True), nullable=True)
    assigned_to_user_id = Column(BigInteger, nullable=True)
    created_at = Column(DateTime(timezone=True), server_default=func.now(), nullable=False)
    updated_at = Column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now(), nullable=False)

    # Relationships
    current_status = relationship("Status", back_populates="tasks_status", foreign_keys=[current_status_id])
    priority = relationship("Status", back_populates="tasks_priority", foreign_keys=[priority_id])


# Functional data models (TypedDict)
class TaskData(TypedDict, total=False):
    """Functional task data model for business logic."""
    id: int
    taskable_type: Optional[str]
    taskable_id: Optional[int]
    title: str
    description: Optional[str]
    current_status_id: Optional[int]
    priority_id: Optional[int]
    due_date: Optional[date]
    completed_at: Optional[datetime]
    assigned_to_user_id: Optional[int]
    created_at: Optional[datetime]
    updated_at: Optional[datetime]
    # Nested relationships when loaded
    current_status: Optional[dict]
    priority: Optional[dict]


class TaskCreateData(TypedDict, total=False):
    """Functional task creation data model."""
    taskable_type: Optional[str]
    taskable_id: Optional[int]
    title: str
    description: Optional[str]
    current_status_id: Optional[int]
    priority_id: Optional[int]
    due_date: Optional[date]
    completed_at: Optional[datetime]
    assigned_to_user_id: Optional[int]


class TaskUpdateData(TypedDict, total=False):
    """Functional task update data model."""
    taskable_type: Optional[str]
    taskable_id: Optional[int]
    title: Optional[str]
    description: Optional[str]
    current_status_id: Optional[int]
    priority_id: Optional[int]
    due_date: Optional[date]
    completed_at: Optional[datetime]
    assigned_to_user_id: Optional[int]


# Conversion functions
def orm_to_dict(task: Task) -> TaskData:
    """Convert SQLAlchemy model to functional dict."""
    from models.Status import orm_to_dict as status_to_dict

    return TaskData(
        id=task.id,
        taskable_type=task.taskable_type,
        taskable_id=task.taskable_id,
        title=task.title or "",
        description=task.description,
        current_status_id=task.current_status_id,
        priority_id=task.priority_id,
        due_date=task.due_date,
        completed_at=task.completed_at,
        assigned_to_user_id=task.assigned_to_user_id,
        created_at=task.created_at,
        updated_at=task.updated_at,
        current_status=status_to_dict(task.current_status) if task.current_status else None,
        priority=status_to_dict(task.priority) if task.priority else None,
    )


def dict_to_orm(data: TaskCreateData) -> Task:
    """Convert functional dict to SQLAlchemy model."""
    return Task(
        taskable_type=data.get("taskable_type"),
        taskable_id=data.get("taskable_id"),
        title=data.get("title", ""),
        description=data.get("description"),
        current_status_id=data.get("current_status_id"),
        priority_id=data.get("priority_id"),
        due_date=data.get("due_date"),
        completed_at=data.get("completed_at"),
        assigned_to_user_id=data.get("assigned_to_user_id"),
    )


def update_orm_from_dict(task: Task, data: TaskUpdateData) -> Task:
    """Update SQLAlchemy model from functional dict."""
    if "taskable_type" in data:
        task.taskable_type = data["taskable_type"]
    if "taskable_id" in data:
        task.taskable_id = data["taskable_id"]
    if "title" in data and data["title"] is not None:
        task.title = data["title"]
    if "description" in data:
        task.description = data["description"]
    if "current_status_id" in data:
        task.current_status_id = data["current_status_id"]
    if "priority_id" in data:
        task.priority_id = data["priority_id"]
    if "due_date" in data:
        task.due_date = data["due_date"]
    if "completed_at" in data:
        task.completed_at = data["completed_at"]
    if "assigned_to_user_id" in data:
        task.assigned_to_user_id = data["assigned_to_user_id"]
    return task
