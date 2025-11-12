"""Deal model."""

from datetime import datetime, date
from typing import TypedDict, Optional
from decimal import Decimal
from sqlalchemy import Column, Text, DateTime, BigInteger, Numeric, Date, ForeignKey, func, Index
from sqlalchemy.orm import relationship

from models import Base


class Deal(Base):
    """Deal SQLAlchemy model."""

    __tablename__ = "Deal"

    id = Column(BigInteger, primary_key=True, autoincrement=True)
    tenant_id = Column(BigInteger, ForeignKey("Tenant.id", ondelete="CASCADE"), nullable=True)
    name = Column(Text, nullable=False)
    description = Column(Text, nullable=True)
    owner_user_id = Column(BigInteger, nullable=True)
    current_status_id = Column(BigInteger, ForeignKey("Status.id", ondelete="SET NULL"), nullable=True)
    value_amount = Column(Numeric(18, 2), nullable=True)
    value_currency = Column(Text, default="USD")
    probability = Column(Numeric(5, 2), nullable=True)
    expected_close_date = Column(Date, nullable=True)
    created_at = Column(DateTime(timezone=True), server_default=func.now(), nullable=False)
    updated_at = Column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now(), nullable=False)

    # Relationships
    tenant = relationship("Tenant", back_populates="deals")
    current_status = relationship("Status", back_populates="deals")
    leads = relationship("Lead", back_populates="deal")

    __table_args__ = (
        Index('idx_deal_tenant_id', 'tenant_id'),
    )


# Functional data models (TypedDict)
class DealData(TypedDict, total=False):
    """Functional deal data model."""
    id: int
    tenant_id: Optional[int]
    tenant: Optional[dict]  # Nested TenantData when loaded
    name: str
    description: Optional[str]
    owner_user_id: Optional[int]
    current_status_id: Optional[int]
    current_status: Optional[dict]  # Nested StatusData when loaded
    value_amount: Optional[Decimal]
    value_currency: str
    probability: Optional[Decimal]
    expected_close_date: Optional[date]
    created_at: Optional[datetime]
    updated_at: Optional[datetime]


class DealCreateData(TypedDict, total=False):
    """Functional deal creation data model."""
    tenant_id: Optional[int]
    name: str
    description: Optional[str]
    owner_user_id: Optional[int]
    current_status_id: Optional[int]
    value_amount: Optional[Decimal]
    value_currency: str
    probability: Optional[Decimal]
    expected_close_date: Optional[date]


class DealUpdateData(TypedDict, total=False):
    """Functional deal update data model."""
    tenant_id: Optional[int]
    name: Optional[str]
    description: Optional[str]
    owner_user_id: Optional[int]
    current_status_id: Optional[int]
    value_amount: Optional[Decimal]
    value_currency: Optional[str]
    probability: Optional[Decimal]
    expected_close_date: Optional[date]


# Conversion functions
def orm_to_dict(deal: Deal) -> DealData:
    """Convert SQLAlchemy model to functional dict."""
    from models.Status import orm_to_dict as status_to_dict
    from models.Tenant import orm_to_dict as tenant_to_dict

    result = DealData(
        id=deal.id,
        tenant_id=deal.tenant_id,
        name=deal.name or "",
        description=deal.description,
        owner_user_id=deal.owner_user_id,
        current_status_id=deal.current_status_id,
        value_amount=deal.value_amount,
        value_currency=deal.value_currency or "USD",
        probability=deal.probability,
        expected_close_date=deal.expected_close_date,
        created_at=deal.created_at,
        updated_at=deal.updated_at
    )

    if deal.tenant:
        result["tenant"] = tenant_to_dict(deal.tenant)
    if deal.current_status:
        result["current_status"] = status_to_dict(deal.current_status)

    return result


def dict_to_orm(data: DealCreateData) -> Deal:
    """Convert functional dict to SQLAlchemy model."""
    return Deal(
        tenant_id=data.get("tenant_id"),
        name=data.get("name", ""),
        description=data.get("description"),
        owner_user_id=data.get("owner_user_id"),
        current_status_id=data.get("current_status_id"),
        value_amount=data.get("value_amount"),
        value_currency=data.get("value_currency", "USD"),
        probability=data.get("probability"),
        expected_close_date=data.get("expected_close_date")
    )


def update_orm_from_dict(deal: Deal, data: DealUpdateData) -> Deal:
    """Update SQLAlchemy model from functional dict."""
    if "tenant_id" in data:
        deal.tenant_id = data["tenant_id"]
    if "name" in data and data["name"] is not None:
        deal.name = data["name"]
    if "description" in data:
        deal.description = data["description"]
    if "owner_user_id" in data:
        deal.owner_user_id = data["owner_user_id"]
    if "current_status_id" in data:
        deal.current_status_id = data["current_status_id"]
    if "value_amount" in data:
        deal.value_amount = data["value_amount"]
    if "value_currency" in data and data["value_currency"] is not None:
        deal.value_currency = data["value_currency"]
    if "probability" in data:
        deal.probability = data["probability"]
    if "expected_close_date" in data:
        deal.expected_close_date = data["expected_close_date"]
    return deal
