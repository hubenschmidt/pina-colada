from sqlalchemy import Column, Text, DateTime, BigInteger, func
from sqlalchemy.orm import relationship
from models import Base

"""Role model for system-defined roles."""




class Role(Base):
    """Role SQLAlchemy model (system-defined roles)."""

    __tablename__ = "Role"

    id = Column(BigInteger, primary_key=True, autoincrement=True)
    name = Column(Text, nullable=False, unique=True)
    description = Column(Text, nullable=True)
    created_at = Column(DateTime(timezone=True), server_default=func.now(), nullable=False)

    # Relationships
    user_roles = relationship("UserRole", back_populates="role")
