"""Signal model for intent signals on Accounts."""

from sqlalchemy import Column, Text, DateTime, BigInteger, Date, Numeric, func, ForeignKey
from sqlalchemy.orm import relationship
from models import Base


class Signal(Base):
    """Signal SQLAlchemy model for intent signals."""

    __tablename__ = "Signal"

    id = Column(BigInteger, primary_key=True, autoincrement=True)
    account_id = Column(BigInteger, ForeignKey("Account.id", ondelete="CASCADE"), nullable=False)
    signal_type = Column(Text, nullable=False)   # 'hiring', 'expansion', 'product_launch', 'partnership', 'leadership_change', 'news'
    headline = Column(Text, nullable=False)
    description = Column(Text, nullable=True)
    signal_date = Column(Date, nullable=True)
    source = Column(Text, nullable=True)         # 'linkedin', 'news', 'crunchbase', 'agent'
    source_url = Column(Text, nullable=True)
    sentiment = Column(Text, nullable=True)      # 'positive', 'neutral', 'negative'
    relevance_score = Column(Numeric(3, 2), nullable=True)  # 0.00 to 1.00
    created_at = Column(DateTime(timezone=True), server_default=func.now(), nullable=False)
    updated_at = Column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now(), nullable=False)

    # Relationships
    account = relationship("Account", back_populates="signals")
