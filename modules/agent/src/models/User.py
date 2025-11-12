"""User model for multi-tenant users."""

from datetime import datetime
from typing import TypedDict, Optional
from sqlalchemy import Column, Text, DateTime, BigInteger, ForeignKey, func, Index, UniqueConstraint, CheckConstraint
from sqlalchemy.orm import relationship

from models import Base


class User(Base):
    """User SQLAlchemy model (belongs to one Tenant)."""

    __tablename__ = "User"

    id = Column(BigInteger, primary_key=True, autoincrement=True)
    tenant_id = Column(BigInteger, ForeignKey("Tenant.id", ondelete="CASCADE"), nullable=False)
    email = Column(Text, nullable=False)
    first_name = Column(Text, nullable=True)
    last_name = Column(Text, nullable=True)
    avatar_url = Column(Text, nullable=True)
    status = Column(Text, nullable=False, default='active')
    last_login_at = Column(DateTime(timezone=True), nullable=True)
    created_at = Column(DateTime(timezone=True), server_default=func.now(), nullable=False)
    updated_at = Column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now(), nullable=False)

    # Relationships
    tenant = relationship("Tenant", back_populates="users")
    user_roles = relationship("UserRole", back_populates="user")

    __table_args__ = (
        UniqueConstraint('tenant_id', 'email', name='user_tenant_email_unique'),
        Index('idx_user_tenant_id', 'tenant_id'),
        Index('idx_user_email', 'email'),
        CheckConstraint("status IN ('active', 'inactive', 'invited')", name='user_status_check'),
    )


# Functional data models (TypedDict)
class UserData(TypedDict, total=False):
    """Functional user data model."""
    id: int
    tenant_id: int
    tenant: Optional[dict]  # Nested TenantData when loaded
    email: str
    first_name: Optional[str]
    last_name: Optional[str]
    avatar_url: Optional[str]
    status: str
    last_login_at: Optional[datetime]
    created_at: Optional[datetime]
    updated_at: Optional[datetime]


class UserCreateData(TypedDict, total=False):
    """Functional user creation data model."""
    tenant_id: int
    email: str
    first_name: Optional[str]
    last_name: Optional[str]
    avatar_url: Optional[str]
    status: str


class UserUpdateData(TypedDict, total=False):
    """Functional user update data model."""
    email: Optional[str]
    first_name: Optional[str]
    last_name: Optional[str]
    avatar_url: Optional[str]
    status: Optional[str]
    last_login_at: Optional[datetime]


# Conversion functions
def orm_to_dict(user: User) -> UserData:
    """Convert SQLAlchemy model to functional dict."""
    from models.Tenant import orm_to_dict as tenant_to_dict

    result = UserData(
        id=user.id,
        tenant_id=user.tenant_id,
        email=user.email,
        first_name=user.first_name,
        last_name=user.last_name,
        avatar_url=user.avatar_url,
        status=user.status or 'active',
        last_login_at=user.last_login_at,
        created_at=user.created_at,
        updated_at=user.updated_at
    )

    if user.tenant:
        result["tenant"] = tenant_to_dict(user.tenant)

    return result


def dict_to_orm(data: UserCreateData) -> User:
    """Convert functional dict to SQLAlchemy model."""
    return User(
        tenant_id=data.get("tenant_id"),
        email=data.get("email", ""),
        first_name=data.get("first_name"),
        last_name=data.get("last_name"),
        avatar_url=data.get("avatar_url"),
        status=data.get("status", "active")
    )


def update_orm_from_dict(user: User, data: UserUpdateData) -> User:
    """Update SQLAlchemy model from functional dict."""
    if "email" in data and data["email"] is not None:
        user.email = data["email"]
    if "first_name" in data:
        user.first_name = data["first_name"]
    if "last_name" in data:
        user.last_name = data["last_name"]
    if "avatar_url" in data:
        user.avatar_url = data["avatar_url"]
    if "status" in data and data["status"] is not None:
        user.status = data["status"]
    if "last_login_at" in data:
        user.last_login_at = data["last_login_at"]
    return user
