from sqlalchemy import Column, Text, DateTime, BigInteger, ForeignKey, func
from sqlalchemy.orm import relationship
from models import Base


class OrganizationRelationship(Base):
    """Relationship between two organizations."""

    __tablename__ = "Organization_Relationship"

    id = Column(BigInteger, primary_key=True, autoincrement=True)
    from_organization_id = Column(BigInteger, ForeignKey("Organization.id", ondelete="CASCADE"), nullable=False)
    to_organization_id = Column(BigInteger, ForeignKey("Organization.id", ondelete="CASCADE"), nullable=False)
    relationship_type = Column(Text, nullable=True)
    notes = Column(Text, nullable=True)
    created_at = Column(DateTime(timezone=True), server_default=func.now(), nullable=False)
    updated_at = Column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now(), nullable=False)

    # Relationships
    from_organization = relationship("Organization", foreign_keys=[from_organization_id], backref="outgoing_relationships")
    to_organization = relationship("Organization", foreign_keys=[to_organization_id], backref="incoming_relationships")
