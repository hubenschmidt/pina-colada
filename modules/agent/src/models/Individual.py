"""Individual model for people."""

from sqlalchemy import Column, Text, DateTime, BigInteger, ForeignKey, func, Index
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
    notes = Column(Text, nullable=True)
    created_at = Column(DateTime(timezone=True), server_default=func.now(), nullable=False)
    updated_at = Column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now(), nullable=False)

    # Relationships
    account = relationship("Account", back_populates="individuals")
    contacts = relationship("Contact", back_populates="individual")
    user = relationship("User", back_populates="individual", uselist=False)

    __table_args__ = (
        Index('idx_individual_email_lower', func.lower(email), unique=True),
    )
