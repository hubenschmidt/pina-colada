"""Individual model for people."""

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
