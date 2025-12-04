from sqlalchemy import Column, Text, BigInteger, Date, ForeignKey
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

    # Relationships
    lead = relationship("Lead", back_populates="partnership")
