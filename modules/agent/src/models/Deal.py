"""Deal model."""

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
    owner_individual_id = Column(BigInteger, ForeignKey("Individual.id", ondelete="SET NULL"), nullable=True)
    current_status_id = Column(BigInteger, ForeignKey("Status.id", ondelete="SET NULL"), nullable=True)
    value_amount = Column(Numeric(18, 2), nullable=True)
    value_currency = Column(Text, default="USD")
    probability = Column(Numeric(5, 2), nullable=True)
    expected_close_date = Column(Date, nullable=True)
    close_date = Column(Date, nullable=True)
    created_at = Column(DateTime(timezone=True), server_default=func.now(), nullable=False)
    updated_at = Column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now(), nullable=False)

    # Relationships
    tenant = relationship("Tenant", back_populates="deals")
    current_status = relationship("Status", back_populates="deals")
    leads = relationship("Lead", back_populates="deal")

    __table_args__ = (
        Index('idx_deal_tenant_id', 'tenant_id'),
    )
