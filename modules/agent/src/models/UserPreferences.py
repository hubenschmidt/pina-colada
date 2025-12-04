from sqlalchemy import Column, Text, DateTime, BigInteger, ForeignKey, func
from sqlalchemy.orm import relationship
from models import Base


class UserPreferences(Base):
    """UserPreferences SQLAlchemy model (theme and locale settings for individual user)."""

    __tablename__ = "User_Preferences"

    id = Column(BigInteger, primary_key=True, autoincrement=True)
    user_id = Column(BigInteger, ForeignKey("User.id", ondelete="CASCADE"), nullable=False, unique=True, index=True)
    theme = Column(Text, nullable=True)
    timezone = Column(Text, nullable=True, default="America/New_York")
    created_at = Column(DateTime(timezone=True), server_default=func.now(), nullable=False)
    updated_at = Column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now(), nullable=False)

    # Relationships
    user = relationship("User", back_populates="preferences", uselist=False, foreign_keys=[user_id])
