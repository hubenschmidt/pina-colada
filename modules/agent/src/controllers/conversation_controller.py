"""Controller layer for conversation routing to services."""

from typing import List
from uuid import UUID

from fastapi import Request

from lib.decorators import handle_http_exceptions
from schemas.conversation import ConversationUpdate
from services import conversation_service as service


@handle_http_exceptions
async def list_conversations(
    request: Request,
    limit: int = 50,
    include_archived: bool = False,
) -> List[dict]:
    """List conversations for the current user."""
    user_id = request.state.user_id
    tenant_id = request.state.tenant_id
    conversations = await service.list_conversations(
        user_id=user_id,
        tenant_id=tenant_id,
        limit=limit,
        include_archived=include_archived,
    )
    return [c.model_dump() for c in conversations]


@handle_http_exceptions
async def list_tenant_conversations(
    request: Request,
    search: str = None,
    limit: int = 100,
    offset: int = 0,
    include_archived: bool = False,
) -> dict:
    """List all conversations for the tenant (with search and pagination)."""
    tenant_id = request.state.tenant_id
    conversations, total = await service.list_tenant_conversations(
        tenant_id=tenant_id,
        search=search,
        limit=limit,
        offset=offset,
        include_archived=include_archived,
    )
    return {
        "data": [c.model_dump() for c in conversations],
        "total": total,
        "limit": limit,
        "offset": offset,
    }


@handle_http_exceptions
async def get_conversation(request: Request, thread_id: UUID) -> dict:
    """Get a conversation by thread_id."""
    result = await service.get_conversation_with_messages(thread_id)
    return result.model_dump()


@handle_http_exceptions
async def update_title(request: Request, thread_id: UUID, data: ConversationUpdate) -> dict:
    """Update conversation title."""
    conversation = await service.update_title(thread_id, data.title)
    return conversation.model_dump()


@handle_http_exceptions
async def archive_conversation(request: Request, thread_id: UUID) -> dict:
    """Archive a conversation."""
    await service.archive_conversation(thread_id)
    return {"status": "archived"}


@handle_http_exceptions
async def unarchive_conversation(request: Request, thread_id: UUID) -> dict:
    """Unarchive a conversation."""
    await service.unarchive_conversation(thread_id)
    return {"status": "unarchived"}


@handle_http_exceptions
async def delete_permanent(request: Request, thread_id: UUID) -> dict:
    """Permanently delete an archived conversation."""
    await service.delete_permanent(thread_id)
    return {"status": "deleted"}
