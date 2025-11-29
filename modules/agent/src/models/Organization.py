"""Organization model for companies."""

from sqlalchemy import Column, Text, Integer, DateTime, BigInteger, ForeignKey, func, Index
from sqlalchemy.orm import relationship

from models import Base


class Organization(Base):
    """Organization SQLAlchemy model for companies."""

    __tablename__ = "Organization"

    id = Column(BigInteger, primary_key=True, autoincrement=True)
    account_id = Column(BigInteger, ForeignKey("Account.id", ondelete="SET NULL"), nullable=True)
    name = Column(Text, nullable=False)
    website = Column(Text, nullable=True)
    phone = Column(Text, nullable=True)
    employee_count = Column(Integer, nullable=True)  # Legacy field
    employee_count_range_id = Column(BigInteger, ForeignKey("EmployeeCountRange.id", ondelete="SET NULL"), nullable=True)
    funding_stage_id = Column(BigInteger, ForeignKey("FundingStage.id", ondelete="SET NULL"), nullable=True)
    description = Column(Text, nullable=True)
    notes = Column(Text, nullable=True)
    created_at = Column(DateTime(timezone=True), server_default=func.now(), nullable=False)
    updated_at = Column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now(), nullable=False)

    # Relationships
    account = relationship("Account", back_populates="organizations")
    contacts = relationship(
        "Contact",
        secondary="ContactOrganization",
        back_populates="organizations",
        lazy="selectin"
    )
    employee_count_range = relationship("EmployeeCountRange")
    funding_stage = relationship("FundingStage")

    __table_args__ = (
        Index('idx_organization_name_lower', func.lower(name), unique=True),
    )
