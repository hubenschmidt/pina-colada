from sqlalchemy import Column, Text, BigInteger, Date, DateTime, ForeignKey, func
from sqlalchemy.orm import relationship
from models import Base


class Partnership(Base):
    """Partnership SQLAlchemy model (extends Lead)."""

    __tablename__ = "Partnership"

    id = Column(BigInteger, ForeignKey("Lead.id", ondelete="CASCADE"), primary_key=True)
    partnership_type = Column(Text, nullable=True)
    partnership_name = Column(Text, nullable=False)
    start_date = Column(Date, nullable=True)
    end_date = Column(Date, nullable=True)
    description = Column(Text, nullable=True)
    created_at = Column(DateTime(timezone=True), server_default=func.now(), nullable=False)
    updated_at = Column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now(), nullable=False)

    # Relationships
    lead = relationship("Lead", back_populates="partnership")
