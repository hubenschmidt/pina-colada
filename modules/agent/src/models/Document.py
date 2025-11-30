"""Document model - extends Asset for file storage."""

from sqlalchemy import Column, Text, BigInteger, ForeignKey

from models.Asset import Asset


class Document(Asset):
    """Document asset - files stored in external storage (local/R2)."""

    __tablename__ = "Document"

    id = Column(BigInteger, ForeignKey("Asset.id", ondelete="CASCADE"), primary_key=True)
    storage_path = Column(Text, nullable=False)  # path in storage backend
    file_size = Column(BigInteger, nullable=False)

    __mapper_args__ = {
        "polymorphic_identity": "document",
    }
