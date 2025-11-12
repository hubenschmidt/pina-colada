"""Organization model for companies."""

from datetime import datetime
from typing import TypedDict, Optional
from sqlalchemy import Column, Text, Integer, DateTime, BigInteger, ForeignKey, func, Index
from sqlalchemy.orm import relationship

from models import Base


class Organization(Base):
    """Organization SQLAlchemy model for companies."""

    __tablename__ = "Organization"

    id = Column(BigInteger, primary_key=True, autoincrement=True)
    tenant_id = Column(BigInteger, ForeignKey("Tenant.id", ondelete="CASCADE"), nullable=True)
    name = Column(Text, nullable=False)
    website = Column(Text, nullable=True)
    phone = Column(Text, nullable=True)
    industry = Column(Text, nullable=True)
    employee_count = Column(Integer, nullable=True)
    description = Column(Text, nullable=True)
    notes = Column(Text, nullable=True)
    created_at = Column(DateTime(timezone=True), server_default=func.now(), nullable=False)
    updated_at = Column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now(), nullable=False)

    # Relationships
    tenant = relationship("Tenant", back_populates="organizations")
    jobs = relationship("Job", back_populates="organization")
    opportunities = relationship("Opportunity", back_populates="organization")
    partnerships = relationship("Partnership", back_populates="organization")
    contacts = relationship("Contact", back_populates="organization")

    __table_args__ = (
        Index('idx_organization_tenant_id', 'tenant_id'),
        Index('idx_organization_name_lower_tenant', 'tenant_id', func.lower(name), unique=True),
    )


# Functional data models (TypedDict)
class OrganizationData(TypedDict, total=False):
    """Functional organization data model."""
    id: int
    tenant_id: Optional[int]
    tenant: Optional[dict]  # Nested TenantData when loaded
    name: str
    website: Optional[str]
    phone: Optional[str]
    industry: Optional[str]
    employee_count: Optional[int]
    description: Optional[str]
    notes: Optional[str]
    created_at: Optional[datetime]
    updated_at: Optional[datetime]


class OrganizationCreateData(TypedDict, total=False):
    """Functional organization creation data model."""
    tenant_id: Optional[int]
    name: str
    website: Optional[str]
    phone: Optional[str]
    industry: Optional[str]
    employee_count: Optional[int]
    description: Optional[str]
    notes: Optional[str]


class OrganizationUpdateData(TypedDict, total=False):
    """Functional organization update data model."""
    tenant_id: Optional[int]
    name: Optional[str]
    website: Optional[str]
    phone: Optional[str]
    industry: Optional[str]
    employee_count: Optional[int]
    description: Optional[str]
    notes: Optional[str]


# Conversion functions
def orm_to_dict(org: Organization) -> OrganizationData:
    """Convert SQLAlchemy model to functional dict."""
    from models.Tenant import orm_to_dict as tenant_to_dict

    result = OrganizationData(
        id=org.id,
        tenant_id=org.tenant_id,
        name=org.name or "",
        website=org.website,
        phone=org.phone,
        industry=org.industry,
        employee_count=org.employee_count,
        description=org.description,
        notes=org.notes,
        created_at=org.created_at,
        updated_at=org.updated_at
    )

    if org.tenant:
        result["tenant"] = tenant_to_dict(org.tenant)

    return result


def dict_to_orm(data: OrganizationCreateData) -> Organization:
    """Convert functional dict to SQLAlchemy model."""
    return Organization(
        tenant_id=data.get("tenant_id"),
        name=data.get("name", ""),
        website=data.get("website"),
        phone=data.get("phone"),
        industry=data.get("industry"),
        employee_count=data.get("employee_count"),
        description=data.get("description"),
        notes=data.get("notes")
    )


def update_orm_from_dict(org: Organization, data: OrganizationUpdateData) -> Organization:
    """Update SQLAlchemy model from functional dict."""
    if "tenant_id" in data:
        org.tenant_id = data["tenant_id"]
    if "name" in data and data["name"] is not None:
        org.name = data["name"]
    if "website" in data:
        org.website = data["website"]
    if "phone" in data:
        org.phone = data["phone"]
    if "industry" in data:
        org.industry = data["industry"]
    if "employee_count" in data:
        org.employee_count = data["employee_count"]
    if "description" in data:
        org.description = data["description"]
    if "notes" in data:
        org.notes = data["notes"]
    return org
