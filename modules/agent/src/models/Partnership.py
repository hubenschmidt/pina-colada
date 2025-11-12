"""Partnership model (extends Lead via Joined Table Inheritance)."""

from datetime import datetime, date
from typing import TypedDict, Optional
from sqlalchemy import Column, Text, DateTime, BigInteger, Date, ForeignKey, func
from sqlalchemy.orm import relationship

from models import Base


class Partnership(Base):
    """Partnership SQLAlchemy model (extends Lead)."""

    __tablename__ = "Partnership"

    id = Column(BigInteger, ForeignKey("Lead.id", ondelete="CASCADE"), primary_key=True)
    organization_id = Column(BigInteger, ForeignKey("Organization.id", ondelete="CASCADE"), nullable=False)
    partnership_type = Column(Text, nullable=True)
    partnership_name = Column(Text, nullable=False)
    start_date = Column(Date, nullable=True)
    end_date = Column(Date, nullable=True)
    notes = Column(Text, nullable=True)
    created_at = Column(DateTime(timezone=True), server_default=func.now(), nullable=False)
    updated_at = Column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now(), nullable=False)

    # Relationships
    lead = relationship("Lead", back_populates="partnership")
    organization = relationship("Organization", back_populates="partnerships")


# Functional data models (TypedDict)
class PartnershipData(TypedDict, total=False):
    """Functional partnership data model."""
    id: int
    organization_id: int
    organization: Optional[dict]  # Nested OrganizationData when loaded
    partnership_type: Optional[str]
    partnership_name: str
    start_date: Optional[date]
    end_date: Optional[date]
    notes: Optional[str]
    created_at: Optional[datetime]
    updated_at: Optional[datetime]
    # Lead fields (when joined)
    lead: Optional[dict]


class PartnershipCreateData(TypedDict, total=False):
    """Functional partnership creation data model."""
    organization_id: int
    partnership_type: Optional[str]
    partnership_name: str
    start_date: Optional[date]
    end_date: Optional[date]
    notes: Optional[str]
    # Lead fields
    deal_id: int
    title: str
    description: Optional[str]
    source: Optional[str]
    current_status_id: Optional[int]
    owner_user_id: Optional[int]


class PartnershipUpdateData(TypedDict, total=False):
    """Functional partnership update data model."""
    organization_id: Optional[int]
    partnership_type: Optional[str]
    partnership_name: Optional[str]
    start_date: Optional[date]
    end_date: Optional[date]
    notes: Optional[str]


# Conversion functions
def orm_to_dict(partnership: Partnership) -> PartnershipData:
    """Convert SQLAlchemy model to functional dict."""
    from models.Organization import orm_to_dict as org_to_dict
    from models.Lead import orm_to_dict as lead_to_dict

    result = PartnershipData(
        id=partnership.id,
        organization_id=partnership.organization_id,
        partnership_type=partnership.partnership_type,
        partnership_name=partnership.partnership_name or "",
        start_date=partnership.start_date,
        end_date=partnership.end_date,
        notes=partnership.notes,
        created_at=partnership.created_at,
        updated_at=partnership.updated_at
    )

    if partnership.organization:
        result["organization"] = org_to_dict(partnership.organization)
    if partnership.lead:
        result["lead"] = lead_to_dict(partnership.lead)

    return result


def dict_to_orm(data: PartnershipCreateData) -> Partnership:
    """Convert functional dict to SQLAlchemy model."""
    return Partnership(
        organization_id=data.get("organization_id"),
        partnership_type=data.get("partnership_type"),
        partnership_name=data.get("partnership_name", ""),
        start_date=data.get("start_date"),
        end_date=data.get("end_date"),
        notes=data.get("notes")
    )


def update_orm_from_dict(partnership: Partnership, data: PartnershipUpdateData) -> Partnership:
    """Update SQLAlchemy model from functional dict."""
    if "organization_id" in data and data["organization_id"] is not None:
        partnership.organization_id = data["organization_id"]
    if "partnership_type" in data:
        partnership.partnership_type = data["partnership_type"]
    if "partnership_name" in data and data["partnership_name"] is not None:
        partnership.partnership_name = data["partnership_name"]
    if "start_date" in data:
        partnership.start_date = data["start_date"]
    if "end_date" in data:
        partnership.end_date = data["end_date"]
    if "notes" in data:
        partnership.notes = data["notes"]
    return partnership
