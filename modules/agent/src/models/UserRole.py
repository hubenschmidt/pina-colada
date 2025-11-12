"""UserRole model for user-role junction table."""

from datetime import datetime
from typing import TypedDict, Optional
from sqlalchemy import Column, DateTime, BigInteger, ForeignKey, func, PrimaryKeyConstraint
from sqlalchemy.orm import relationship

from models import Base


class UserRole(Base):
    """UserRole SQLAlchemy model (many-to-many junction)."""

    __tablename__ = "UserRole"

    user_id = Column(BigInteger, ForeignKey("User.id", ondelete="CASCADE"), nullable=False)
    role_id = Column(BigInteger, ForeignKey("Role.id", ondelete="CASCADE"), nullable=False)
    created_at = Column(DateTime(timezone=True), server_default=func.now(), nullable=False)

    # Relationships
    user = relationship("User", back_populates="user_roles")
    role = relationship("Role", back_populates="user_roles")

    __table_args__ = (
        PrimaryKeyConstraint('user_id', 'role_id'),
    )


# Functional data models (TypedDict)
class UserRoleData(TypedDict, total=False):
    """Functional user-role data model."""
    user_id: int
    user: Optional[dict]  # Nested UserData when loaded
    role_id: int
    role: Optional[dict]  # Nested RoleData when loaded
    created_at: Optional[datetime]


class UserRoleCreateData(TypedDict, total=False):
    """Functional user-role creation data model."""
    user_id: int
    role_id: int


# No update model needed for junction table (delete and recreate instead)


# Conversion functions
def orm_to_dict(user_role: UserRole) -> UserRoleData:
    """Convert SQLAlchemy model to functional dict."""
    from models.User import orm_to_dict as user_to_dict
    from models.Role import orm_to_dict as role_to_dict

    result = UserRoleData(
        user_id=user_role.user_id,
        role_id=user_role.role_id,
        created_at=user_role.created_at
    )

    if user_role.user:
        result["user"] = user_to_dict(user_role.user)
    if user_role.role:
        result["role"] = role_to_dict(user_role.role)

    return result


def dict_to_orm(data: UserRoleCreateData) -> UserRole:
    """Convert functional dict to SQLAlchemy model."""
    return UserRole(
        user_id=data.get("user_id"),
        role_id=data.get("role_id")
    )
