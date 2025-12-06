"""Service layer for conversation business logic."""

import asyncio
import logging
from typing import Optional, Dict, Any, List
from uuid import UUID

from fastapi import HTTPException
from langchain_anthropic import ChatAnthropic
from langchain_core.messages import HumanMessage, SystemMessage

from repositories import conversation_repository as repo
from schemas.conversation import (
    ConversationResponse,
    ConversationWithUserResponse,
    ConversationWithMessagesResponse,
    MessageResponse,
)

logger = logging.getLogger(__name__)


async def list_conversations(
    user_id: int,
    tenant_id: int,
    limit: int = 50,
    include_archived: bool = False,
) -> List[ConversationResponse]:
    """List conversations for a user."""
    conversations = await repo.list_conversations(
        user_id=user_id,
        tenant_id=tenant_id,
        limit=limit,
        include_archived=include_archived,
    )
    return [ConversationResponse.model_validate(c) for c in conversations]


async def list_tenant_conversations(
    tenant_id: int,
    search: Optional[str] = None,
    limit: int = 100,
    offset: int = 0,
    include_archived: bool = False,
) -> tuple[List[ConversationWithUserResponse], int]:
    """List all conversations for a tenant with search and pagination."""
    conversations, total = await repo.list_tenant_conversations(
        tenant_id=tenant_id,
        search=search,
        limit=limit,
        offset=offset,
        include_archived=include_archived,
    )
    return [ConversationWithUserResponse.model_validate(c) for c in conversations], total


async def get_conversation(thread_id: UUID) -> ConversationResponse:
    """Get conversation by thread_id."""
    conversation = await repo.get_conversation(thread_id)
    if not conversation:
        raise HTTPException(status_code=404, detail="Conversation not found")
    return ConversationResponse.model_validate(conversation)


async def get_conversation_with_messages(thread_id: UUID) -> ConversationWithMessagesResponse:
    """Get conversation with all messages."""
    result = await repo.get_conversation_with_messages(thread_id)
    if not result:
        raise HTTPException(status_code=404, detail="Conversation not found")

    return ConversationWithMessagesResponse(
        conversation=ConversationResponse.model_validate(result["conversation"]),
        messages=[MessageResponse.model_validate(m) for m in result["messages"]],
    )


async def create_conversation(
    user_id: int,
    tenant_id: int,
    thread_id: UUID,
    title: Optional[str] = None,
) -> ConversationResponse:
    """Create a new conversation."""
    conversation = await repo.create_conversation(
        user_id=user_id,
        tenant_id=tenant_id,
        thread_id=thread_id,
        title=title,
    )
    return ConversationResponse.model_validate(conversation)


async def update_title(thread_id: UUID, title: str) -> ConversationResponse:
    """Update conversation title."""
    if not title.strip():
        raise HTTPException(status_code=400, detail="Title cannot be empty")

    conversation = await repo.update_conversation_title(thread_id, title.strip())
    if not conversation:
        raise HTTPException(status_code=404, detail="Conversation not found")
    return ConversationResponse.model_validate(conversation)


async def archive_conversation(thread_id: UUID) -> bool:
    """Archive a conversation."""
    success = await repo.archive_conversation(thread_id)
    if not success:
        raise HTTPException(status_code=404, detail="Conversation not found")
    return True


async def unarchive_conversation(thread_id: UUID) -> bool:
    """Unarchive a conversation."""
    success = await repo.unarchive_conversation(thread_id)
    if not success:
        raise HTTPException(status_code=404, detail="Conversation not found")
    return True


async def delete_permanent(thread_id: UUID) -> bool:
    """Permanently delete a conversation (must be archived first)."""
    success = await repo.delete_conversation_permanent(thread_id)
    if not success:
        raise HTTPException(status_code=400, detail="Conversation not found or not archived")
    return True


async def add_message(
    thread_id: UUID,
    user_id: int,
    tenant_id: int,
    role: str,
    content: str,
    token_usage: Optional[Dict[str, int]] = None,
) -> MessageResponse:
    """Add a message to a conversation, creating the conversation if needed."""
    conversation = await repo.get_or_create_conversation(user_id, tenant_id, thread_id)
    message = await repo.add_message(
        conversation_id=conversation.id,
        role=role,
        content=content,
        token_usage=token_usage,
    )
    await repo.touch_conversation(thread_id)
    return MessageResponse.model_validate(message)


async def generate_title(user_message: str, assistant_response: str) -> str:
    """Generate a 3-5 word title using Claude Haiku."""
    try:
        llm = ChatAnthropic(
            model="claude-haiku-4-5-20251001",
            temperature=0,
            max_tokens=30,
        )

        prompt = f"""Generate a 3-5 word title summarizing this conversation. Return ONLY the title, no quotes or punctuation.

User: {user_message[:200]}
Assistant: {assistant_response[:200]}

Title:"""

        messages = [
            SystemMessage(content="You generate short, descriptive titles. Return only the title text."),
            HumanMessage(content=prompt),
        ]

        response = await llm.ainvoke(messages)
        title = response.content.strip().strip('"').strip("'")

        if len(title) > 100:
            title = title[:97] + "..."

        return title

    except Exception as e:
        logger.warning(f"Failed to generate title: {e}")
        return user_message[:50] + ("..." if len(user_message) > 50 else "")


async def generate_and_set_title(
    thread_id: UUID,
    user_message: str,
    assistant_response: str,
) -> None:
    """Generate title in background and update conversation."""
    try:
        title = await generate_title(user_message, assistant_response)
        await repo.update_conversation_title(thread_id, title)
        logger.info(f"Set conversation title: {title}")
    except Exception as e:
        logger.error(f"Failed to set conversation title: {e}")


def schedule_title_generation(
    thread_id: UUID,
    user_message: str,
    assistant_response: str,
) -> None:
    """Schedule title generation as background task."""
    asyncio.create_task(
        generate_and_set_title(thread_id, user_message, assistant_response)
    )
