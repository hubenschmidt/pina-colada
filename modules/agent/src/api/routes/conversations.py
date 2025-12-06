"""Routes for conversations API endpoints."""

from uuid import UUID

from fastapi import APIRouter, Request

from controllers.conversation_controller import (
    list_conversations,
    list_tenant_conversations,
    get_conversation,
    update_title,
    archive_conversation,
    unarchive_conversation,
    delete_permanent,
)
from schemas.conversation import ConversationUpdate


router = APIRouter(prefix="/conversations", tags=["conversations"])


@router.get("")
async def list_conversations_route(
    request: Request,
    limit: int = 50,
    include_archived: bool = False,
):
    """List conversations for the current user."""
    return await list_conversations(request, limit, include_archived)


@router.get("/all")
async def list_tenant_conversations_route(
    request: Request,
    search: str = None,
    limit: int = 100,
    offset: int = 0,
    include_archived: bool = False,
):
    """List all conversations for the tenant (with search and pagination)."""
    return await list_tenant_conversations(request, search, limit, offset, include_archived)


@router.get("/{thread_id}")
async def get_conversation_route(request: Request, thread_id: UUID):
    """Get a conversation with messages by thread_id."""
    return await get_conversation(request, thread_id)


@router.patch("/{thread_id}")
async def update_title_route(request: Request, thread_id: UUID, data: ConversationUpdate):
    """Update conversation title."""
    return await update_title(request, thread_id, data)


@router.delete("/{thread_id}")
async def archive_conversation_route(request: Request, thread_id: UUID):
    """Archive a conversation (soft delete)."""
    return await archive_conversation(request, thread_id)


@router.post("/{thread_id}/unarchive")
async def unarchive_conversation_route(request: Request, thread_id: UUID):
    """Unarchive a conversation."""
    return await unarchive_conversation(request, thread_id)


@router.delete("/{thread_id}/permanent")
async def delete_permanent_route(request: Request, thread_id: UUID):
    """Permanently delete an archived conversation."""
    return await delete_permanent(request, thread_id)
