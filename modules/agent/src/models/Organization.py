"""Organization model for companies."""

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
