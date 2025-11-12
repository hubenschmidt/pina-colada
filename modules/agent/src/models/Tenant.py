"""Tenant model for multi-tenant isolation."""

from datetime import datetime
from typing import TypedDict, Optional
from sqlalchemy import Column, Text, DateTime, BigInteger, func, CheckConstraint
from sqlalchemy.dialects.postgresql import JSONB
from sqlalchemy.orm import relationship

from models import Base


class Tenant(Base):
    """Tenant SQLAlchemy model (app customer/company)."""

    __tablename__ = "Tenant"

    id = Column(BigInteger, primary_key=True, autoincrement=True)
    name = Column(Text, nullable=False)
    slug = Column(Text, nullable=False, unique=True)
    status = Column(Text, nullable=False, default='active')
    plan = Column(Text, nullable=False, default='free')
    settings = Column(JSONB, nullable=False, default={}, server_default='{}')
    created_at = Column(DateTime(timezone=True), server_default=func.now(), nullable=False)
    updated_at = Column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now(), nullable=False)

    # Relationships
    users = relationship("User", back_populates="tenant")
    deals = relationship("Deal", back_populates="tenant")
    organizations = relationship("Organization", back_populates="tenant")
    individuals = relationship("Individual", back_populates="tenant")
    projects = relationship("Project", back_populates="tenant")

    __table_args__ = (
        CheckConstraint("status IN ('active', 'suspended', 'trial', 'cancelled')", name='tenant_status_check'),
        CheckConstraint("plan IN ('free', 'starter', 'professional', 'enterprise')", name='tenant_plan_check'),
    )


# Functional data models (TypedDict)
class TenantData(TypedDict, total=False):
    """Functional tenant data model."""
    id: int
    name: str
    slug: str
    status: str
    plan: str
    settings: dict
    created_at: Optional[datetime]
    updated_at: Optional[datetime]


class TenantCreateData(TypedDict, total=False):
    """Functional tenant creation data model."""
    name: str
    slug: str
    status: str
    plan: str
    settings: dict


class TenantUpdateData(TypedDict, total=False):
    """Functional tenant update data model."""
    name: Optional[str]
    slug: Optional[str]
    status: Optional[str]
    plan: Optional[str]
    settings: Optional[dict]


# Conversion functions
def orm_to_dict(tenant: Tenant) -> TenantData:
    """Convert SQLAlchemy model to functional dict."""
    return TenantData(
        id=tenant.id,
        name=tenant.name,
        slug=tenant.slug,
        status=tenant.status or 'active',
        plan=tenant.plan or 'free',
        settings=tenant.settings or {},
        created_at=tenant.created_at,
        updated_at=tenant.updated_at
    )


def dict_to_orm(data: TenantCreateData) -> Tenant:
    """Convert functional dict to SQLAlchemy model."""
    return Tenant(
        name=data.get("name", ""),
        slug=data.get("slug", ""),
        status=data.get("status", "active"),
        plan=data.get("plan", "free"),
        settings=data.get("settings", {})
    )


def update_orm_from_dict(tenant: Tenant, data: TenantUpdateData) -> Tenant:
    """Update SQLAlchemy model from functional dict."""
    if "name" in data and data["name"] is not None:
        tenant.name = data["name"]
    if "slug" in data and data["slug"] is not None:
        tenant.slug = data["slug"]
    if "status" in data and data["status"] is not None:
        tenant.status = data["status"]
    if "plan" in data and data["plan"] is not None:
        tenant.plan = data["plan"]
    if "settings" in data and data["settings"] is not None:
        tenant.settings = data["settings"]
    return tenant
