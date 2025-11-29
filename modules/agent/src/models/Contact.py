from sqlalchemy import Column, Text, DateTime, BigInteger, Boolean, ForeignKey, func, Table
from sqlalchemy.orm import relationship
from models import Base

"""Contact model with many-to-many relationships to Individual and Organization."""


class ContactIndividual(Base):
    """Junction table for Contact-Individual many-to-many relationship."""
    __tablename__ = "ContactIndividual"

    id = Column(BigInteger, primary_key=True, autoincrement=True)
    contact_id = Column(BigInteger, ForeignKey("Contact.id", ondelete="CASCADE"), nullable=False)
    individual_id = Column(BigInteger, ForeignKey("Individual.id", ondelete="CASCADE"), nullable=False)
    created_at = Column(DateTime(timezone=True), server_default=func.now(), nullable=False)


class ContactOrganization(Base):
    """Junction table for Contact-Organization many-to-many relationship."""
    __tablename__ = "ContactOrganization"

    id = Column(BigInteger, primary_key=True, autoincrement=True)
    contact_id = Column(BigInteger, ForeignKey("Contact.id", ondelete="CASCADE"), nullable=False)
    organization_id = Column(BigInteger, ForeignKey("Organization.id", ondelete="CASCADE"), nullable=False)
    is_primary = Column(Boolean, default=False)
    created_at = Column(DateTime(timezone=True), server_default=func.now(), nullable=False)


class Contact(Base):
    """Contact SQLAlchemy model - standalone entity that can link to multiple Individuals/Organizations."""

    __tablename__ = "Contact"

    id = Column(BigInteger, primary_key=True, autoincrement=True)
    first_name = Column(Text, nullable=False)
    last_name = Column(Text, nullable=False)
    title = Column(Text, nullable=True)
    department = Column(Text, nullable=True)
    role = Column(Text, nullable=True)
    email = Column(Text, nullable=True)
    phone = Column(Text, nullable=True)
    is_primary = Column(Boolean, nullable=False, default=False)
    notes = Column(Text, nullable=True)
    created_at = Column(DateTime(timezone=True), server_default=func.now(), nullable=False)
    updated_at = Column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now(), nullable=False)

    # Many-to-many relationships via junction tables
    individuals = relationship(
        "Individual",
        secondary="ContactIndividual",
        back_populates="contacts",
        lazy="selectin"
    )
    organizations = relationship(
        "Organization",
        secondary="ContactOrganization",
        back_populates="contacts",
        lazy="selectin"
    )
