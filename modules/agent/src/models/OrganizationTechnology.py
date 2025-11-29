"""OrganizationTechnology model for tech stack junction table."""

from sqlalchemy import Column, Text, DateTime, BigInteger, Numeric, ForeignKey, func
from sqlalchemy.orm import relationship

from models import Base


class OrganizationTechnology(Base):
    """OrganizationTechnology SQLAlchemy model (junction table)."""

    __tablename__ = "OrganizationTechnology"

    organization_id = Column(BigInteger, ForeignKey("Organization.id", ondelete="CASCADE"), primary_key=True)
    technology_id = Column(BigInteger, ForeignKey("Technology.id", ondelete="CASCADE"), primary_key=True)
    detected_at = Column(DateTime(timezone=True), server_default=func.now(), nullable=False)
    source = Column(Text, nullable=True)       # 'builtwith', 'wappalyzer', 'agent', 'manual'
    confidence = Column(Numeric(3, 2), nullable=True)

    # Relationships
    organization = relationship("Organization", back_populates="technologies")
    technology = relationship("Technology")
