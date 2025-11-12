"""Project model."""

from datetime import datetime, date
from typing import TypedDict, Optional
from sqlalchemy import Column, Text, DateTime, BigInteger, Date, ForeignKey, func, Index
from sqlalchemy.orm import relationship

from models import Base


class Project(Base):
    """Project SQLAlchemy model."""

    __tablename__ = "Project"

    id = Column(BigInteger, primary_key=True, autoincrement=True)
    tenant_id = Column(BigInteger, ForeignKey("Tenant.id", ondelete="CASCADE"), nullable=True)
    name = Column(Text, nullable=False)
    description = Column(Text, nullable=True)
    owner_user_id = Column(BigInteger, nullable=True)
    status = Column(Text, nullable=True)
    start_date = Column(Date, nullable=True)
    end_date = Column(Date, nullable=True)
    created_at = Column(DateTime(timezone=True), server_default=func.now(), nullable=False)
    updated_at = Column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now(), nullable=False)

    # Relationships
    tenant = relationship("Tenant", back_populates="projects")

    __table_args__ = (
        Index('idx_project_tenant_id', 'tenant_id'),
    )


# Functional data models (TypedDict)
class ProjectData(TypedDict, total=False):
    """Functional project data model."""
    id: int
    tenant_id: Optional[int]
    tenant: Optional[dict]  # Nested TenantData when loaded
    name: str
    description: Optional[str]
    owner_user_id: Optional[int]
    status: Optional[str]
    start_date: Optional[date]
    end_date: Optional[date]
    created_at: Optional[datetime]
    updated_at: Optional[datetime]


class ProjectCreateData(TypedDict, total=False):
    """Functional project creation data model."""
    tenant_id: Optional[int]
    name: str
    description: Optional[str]
    owner_user_id: Optional[int]
    status: Optional[str]
    start_date: Optional[date]
    end_date: Optional[date]


class ProjectUpdateData(TypedDict, total=False):
    """Functional project update data model."""
    tenant_id: Optional[int]
    name: Optional[str]
    description: Optional[str]
    owner_user_id: Optional[int]
    status: Optional[str]
    start_date: Optional[date]
    end_date: Optional[date]


# Conversion functions
def orm_to_dict(project: Project) -> ProjectData:
    """Convert SQLAlchemy model to functional dict."""
    from models.Tenant import orm_to_dict as tenant_to_dict

    result = ProjectData(
        id=project.id,
        tenant_id=project.tenant_id,
        name=project.name or "",
        description=project.description,
        owner_user_id=project.owner_user_id,
        status=project.status,
        start_date=project.start_date,
        end_date=project.end_date,
        created_at=project.created_at,
        updated_at=project.updated_at
    )

    if project.tenant:
        result["tenant"] = tenant_to_dict(project.tenant)

    return result


def dict_to_orm(data: ProjectCreateData) -> Project:
    """Convert functional dict to SQLAlchemy model."""
    return Project(
        tenant_id=data.get("tenant_id"),
        name=data.get("name", ""),
        description=data.get("description"),
        owner_user_id=data.get("owner_user_id"),
        status=data.get("status"),
        start_date=data.get("start_date"),
        end_date=data.get("end_date")
    )


def update_orm_from_dict(project: Project, data: ProjectUpdateData) -> Project:
    """Update SQLAlchemy model from functional dict."""
    if "tenant_id" in data:
        project.tenant_id = data["tenant_id"]
    if "name" in data and data["name"] is not None:
        project.name = data["name"]
    if "description" in data:
        project.description = data["description"]
    if "owner_user_id" in data:
        project.owner_user_id = data["owner_user_id"]
    if "status" in data:
        project.status = data["status"]
    if "start_date" in data:
        project.start_date = data["start_date"]
    if "end_date" in data:
        project.end_date = data["end_date"]
    return project
