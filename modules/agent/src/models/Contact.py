"""Contact model (links Individual to Organization)."""

from datetime import datetime
from typing import TypedDict, Optional
from sqlalchemy import Column, Text, DateTime, BigInteger, Boolean, ForeignKey, func
from sqlalchemy.orm import relationship

from models import Base


class Contact(Base):
    """Contact SQLAlchemy model (links Individual to Organization)."""

    __tablename__ = "Contact"

    id = Column(BigInteger, primary_key=True, autoincrement=True)
    individual_id = Column(BigInteger, ForeignKey("Individual.id", ondelete="CASCADE"), nullable=False)
    organization_id = Column(BigInteger, ForeignKey("Organization.id", ondelete="SET NULL"), nullable=True)
    title = Column(Text, nullable=True)
    department = Column(Text, nullable=True)
    role = Column(Text, nullable=True)
    email = Column(Text, nullable=True)
    phone = Column(Text, nullable=True)
    is_primary = Column(Boolean, nullable=False, default=False)
    notes = Column(Text, nullable=True)
    created_at = Column(DateTime(timezone=True), server_default=func.now(), nullable=False)
    updated_at = Column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now(), nullable=False)

    # Relationships
    individual = relationship("Individual", back_populates="contacts")
    organization = relationship("Organization", back_populates="contacts")


# Functional data models (TypedDict)
class ContactData(TypedDict, total=False):
    """Functional contact data model."""
    id: int
    individual_id: int
    individual: Optional[dict]  # Nested IndividualData when loaded
    organization_id: Optional[int]
    organization: Optional[dict]  # Nested OrganizationData when loaded
    title: Optional[str]
    department: Optional[str]
    role: Optional[str]
    email: Optional[str]
    phone: Optional[str]
    is_primary: bool
    notes: Optional[str]
    created_at: Optional[datetime]
    updated_at: Optional[datetime]


class ContactCreateData(TypedDict, total=False):
    """Functional contact creation data model."""
    individual_id: int
    organization_id: Optional[int]
    title: Optional[str]
    department: Optional[str]
    role: Optional[str]
    email: Optional[str]
    phone: Optional[str]
    is_primary: bool
    notes: Optional[str]


class ContactUpdateData(TypedDict, total=False):
    """Functional contact update data model."""
    individual_id: Optional[int]
    organization_id: Optional[int]
    title: Optional[str]
    department: Optional[str]
    role: Optional[str]
    email: Optional[str]
    phone: Optional[str]
    is_primary: Optional[bool]
    notes: Optional[str]


# Conversion functions
def orm_to_dict(contact: Contact) -> ContactData:
    """Convert SQLAlchemy model to functional dict."""
    from models.Individual import orm_to_dict as individual_to_dict
    from models.Organization import orm_to_dict as org_to_dict

    result = ContactData(
        id=contact.id,
        individual_id=contact.individual_id,
        organization_id=contact.organization_id,
        title=contact.title,
        department=contact.department,
        role=contact.role,
        email=contact.email,
        phone=contact.phone,
        is_primary=contact.is_primary or False,
        notes=contact.notes,
        created_at=contact.created_at,
        updated_at=contact.updated_at
    )

    if contact.individual:
        result["individual"] = individual_to_dict(contact.individual)
    if contact.organization:
        result["organization"] = org_to_dict(contact.organization)

    return result


def dict_to_orm(data: ContactCreateData) -> Contact:
    """Convert functional dict to SQLAlchemy model."""
    return Contact(
        individual_id=data.get("individual_id"),
        organization_id=data.get("organization_id"),
        title=data.get("title"),
        department=data.get("department"),
        role=data.get("role"),
        email=data.get("email"),
        phone=data.get("phone"),
        is_primary=data.get("is_primary", False),
        notes=data.get("notes")
    )


def update_orm_from_dict(contact: Contact, data: ContactUpdateData) -> Contact:
    """Update SQLAlchemy model from functional dict."""
    if "individual_id" in data and data["individual_id"] is not None:
        contact.individual_id = data["individual_id"]
    if "organization_id" in data:
        contact.organization_id = data["organization_id"]
    if "title" in data:
        contact.title = data["title"]
    if "department" in data:
        contact.department = data["department"]
    if "role" in data:
        contact.role = data["role"]
    if "email" in data:
        contact.email = data["email"]
    if "phone" in data:
        contact.phone = data["phone"]
    if "is_primary" in data and data["is_primary"] is not None:
        contact.is_primary = data["is_primary"]
    if "notes" in data:
        contact.notes = data["notes"]
    return contact
