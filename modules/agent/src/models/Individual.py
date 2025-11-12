"""Individual model for people."""

from datetime import datetime
from typing import TypedDict, Optional
from sqlalchemy import Column, Text, DateTime, BigInteger, ForeignKey, func, Index
from sqlalchemy.orm import relationship

from models import Base


class Individual(Base):
    """Individual SQLAlchemy model for people."""

    __tablename__ = "Individual"

    id = Column(BigInteger, primary_key=True, autoincrement=True)
    tenant_id = Column(BigInteger, ForeignKey("Tenant.id", ondelete="CASCADE"), nullable=True)
    first_name = Column(Text, nullable=False)
    last_name = Column(Text, nullable=False)
    email = Column(Text, nullable=True)
    phone = Column(Text, nullable=True)
    linkedin_url = Column(Text, nullable=True)
    title = Column(Text, nullable=True)
    notes = Column(Text, nullable=True)
    created_at = Column(DateTime(timezone=True), server_default=func.now(), nullable=False)
    updated_at = Column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now(), nullable=False)

    # Relationships
    tenant = relationship("Tenant", back_populates="individuals")
    contacts = relationship("Contact", back_populates="individual")

    __table_args__ = (
        Index('idx_individual_tenant_id', 'tenant_id'),
        Index('idx_individual_email_lower_tenant', 'tenant_id', func.lower(email), unique=True),
    )


# Functional data models (TypedDict)
class IndividualData(TypedDict, total=False):
    """Functional individual data model."""
    id: int
    tenant_id: Optional[int]
    tenant: Optional[dict]  # Nested TenantData when loaded
    first_name: str
    last_name: str
    email: Optional[str]
    phone: Optional[str]
    linkedin_url: Optional[str]
    title: Optional[str]
    notes: Optional[str]
    created_at: Optional[datetime]
    updated_at: Optional[datetime]


class IndividualCreateData(TypedDict, total=False):
    """Functional individual creation data model."""
    tenant_id: Optional[int]
    first_name: str
    last_name: str
    email: Optional[str]
    phone: Optional[str]
    linkedin_url: Optional[str]
    title: Optional[str]
    notes: Optional[str]


class IndividualUpdateData(TypedDict, total=False):
    """Functional individual update data model."""
    tenant_id: Optional[int]
    first_name: Optional[str]
    last_name: Optional[str]
    email: Optional[str]
    phone: Optional[str]
    linkedin_url: Optional[str]
    title: Optional[str]
    notes: Optional[str]


# Conversion functions
def orm_to_dict(individual: Individual) -> IndividualData:
    """Convert SQLAlchemy model to functional dict."""
    from models.Tenant import orm_to_dict as tenant_to_dict

    result = IndividualData(
        id=individual.id,
        tenant_id=individual.tenant_id,
        first_name=individual.first_name or "",
        last_name=individual.last_name or "",
        email=individual.email,
        phone=individual.phone,
        linkedin_url=individual.linkedin_url,
        title=individual.title,
        notes=individual.notes,
        created_at=individual.created_at,
        updated_at=individual.updated_at
    )

    if individual.tenant:
        result["tenant"] = tenant_to_dict(individual.tenant)

    return result


def dict_to_orm(data: IndividualCreateData) -> Individual:
    """Convert functional dict to SQLAlchemy model."""
    return Individual(
        tenant_id=data.get("tenant_id"),
        first_name=data.get("first_name", ""),
        last_name=data.get("last_name", ""),
        email=data.get("email"),
        phone=data.get("phone"),
        linkedin_url=data.get("linkedin_url"),
        title=data.get("title"),
        notes=data.get("notes")
    )


def update_orm_from_dict(individual: Individual, data: IndividualUpdateData) -> Individual:
    """Update SQLAlchemy model from functional dict."""
    if "tenant_id" in data:
        individual.tenant_id = data["tenant_id"]
    if "first_name" in data and data["first_name"] is not None:
        individual.first_name = data["first_name"]
    if "last_name" in data and data["last_name"] is not None:
        individual.last_name = data["last_name"]
    if "email" in data:
        individual.email = data["email"]
    if "phone" in data:
        individual.phone = data["phone"]
    if "linkedin_url" in data:
        individual.linkedin_url = data["linkedin_url"]
    if "title" in data:
        individual.title = data["title"]
    if "notes" in data:
        individual.notes = data["notes"]
    return individual
