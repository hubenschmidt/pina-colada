"""Role model for system-defined roles."""

from datetime import datetime
from typing import TypedDict, Optional
from sqlalchemy import Column, Text, DateTime, BigInteger, func
from sqlalchemy.orm import relationship

from models import Base


class Role(Base):
    """Role SQLAlchemy model (system-defined roles)."""

    __tablename__ = "Role"

    id = Column(BigInteger, primary_key=True, autoincrement=True)
    name = Column(Text, nullable=False, unique=True)
    description = Column(Text, nullable=True)
    created_at = Column(DateTime(timezone=True), server_default=func.now(), nullable=False)

    # Relationships
    user_roles = relationship("UserRole", back_populates="role")


# Functional data models (TypedDict)
class RoleData(TypedDict, total=False):
    """Functional role data model."""
    id: int
    name: str
    description: Optional[str]
    created_at: Optional[datetime]


class RoleCreateData(TypedDict, total=False):
    """Functional role creation data model."""
    name: str
    description: Optional[str]


class RoleUpdateData(TypedDict, total=False):
    """Functional role update data model."""
    name: Optional[str]
    description: Optional[str]


# Conversion functions
def orm_to_dict(role: Role) -> RoleData:
    """Convert SQLAlchemy model to functional dict."""
    return RoleData(
        id=role.id,
        name=role.name,
        description=role.description,
        created_at=role.created_at
    )


def dict_to_orm(data: RoleCreateData) -> Role:
    """Convert functional dict to SQLAlchemy model."""
    return Role(
        name=data.get("name", ""),
        description=data.get("description")
    )


def update_orm_from_dict(role: Role, data: RoleUpdateData) -> Role:
    """Update SQLAlchemy model from functional dict."""
    if "name" in data and data["name"] is not None:
        role.name = data["name"]
    if "description" in data:
        role.description = data["description"]
    return role
