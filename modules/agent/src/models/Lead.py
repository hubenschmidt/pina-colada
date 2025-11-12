"""Lead model (base table for Joined Table Inheritance)."""

from datetime import datetime
from typing import TypedDict, Optional
from sqlalchemy import Column, Text, DateTime, BigInteger, ForeignKey, func
from sqlalchemy.orm import relationship

from models import Base


class Lead(Base):
    """Lead SQLAlchemy model (base table for Joined Table Inheritance)."""

    __tablename__ = "Lead"

    id = Column(BigInteger, primary_key=True, autoincrement=True)
    deal_id = Column(BigInteger, ForeignKey("Deal.id", ondelete="CASCADE"), nullable=False)
    type = Column(Text, nullable=False)  # Discriminator: 'Job', 'Opportunity', 'Partnership', etc.
    title = Column(Text, nullable=False)
    description = Column(Text, nullable=True)
    source = Column(Text, nullable=True)  # inbound/referral/event/campaign/agent/manual/etc.
    current_status_id = Column(BigInteger, ForeignKey("Status.id", ondelete="SET NULL"), nullable=True)
    owner_user_id = Column(BigInteger, nullable=True)
    created_at = Column(DateTime(timezone=True), server_default=func.now(), nullable=False)
    updated_at = Column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now(), nullable=False)

    # Relationships
    deal = relationship("Deal", back_populates="leads")
    current_status = relationship("Status", back_populates="leads", foreign_keys=[current_status_id])

    # Joined table inheritance relationships
    job = relationship("Job", back_populates="lead", uselist=False)
    opportunity = relationship("Opportunity", back_populates="lead", uselist=False)
    partnership = relationship("Partnership", back_populates="lead", uselist=False)


# Functional data models (TypedDict)
class LeadData(TypedDict, total=False):
    """Functional lead data model."""
    id: int
    deal_id: int
    deal: Optional[dict]  # Nested DealData when loaded
    type: str
    title: str
    description: Optional[str]
    source: Optional[str]
    current_status_id: Optional[int]
    current_status: Optional[dict]  # Nested StatusData when loaded
    owner_user_id: Optional[int]
    created_at: Optional[datetime]
    updated_at: Optional[datetime]


class LeadCreateData(TypedDict, total=False):
    """Functional lead creation data model."""
    deal_id: int
    type: str
    title: str
    description: Optional[str]
    source: Optional[str]
    current_status_id: Optional[int]
    owner_user_id: Optional[int]


class LeadUpdateData(TypedDict, total=False):
    """Functional lead update data model."""
    deal_id: Optional[int]
    type: Optional[str]
    title: Optional[str]
    description: Optional[str]
    source: Optional[str]
    current_status_id: Optional[int]
    owner_user_id: Optional[int]


# Conversion functions
def orm_to_dict(lead: Lead) -> LeadData:
    """Convert SQLAlchemy model to functional dict."""
    from models.Deal import orm_to_dict as deal_to_dict
    from models.Status import orm_to_dict as status_to_dict

    result = LeadData(
        id=lead.id,
        deal_id=lead.deal_id,
        type=lead.type or "",
        title=lead.title or "",
        description=lead.description,
        source=lead.source,
        current_status_id=lead.current_status_id,
        owner_user_id=lead.owner_user_id,
        created_at=lead.created_at,
        updated_at=lead.updated_at
    )

    # Include relationships if loaded
    if lead.deal:
        result["deal"] = deal_to_dict(lead.deal)
    if lead.current_status:
        result["current_status"] = status_to_dict(lead.current_status)

    return result


def dict_to_orm(data: LeadCreateData) -> Lead:
    """Convert functional dict to SQLAlchemy model."""
    return Lead(
        deal_id=data.get("deal_id"),
        type=data.get("type", ""),
        title=data.get("title", ""),
        description=data.get("description"),
        source=data.get("source"),
        current_status_id=data.get("current_status_id"),
        owner_user_id=data.get("owner_user_id")
    )


def update_orm_from_dict(lead: Lead, data: LeadUpdateData) -> Lead:
    """Update SQLAlchemy model from functional dict."""
    if "deal_id" in data and data["deal_id"] is not None:
        lead.deal_id = data["deal_id"]
    if "type" in data and data["type"] is not None:
        lead.type = data["type"]
    if "title" in data and data["title"] is not None:
        lead.title = data["title"]
    if "description" in data:
        lead.description = data["description"]
    if "source" in data:
        lead.source = data["source"]
    if "current_status_id" in data:
        lead.current_status_id = data["current_status_id"]
    if "owner_user_id" in data:
        lead.owner_user_id = data["owner_user_id"]
    return lead
