"""
Orchestrator using OpenAI Agents SDK.

Uses Claude Haiku for routing, GPT-4.1 for workers.
Structured outputs, handoffs, and @function_tool decorators.
"""

import asyncio
import json
import logging
from typing import Any, Callable, Dict, Optional
from uuid import UUID

import os

from dotenv import load_dotenv
load_dotenv(override=True)

from openai import AsyncOpenAI
from agents import Runner, function_tool, trace, set_default_openai_api, OpenAIChatCompletionsModel

from agent.workers.general import create_general_worker
from agent.workers.career.job_search import create_job_search_worker
from agent.workers.career.cover_letter import create_cover_letter_worker
from agent.workers.crm.crm_worker import create_crm_worker
from agent.routers.triage import create_triage_router
from agent.util.agent_hooks import logging_hooks
from services.reasoning_service import get_reasoning_schema, format_schema_context
from services import conversation_service
from services.conversation_service import load_messages, save_message
from services import usage_service
from lib.tenant_context import set_tenant_id

logger = logging.getLogger(__name__)

# Use Chat Completions API
set_default_openai_api("chat_completions")

# Claude client for triage (fast routing)
_claude_client = AsyncOpenAI(
    api_key=os.getenv("ANTHROPIC_API_KEY"),
    base_url="https://api.anthropic.com/v1/",
)

# Models
GPT_5 = "gpt-5.1"
CLAUDE_HAIKU = OpenAIChatCompletionsModel(
    model="claude-haiku-4-5-20251001",
    openai_client=_claude_client,
)

# Track running tasks for cancellation
_running_tasks: Dict[str, asyncio.Task] = {}

# Track cumulative token usage per thread
_thread_token_totals: Dict[str, Dict[str, int]] = {}


# --- Tools ---

@function_tool
def web_search(query: str) -> str:
    """Search the web for current information (news, docs, general queries)."""
    from agent.tools.worker_tools import serper_search
    return serper_search(query)


@function_tool
async def job_search(query: str) -> str:
    """Search for jobs, filtering out already-applied positions."""
    from agent.tools.worker_tools import job_search_with_filter
    return await job_search_with_filter(query)


@function_tool
async def check_applied_jobs(query: str = "") -> str:
    """Check job applications in database. Empty string lists all."""
    from agent.tools.worker_tools import check_applied_jobs as _check
    return await _check(query)


@function_tool
def update_job_status(
    company: str,
    job_title: str,
    status: str,
    job_url: str = "",
    notes: str = "",
) -> str:
    """Update job application status (applied, interviewing, rejected, offer, accepted, do_not_apply)."""
    from agent.tools.worker_tools import update_job_status as _update
    return _update(company, job_title, status, job_url, notes)


@function_tool
def send_email(to_email: str, subject: str, body: str, job_listings: str = "") -> str:
    """Send an email to the specified address."""
    from agent.tools.worker_tools import send_email as _send
    return _send(to_email, subject, body, job_listings)


@function_tool
async def read_document(file_path: str) -> str:
    """Read a document file (PDF, DOCX, TXT)."""
    from agent.tools.document_tools import read_document as _read
    return await _read(file_path)


@function_tool
async def list_documents() -> str:
    """List available documents."""
    from agent.tools.document_tools import list_documents as _list
    return await _list()


@function_tool
async def search_entity_documents(entity_type: str, entity_id: int) -> str:
    """Search for documents linked to a specific entity (Individual, Organization, Account, Contact)."""
    from agent.tools.document_tools import search_entity_documents as _search
    return await _search(entity_type, entity_id)


@function_tool
async def crm_lookup(entity_type: str, query: str) -> str:
    """Look up CRM entities (individual, account, organization, contact)."""
    from agent.tools.crm_tools import crm_lookup as _lookup
    return await _lookup(entity_type, query)


@function_tool
async def crm_count(entity_type: str) -> str:
    """Count CRM entities of a type."""
    from agent.tools.crm_tools import crm_count as _count
    return await _count(entity_type)


@function_tool
async def crm_list(entity_type: str) -> str:
    """List all CRM entities of a type."""
    from agent.tools.crm_tools import crm_list as _list
    return await _list(entity_type)


# --- Agent Definitions ---

def _create_agents(schema_context: str = ""):
    """Create all agents with current context."""
    # Create workers with their tool sets
    general_worker = create_general_worker(
        model=GPT_5,
        tools=[web_search, read_document, list_documents],
    )

    job_search_worker = create_job_search_worker(
        model=GPT_5,
        tools=[job_search, check_applied_jobs, update_job_status, send_email, web_search],
    )

    cover_letter_worker = create_cover_letter_worker(
        model=GPT_5,
        tools=[read_document, list_documents],
    )

    crm_worker = create_crm_worker(
        model=GPT_5,
        tools=[crm_lookup, crm_count, crm_list, search_entity_documents, read_document, list_documents, web_search],
        schema_context=schema_context,
    )

    # Create triage router with all workers
    return create_triage_router(
        model=CLAUDE_HAIKU,
        workers=[general_worker, job_search_worker, cover_letter_worker, crm_worker],
    )


# Cached agents
_triage_agent = None


async def _get_triage_agent():
    """Get or create triage agent with schema context."""
    global _triage_agent

    if _triage_agent is None:
        schema_context = ""
        try:
            schema = await get_reasoning_schema("crm")
            schema_context = format_schema_context(schema) if schema else ""
        except Exception as e:
            logger.warning(f"Failed to load CRM schema: {e}")

        _triage_agent = _create_agents(schema_context)
        logger.info("Agents initialized")

    return _triage_agent


async def cancel_streaming(thread_id: str) -> bool:
    """Cancel a running streaming task."""
    task = _running_tasks.get(thread_id)
    if task and not task.done():
        logger.info(f"Cancelling task for thread: {thread_id}")
        task.cancel()
        return True
    return False


async def create_orchestrator() -> Dict[str, Callable]:
    """
    Create orchestrator interface (maintains API compatibility).

    Returns dict with:
        - run_streaming: Main execution function
        - set_websocket_sender: Set WebSocket sender for streaming
    """
    logger.info("Initializing SDK-based orchestrator...")

    # Pre-initialize agents
    await _get_triage_agent()

    ws_sender_ref = {"send_ws": None}

    async def run_streaming(
        message: str,
        thread_id: str,
        success_criteria: str = None,
        tenant_id: int = None,
        user_id: int = None,
    ) -> Dict[str, Any]:
        """Run the agent with streaming support."""
        logger.info(f"Starting agent for thread: {thread_id}")
        logger.info(f"   Message: {message[:100]}...")

        current_task = asyncio.current_task()
        if current_task:
            _running_tasks[thread_id] = current_task

        set_tenant_id(tenant_id)
        send_ws = ws_sender_ref["send_ws"]

        # Check for new conversation
        is_new_conversation = False
        thread_uuid = None
        if user_id and tenant_id:
            try:
                thread_uuid = UUID(thread_id)
                existing = await conversation_service.get_conversation(thread_uuid)
            except Exception:
                existing = None
                is_new_conversation = True
            if not existing:
                is_new_conversation = True

            try:
                await save_message(thread_id, user_id, tenant_id, "user", message)
            except Exception as e:
                logger.warning(f"Failed to save user message: {e}")

        # Send start event
        if send_ws:
            await send_ws(json.dumps({"type": "start"}))

        try:
            triage = await _get_triage_agent()

            # Load recent conversation context
            context_parts = []
            if thread_id:
                history = await load_messages(thread_id, max_messages=6)
                for msg in history:
                    role = msg.get("role", "user")
                    content = msg.get("content", "")
                    if role in ("user", "assistant") and content:
                        prefix = "User" if role == "user" else "Assistant"
                        context_parts.append(f"{prefix}: {content[:500]}")

            # Build input with context if available
            if context_parts:
                context_str = "\n".join(context_parts[-6:])  # Last 3 exchanges
                input_message = f"[Recent conversation context]\n{context_str}\n\n[Current request]\n{message}"
            else:
                input_message = message

            # Run with tracing and logging hooks
            with trace(workflow_name="agent", group_id=thread_id):
                result = await Runner.run(
                    triage,
                    input_message,
                    hooks=logging_hooks,
                )

            # Extract response
            final_content = ""
            if isinstance(result.final_output, str):
                final_content = result.final_output
            elif hasattr(result.final_output, "model_dump"):
                final_content = json.dumps(result.final_output.model_dump())
            else:
                final_content = str(result.final_output)

            # Get token usage from result.context_wrapper.usage
            turn_tokens = {"input": 0, "output": 0, "total": 0}
            try:
                usage = result.context_wrapper.usage
                turn_tokens = {
                    "input": usage.input_tokens,
                    "output": usage.output_tokens,
                    "total": usage.total_tokens,
                }
            except (AttributeError, TypeError):
                logger.warning("Could not get token usage from result")

            # Update cumulative totals for this thread
            if thread_id not in _thread_token_totals:
                _thread_token_totals[thread_id] = {"input": 0, "output": 0, "total": 0}
            _thread_token_totals[thread_id]["input"] += turn_tokens["input"]
            _thread_token_totals[thread_id]["output"] += turn_tokens["output"]
            _thread_token_totals[thread_id]["total"] += turn_tokens["total"]
            cumulative_tokens = _thread_token_totals[thread_id]

            # Combined token info for response
            tokens = {
                "turn": turn_tokens,
                "cumulative": cumulative_tokens,
            }

            # Save assistant response
            if user_id and tenant_id and final_content:
                try:
                    await save_message(
                        thread_id, user_id, tenant_id, "assistant", final_content, turn_tokens
                    )
                except Exception as e:
                    logger.warning(f"Failed to save assistant message: {e}")

            # Generate title for new conversations (synchronously so we can send it)
            generated_title = None
            if is_new_conversation and thread_uuid and final_content:
                try:
                    generated_title = await conversation_service.generate_title(message, final_content)
                    await conversation_service.update_title(thread_uuid, generated_title)
                    logger.info(f"Generated title: {generated_title}")
                except Exception as e:
                    logger.warning(f"Failed to generate title: {e}")

            # Record usage with actual worker name
            if tenant_id and turn_tokens.get("total", 0) > 0:
                try:
                    # Get the name of the worker that handled the request
                    worker_name = "agent"
                    if hasattr(result, "last_agent") and result.last_agent:
                        worker_name = result.last_agent.name
                    logger.info(f"Recording usage for worker: {worker_name}")

                    await usage_service.log_usage_records(
                        tenant_id=tenant_id,
                        user_id=user_id,
                        records=[{"model_name": GPT_5, "node_name": worker_name, **turn_tokens}],
                    )
                except Exception as e:
                    logger.warning(f"Failed to record usage: {e}")

            # Send completion
            if send_ws:
                await send_ws(json.dumps({
                    "type": "content",
                    "content": final_content,
                    "is_final": True,
                }))
                logger.info(f"Token usage - turn: {turn_tokens}, cumulative: {cumulative_tokens}")
                complete_payload = {
                    "type": "complete",
                    "final_token_usage": tokens,
                }
                if generated_title:
                    complete_payload["conversation_title"] = generated_title
                await send_ws(json.dumps(complete_payload))

            logger.info("Agent execution completed")
            return {
                "messages": [{"role": "assistant", "content": final_content}],
                "tokens": tokens,
                "done": True,
            }

        except asyncio.CancelledError:
            logger.info(f"Agent cancelled for thread: {thread_id}")
            if send_ws:
                await send_ws(json.dumps({"type": "cancelled"}))
            raise

        except Exception as e:
            logger.error(f"Agent failed: {e}")
            if send_ws:
                await send_ws(json.dumps({"type": "error", "error": str(e)}))
            raise

        finally:
            _running_tasks.pop(thread_id, None)

    def set_websocket_sender(send_ws):
        """Set WebSocket sender for streaming."""
        ws_sender_ref["send_ws"] = send_ws

    logger.info("SDK-based orchestrator initialized")
    return {
        "run_streaming": run_streaming,
        "set_websocket_sender": set_websocket_sender,
    }
