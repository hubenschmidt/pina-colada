"""Individual model for people."""

from sqlalchemy import Column, Text, DateTime, BigInteger, Boolean, ForeignKey, func, Index, CheckConstraint
from sqlalchemy.orm import relationship

from models import Base


class Individual(Base):
    """Individual SQLAlchemy model for people."""

    __tablename__ = "Individual"

    id = Column(BigInteger, primary_key=True, autoincrement=True)
    account_id = Column(BigInteger, ForeignKey("Account.id", ondelete="SET NULL"), nullable=True)
    first_name = Column(Text, nullable=False)
    last_name = Column(Text, nullable=False)
    email = Column(Text, nullable=True)
    phone = Column(Text, nullable=True)
    linkedin_url = Column(Text, nullable=True)
    title = Column(Text, nullable=True)
    description = Column(Text, nullable=True)
    created_at = Column(DateTime(timezone=True), server_default=func.now(), nullable=False)
    updated_at = Column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now(), nullable=False)

    # Contact intelligence columns
    twitter_url = Column(Text, nullable=True)
    github_url = Column(Text, nullable=True)
    bio = Column(Text, nullable=True)
    seniority_level = Column(Text, nullable=True)  # 'C-Level', 'VP', 'Director', 'Manager', 'IC'
    department = Column(Text, nullable=True)       # 'Engineering', 'Sales', 'Marketing', etc.
    is_decision_maker = Column(Boolean, nullable=True, default=False)
    reports_to_id = Column(BigInteger, ForeignKey("Individual.id", ondelete="SET NULL"), nullable=True)

    # Relationships
    account = relationship("Account", back_populates="individuals")
    contacts = relationship(
        "Contact",
        secondary="ContactIndividual",
        back_populates="individuals",
        lazy="selectin"
    )
    user = relationship("User", back_populates="individual", uselist=False)
    reports_to = relationship("Individual", remote_side=[id], foreign_keys=[reports_to_id])
    direct_reports = relationship("Individual", back_populates="reports_to", foreign_keys=[reports_to_id])

    __table_args__ = (
        Index('idx_individual_email_lower', func.lower(email), unique=True),
        CheckConstraint(
            "phone IS NULL OR phone ~ '^\\+1-\\d{3}-\\d{3}-\\d{4}$'",
            name="individual_phone_format_check"
        ),
    )
