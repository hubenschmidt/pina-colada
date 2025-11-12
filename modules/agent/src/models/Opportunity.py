"""Opportunity model (extends Lead via Joined Table Inheritance)."""

from datetime import datetime, date
from typing import TypedDict, Optional
from decimal import Decimal
from sqlalchemy import Column, Text, DateTime, BigInteger, Numeric, Date, ForeignKey, func
from sqlalchemy.orm import relationship

from models import Base


class Opportunity(Base):
    """Opportunity SQLAlchemy model (extends Lead)."""

    __tablename__ = "Opportunity"

    id = Column(BigInteger, ForeignKey("Lead.id", ondelete="CASCADE"), primary_key=True)
    organization_id = Column(BigInteger, ForeignKey("Organization.id", ondelete="CASCADE"), nullable=False)
    opportunity_name = Column(Text, nullable=False)
    estimated_value = Column(Numeric(18, 2), nullable=True)
    probability = Column(Numeric(5, 2), nullable=True)
    expected_close_date = Column(Date, nullable=True)
    notes = Column(Text, nullable=True)
    created_at = Column(DateTime(timezone=True), server_default=func.now(), nullable=False)
    updated_at = Column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now(), nullable=False)

    # Relationships
    lead = relationship("Lead", back_populates="opportunity")
    organization = relationship("Organization", back_populates="opportunities")


# Functional data models (TypedDict)
class OpportunityData(TypedDict, total=False):
    """Functional opportunity data model."""
    id: int
    organization_id: int
    organization: Optional[dict]  # Nested OrganizationData when loaded
    opportunity_name: str
    estimated_value: Optional[Decimal]
    probability: Optional[Decimal]
    expected_close_date: Optional[date]
    notes: Optional[str]
    created_at: Optional[datetime]
    updated_at: Optional[datetime]
    # Lead fields (when joined)
    lead: Optional[dict]


class OpportunityCreateData(TypedDict, total=False):
    """Functional opportunity creation data model."""
    organization_id: int
    opportunity_name: str
    estimated_value: Optional[Decimal]
    probability: Optional[Decimal]
    expected_close_date: Optional[date]
    notes: Optional[str]
    # Lead fields
    deal_id: int
    title: str
    description: Optional[str]
    source: Optional[str]
    current_status_id: Optional[int]
    owner_user_id: Optional[int]


class OpportunityUpdateData(TypedDict, total=False):
    """Functional opportunity update data model."""
    organization_id: Optional[int]
    opportunity_name: Optional[str]
    estimated_value: Optional[Decimal]
    probability: Optional[Decimal]
    expected_close_date: Optional[date]
    notes: Optional[str]


# Conversion functions
def orm_to_dict(opp: Opportunity) -> OpportunityData:
    """Convert SQLAlchemy model to functional dict."""
    from models.Organization import orm_to_dict as org_to_dict
    from models.Lead import orm_to_dict as lead_to_dict

    result = OpportunityData(
        id=opp.id,
        organization_id=opp.organization_id,
        opportunity_name=opp.opportunity_name or "",
        estimated_value=opp.estimated_value,
        probability=opp.probability,
        expected_close_date=opp.expected_close_date,
        notes=opp.notes,
        created_at=opp.created_at,
        updated_at=opp.updated_at
    )

    if opp.organization:
        result["organization"] = org_to_dict(opp.organization)
    if opp.lead:
        result["lead"] = lead_to_dict(opp.lead)

    return result


def dict_to_orm(data: OpportunityCreateData) -> Opportunity:
    """Convert functional dict to SQLAlchemy model."""
    return Opportunity(
        organization_id=data.get("organization_id"),
        opportunity_name=data.get("opportunity_name", ""),
        estimated_value=data.get("estimated_value"),
        probability=data.get("probability"),
        expected_close_date=data.get("expected_close_date"),
        notes=data.get("notes")
    )


def update_orm_from_dict(opp: Opportunity, data: OpportunityUpdateData) -> Opportunity:
    """Update SQLAlchemy model from functional dict."""
    if "organization_id" in data and data["organization_id"] is not None:
        opp.organization_id = data["organization_id"]
    if "opportunity_name" in data and data["opportunity_name"] is not None:
        opp.opportunity_name = data["opportunity_name"]
    if "estimated_value" in data:
        opp.estimated_value = data["estimated_value"]
    if "probability" in data:
        opp.probability = data["probability"]
    if "expected_close_date" in data:
        opp.expected_close_date = data["expected_close_date"]
    if "notes" in data:
        opp.notes = data["notes"]
    return opp
