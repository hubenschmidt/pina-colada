"""Asset model - base class for joined table inheritance."""

from sqlalchemy import Column, Text, DateTime, BigInteger, ForeignKey, func, Index
from sqlalchemy.orm import relationship

from models import Base


class Asset(Base):
    """Base asset class using joined table inheritance.

    Subtypes: Document, Image (future)
    """

    __tablename__ = "Asset"

    id = Column(BigInteger, primary_key=True, autoincrement=True)
    asset_type = Column(Text, nullable=False)  # discriminator: 'document', 'image'
    tenant_id = Column(BigInteger, ForeignKey("Tenant.id", ondelete="CASCADE"), nullable=False)
    user_id = Column(BigInteger, ForeignKey("User.id", ondelete="CASCADE"), nullable=False)
    filename = Column(Text, nullable=False)
    content_type = Column(Text, nullable=False)  # MIME type
    description = Column(Text, nullable=True)
    created_at = Column(DateTime(timezone=True), server_default=func.now(), nullable=False)
    updated_at = Column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now(), nullable=False)

    # Relationships
    tenant = relationship("Tenant", back_populates="assets")
    user = relationship("User", back_populates="assets")

    __mapper_args__ = {
        "polymorphic_on": asset_type,
        "polymorphic_identity": "asset",
    }

    __table_args__ = (
        Index("idx_asset_tenant", "tenant_id"),
        Index("idx_asset_user", "user_id"),
        Index("idx_asset_type", "asset_type"),
    )
