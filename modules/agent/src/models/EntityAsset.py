"""EntityAsset model - polymorphic junction for linking assets to entities."""

from sqlalchemy import Column, BigInteger, Text, DateTime, ForeignKey, func
from sqlalchemy.orm import relationship
from models import Base


class EntityAsset(Base):
    """Polymorphic junction for linking any asset to any entity."""

    __tablename__ = "Entity_Asset"

    asset_id = Column(BigInteger, ForeignKey("Asset.id", ondelete="CASCADE"), primary_key=True)
    entity_type = Column(Text, primary_key=True)  # 'Project', 'Lead', 'Account', etc.
    entity_id = Column(BigInteger, primary_key=True)
    created_at = Column(DateTime(timezone=True), server_default=func.now(), nullable=False)
    updated_at = Column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now(), nullable=False)

    # Relationships
    asset = relationship("Asset", back_populates="entity_links")
