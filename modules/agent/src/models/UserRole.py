from sqlalchemy import Column, DateTime, BigInteger, ForeignKey, func, PrimaryKeyConstraint
from sqlalchemy.orm import relationship
from models import Base

"""UserRole model for user-role junction table."""




class UserRole(Base):
    """UserRole SQLAlchemy model (many-to-many junction)."""

    __tablename__ = "UserRole"

    user_id = Column(BigInteger, ForeignKey("User.id", ondelete="CASCADE"), nullable=False)
    role_id = Column(BigInteger, ForeignKey("Role.id", ondelete="CASCADE"), nullable=False)
    created_at = Column(DateTime(timezone=True), server_default=func.now(), nullable=False)

    # Relationships
    user = relationship("User", back_populates="user_roles")
    role = relationship("Role", back_populates="user_roles")

    __table_args__ = (
        PrimaryKeyConstraint('user_id', 'role_id'),
    )
