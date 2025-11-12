"""Status model for central status/stage definitions."""

from datetime import datetime
from typing import TypedDict, Optional
from sqlalchemy import Column, String, Text, DateTime, Boolean, BigInteger, func
from sqlalchemy.orm import relationship

from models import Base


class Status(Base):
    """Status SQLAlchemy model for central status/stage definitions."""

    __tablename__ = "Status"

    id = Column(BigInteger, primary_key=True, autoincrement=True)
    name = Column(Text, nullable=False, unique=True)
    description = Column(Text, nullable=True)
    category = Column(Text, nullable=True)  # 'job', 'lead', 'deal', 'task_status', 'task_priority'
    is_terminal = Column(Boolean, default=False)
    created_at = Column(DateTime(timezone=True), server_default=func.now(), nullable=False)
    updated_at = Column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now(), nullable=False)

    # Relationships
    leads = relationship("Lead", back_populates="current_status", foreign_keys="Lead.current_status_id")
    deals = relationship("Deal", back_populates="current_status")
    tasks_status = relationship("Task", back_populates="current_status", foreign_keys="Task.current_status_id")
    tasks_priority = relationship("Task", back_populates="priority", foreign_keys="Task.priority_id")


# Functional data models (TypedDict)
class StatusData(TypedDict, total=False):
    """Functional status data model."""
    id: int
    name: str
    description: Optional[str]
    category: Optional[str]
    is_terminal: bool
    created_at: Optional[datetime]
    updated_at: Optional[datetime]


class StatusCreateData(TypedDict, total=False):
    """Functional status creation data model."""
    name: str
    description: Optional[str]
    category: Optional[str]
    is_terminal: bool


class StatusUpdateData(TypedDict, total=False):
    """Functional status update data model."""
    name: Optional[str]
    description: Optional[str]
    category: Optional[str]
    is_terminal: Optional[bool]


# Conversion functions
def orm_to_dict(status: Status) -> StatusData:
    """Convert SQLAlchemy model to functional dict."""
    return StatusData(
        id=status.id,
        name=status.name or "",
        description=status.description,
        category=status.category,
        is_terminal=status.is_terminal or False,
        created_at=status.created_at,
        updated_at=status.updated_at
    )


def dict_to_orm(data: StatusCreateData) -> Status:
    """Convert functional dict to SQLAlchemy model."""
    return Status(
        name=data.get("name", ""),
        description=data.get("description"),
        category=data.get("category"),
        is_terminal=data.get("is_terminal", False)
    )


def update_orm_from_dict(status: Status, data: StatusUpdateData) -> Status:
    """Update SQLAlchemy model from functional dict."""
    if "name" in data and data["name"] is not None:
        status.name = data["name"]
    if "description" in data:
        status.description = data["description"]
    if "category" in data:
        status.category = data["category"]
    if "is_terminal" in data and data["is_terminal"] is not None:
        status.is_terminal = data["is_terminal"]
    return status
