from sqlalchemy import Column, Text, DateTime, BigInteger, Integer, ForeignKey, func
from models import Base


class UsageAnalytics(Base):
    """Token usage analytics for tracking spend by node/tool/model."""

    __tablename__ = "Usage_Analytics"

    id = Column(BigInteger, primary_key=True, autoincrement=True)
    tenant_id = Column(BigInteger, ForeignKey("Tenant.id"), nullable=False)
    user_id = Column(BigInteger, ForeignKey("User.id"), nullable=False)
    conversation_id = Column(BigInteger, ForeignKey("Conversation.id"), nullable=True)
    message_id = Column(BigInteger, ForeignKey("Conversation_Message.id"), nullable=True)

    input_tokens = Column(Integer, nullable=False, default=0)
    output_tokens = Column(Integer, nullable=False, default=0)
    total_tokens = Column(Integer, nullable=False, default=0)

    node_name = Column(Text, nullable=True)
    tool_name = Column(Text, nullable=True)
    model_name = Column(Text, nullable=True)

    created_at = Column(DateTime(timezone=True), server_default=func.now(), nullable=False)
