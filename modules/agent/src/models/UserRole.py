from sqlalchemy import Column, DateTime, BigInteger, ForeignKey, func, PrimaryKeyConstraint
from sqlalchemy.orm import relationship
from models import Base


class UserRole(Base):
    """UserRole SQLAlchemy model (many-to-many junction)."""

    __tablename__ = "User_Role"

    user_id = Column(BigInteger, ForeignKey("User.id", ondelete="CASCADE"), nullable=False)
    role_id = Column(BigInteger, ForeignKey("Role.id", ondelete="CASCADE"), nullable=False)
    created_at = Column(DateTime(timezone=True), server_default=func.now(), nullable=False)
    updated_at = Column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now(), nullable=False)
    created_by = Column(BigInteger, ForeignKey("User.id"), nullable=False)
    updated_by = Column(BigInteger, ForeignKey("User.id"), nullable=False)

    # Relationships
    user = relationship("User", back_populates="user_roles", foreign_keys=[user_id])
    role = relationship("Role", back_populates="user_roles", foreign_keys=[role_id])

    __table_args__ = (
        PrimaryKeyConstraint('user_id', 'role_id'),
    )
