"""Asset model - base class for joined table inheritance."""

from sqlalchemy import Column, Text, DateTime, BigInteger, ForeignKey, func, Index, Integer, Boolean
from sqlalchemy.orm import relationship

from models import Base


class Asset(Base):
    """Base asset class using joined table inheritance.

    Subtypes: Document, Image (future)
    Supports versioning: versions share entity links and are identified by version number.
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
    created_by = Column(BigInteger, ForeignKey("User.id"), nullable=False)
    updated_by = Column(BigInteger, ForeignKey("User.id"), nullable=False)

    # Versioning columns
    parent_id = Column(BigInteger, ForeignKey("Asset.id", ondelete="CASCADE"), nullable=True)
    version_number = Column(Integer, nullable=False, default=1)
    is_current_version = Column(Boolean, nullable=False, default=True)

    # Relationships
    tenant = relationship("Tenant", back_populates="assets")
    user = relationship("User", back_populates="assets", foreign_keys=[user_id])
    parent = relationship("Asset", remote_side=[id], backref="versions")

    __mapper_args__ = {
        "polymorphic_on": asset_type,
        "polymorphic_identity": "asset",
    }

    __table_args__ = (
        Index("idx_asset_tenant", "tenant_id"),
        Index("idx_asset_user", "user_id"),
        Index("idx_asset_type", "asset_type"),
        Index("idx_asset_parent", "parent_id"),
    )
