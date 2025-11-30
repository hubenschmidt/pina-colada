"""Organization model for companies."""

from sqlalchemy import Column, Text, Integer, DateTime, BigInteger, ForeignKey, func, Index, CheckConstraint
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

    # Firmographic columns
    revenue_range_id = Column(BigInteger, ForeignKey("RevenueRange.id", ondelete="SET NULL"), nullable=True)
    founding_year = Column(Integer, nullable=True)
    headquarters_city = Column(Text, nullable=True)
    headquarters_state = Column(Text, nullable=True)
    headquarters_country = Column(Text, nullable=True, default='USA')
    company_type = Column(Text, nullable=True)  # 'private', 'public', 'nonprofit', 'government'
    linkedin_url = Column(Text, nullable=True)
    crunchbase_url = Column(Text, nullable=True)

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
    revenue_range = relationship("RevenueRange")
    technologies = relationship("OrganizationTechnology", back_populates="organization", lazy="selectin")
    funding_rounds = relationship("FundingRound", back_populates="organization", lazy="selectin")
    signals = relationship("CompanySignal", back_populates="organization", lazy="selectin")

    __table_args__ = (
        Index('idx_organization_name_lower', func.lower(name), unique=True),
        CheckConstraint(
            "phone IS NULL OR phone ~ '^\\+1-\\d{3}-\\d{3}-\\d{4}$'",
            name="organization_phone_format_check"
        ),
        CheckConstraint(
            "founding_year IS NULL OR (founding_year >= 1800 AND founding_year <= EXTRACT(YEAR FROM NOW()))",
            name="organization_founding_year_check"
        ),
    )
