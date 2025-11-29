from sqlalchemy import Column, Text, DateTime, BigInteger, ForeignKey, func
from sqlalchemy.orm import relationship
from models import Base

"""Lead model (base table for Joined Table Inheritance)."""




class Lead(Base):
    """Lead SQLAlchemy model (base table for Joined Table Inheritance)."""

    __tablename__ = "Lead"

    id = Column(BigInteger, primary_key=True, autoincrement=True)
    tenant_id = Column(BigInteger, ForeignKey("Tenant.id", ondelete="CASCADE"), nullable=True)
    account_id = Column(BigInteger, ForeignKey("Account.id", ondelete="SET NULL"), nullable=True)
    deal_id = Column(BigInteger, ForeignKey("Deal.id", ondelete="CASCADE"), nullable=False)
    type = Column(Text, nullable=False)  # Discriminator: 'Job', 'Opportunity', 'Partnership', etc.
    title = Column(Text, nullable=False)
    description = Column(Text, nullable=True)
    source = Column(Text, nullable=True)  # inbound/referral/event/campaign/agent/manual/etc.
    current_status_id = Column(BigInteger, ForeignKey("Status.id", ondelete="SET NULL"), nullable=True)
    owner_individual_id = Column(BigInteger, ForeignKey("Individual.id", ondelete="SET NULL"), nullable=True)
    created_at = Column(DateTime(timezone=True), server_default=func.now(), nullable=False)
    updated_at = Column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now(), nullable=False)

    # Relationships
    tenant = relationship("Tenant", back_populates="leads")
    account = relationship("Account", back_populates="leads")
    deal = relationship("Deal", back_populates="leads")
    current_status = relationship("Status", back_populates="leads", foreign_keys=[current_status_id])
    projects = relationship("Project", secondary="LeadProject", back_populates="leads")

    # Joined table inheritance relationships
    job = relationship("Job", back_populates="lead", uselist=False)
    opportunity = relationship("Opportunity", back_populates="lead", uselist=False)
    partnership = relationship("Partnership", back_populates="lead", uselist=False)
