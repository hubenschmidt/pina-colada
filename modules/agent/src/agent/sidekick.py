from langgraph.checkpoint.memory import MemorySaver
import uuid
import json
import logging
from typing import List, Any, Optional, Dict, Annotated, Callable, Awaitable
from typing_extensions import TypedDict
from agent.sidekick_tools import all_tools
from langgraph.graph import StateGraph, START, END
from langchain_openai import ChatOpenAI
from pydantic import BaseModel, Field
from dotenv import load_dotenv
from langgraph.graph.message import add_messages
from langgraph.prebuilt import ToolNode
from datetime import datetime
from langchain_core.messages import (
    AIMessage,
    HumanMessage,
    SystemMessage,
    ToolMessage,
    trim_messages,
)
from langchain_anthropic import ChatAnthropic
from langfuse.langchain import CallbackHandler

logger = logging.getLogger(__name__)
load_dotenv(override=True)


class State(TypedDict):
    messages: Annotated[List[Any], add_messages]
    success_criteria: str
    feedback_on_work: Optional[str]
    success_criteria_met: bool
    user_input_needed: bool
    resume_name: str
    resume_context: str  # Added to hold resume, summary, and cover letters


class EvaluatorOutput(BaseModel):
    feedback: str = Field(description="Feedback on the assistant's response")
    success_criteria_met: bool = Field(
        description="Whether the success criteria have been met"
    )
    user_input_needed: bool = Field(
        description="True if more input is needed from the user, or clarifications, or the assistant is stuck"
    )


class Sidekick:
    def __init__(
        self,
        resume_text: str = "",
        summary: str = "",
        sample_answers: str = "",
        cover_letters: list = None,
    ):
        self.worker_llm_with_tools = None
        self.evaluator_llm_with_output = None
        self.tools = None
        self.graph = None
        self.sidekick_id = str(uuid.uuid4())
        self.memory = MemorySaver()
        self.send_ws: Optional[Callable[[str], Awaitable[None]]] = None

        # Store document context
        self.resume_text = resume_text
        self.summary = summary
        self.sample_answers = sample_answers
        self.cover_letters = cover_letters or []

        # Build resume context string
        self.resume_context = self._build_resume_context()

        # NEW: Create a concise version for most requests
        self.resume_context_concise = self._build_resume_context_concise()

    def _build_resume_context(self) -> str:
        """Build the FULL resume context string from loaded documents"""
        context_parts = []

        if self.summary:
            context_parts.append(f"SUMMARY\n{self.summary}")

        if self.resume_text:
            context_parts.append(f"RESUME\n{self.resume_text}")

        if self.sample_answers:
            context_parts.append(f"SAMPLE_ANSWERS\n{self.sample_answers}")

        if self.cover_letters:
            cover_letters_text = "\n\n".join(self.cover_letters)
            context_parts.append(
                f"COVER_LETTERS (for reference on writing style)\n{cover_letters_text}"
            )

        return "\n\n".join(context_parts)

    def _build_resume_context_concise(self) -> str:
        """Build a CONCISE version with just summary and key facts"""
        context_parts = []

        if self.summary:
            context_parts.append(f"SUMMARY\n{self.summary}")
        logger.info((len(self.resume_text)))

        # Add just the first 500 chars of resume for context
        if self.resume_text:
            preview = (
                self.resume_text[:500] + "..."
                if len(self.resume_text) > 500
                else self.resume_text
            )
            context_parts.append(f"RESUME (excerpt)\n{preview}")
            context_parts.append("\n[Full resume available via file tools if needed]")

        return "\n\n".join(context_parts)

    def _should_use_full_context(self, message: str) -> bool:
        """Determine if we need the full context or if concise is sufficient"""
        # Keywords that indicate we need detailed info
        detailed_keywords = [
            "cover letter",
            "write a letter",
            "detailed",
            "comprehensive",
            "all experience",
            "job search",
            "apply",
        ]

        message_lower = message.lower()
        return any(keyword in message_lower for keyword in detailed_keywords)

    async def setup(self):
        """Initialize the LLMs and build the graph"""
        logger.info("=== SIDEKICK SETUP ===")

        self.tools = await all_tools()
        langfuse_handler = CallbackHandler()

        # Worker LLM setup
        logger.info("Setting up Worker LLM: OpenAI GPT-5 (temperature=0.7)")
        worker_llm = ChatOpenAI(
            model="gpt-5-chat-latest",
            temperature=0.7,
            max_completion_tokens=512,
            max_retries=3,
            callbacks=[langfuse_handler],
        )
        self.worker_llm_with_tools = worker_llm.bind_tools(self.tools)
        logger.info(f"✓ Worker LLM configured with {len(self.tools)} tools")

        # Evaluator LLM setup
        logger.info("Setting up Evaluator LLM: Claude Haiku 4.5 (temperature=0)")
        evaluator_llm = ChatAnthropic(
            model="claude-haiku-4-5-20251001",
            temperature=0,
            max_tokens=1000,
            callbacks=[langfuse_handler],
        )
        self.evaluator_llm_with_output = evaluator_llm.with_structured_output(
            EvaluatorOutput
        )
        logger.info("✓ Evaluator LLM configured")

        await self.build_graph()
        logger.info(f"✓ Sidekick initialized with {len(self.tools)} tools")
        logger.info("===================")

    def set_websocket_sender(self, send_ws: Callable[[str], Awaitable[None]]):
        """Set the WebSocket sender for streaming responses"""
        self.send_ws = send_ws

    async def _stream_if_available(self, content: str, is_final: bool = False):
        """Stream content to WebSocket if available"""
        if self.send_ws:
            payload = {"type": "content", "content": content, "is_final": is_final}
            await self.send_ws(json.dumps(payload))

    def _get_system_prompt(self, state: State, use_full_context: bool = False) -> str:
        """Generate system prompt - concise by default, full only when needed"""

        # Choose context based on need
        context = (
            state["resume_context"] if use_full_context else self.resume_context_concise
        )

        # SIMPLIFIED system prompt
        return f"""ROLE: You are {state['resume_name']} on his website. Answer questions professionally and concisely.

DATA ACCESS:
{context}

TASK: {state['success_criteria']}

STYLE:
- Plain text only (no markdown/formatting)
- Concise responses
- No repeated greetings

INSTRUCTIONS:
- Answer using resume data above
- Use record_user_details for contact info
- Use record_unknown_question if you can't answer
- For job searches, search for matches using the resume on file, search in NYC for jobs posted in the last 7 days, and always return direct posting URLs
- Ask for email after 3rd answer (once only)

Date: {datetime.now().strftime("%Y-%m-%d %H:%M:%S")}
"""

    def _trim_messages(self, messages: List[Any], max_tokens: int = 8000) -> List[Any]:
        """Trim message history to stay under token limit"""
        # Always keep system message and last few exchanges
        try:
            trimmed = trim_messages(
                messages,
                max_tokens=max_tokens,
                strategy="last",  # Keep most recent
                token_counter=len,  # Simple approximation
                include_system=True,
                allow_partial=False,
            )

            if len(trimmed) < len(messages):
                logger.info(f"Trimmed messages from {len(messages)} to {len(trimmed)}")

            return trimmed
        except Exception as e:
            logger.warning(f"Message trimming failed: {e}, using last 10 messages")
            # Fallback: just keep last 10 messages
            return messages[-10:]

    def worker(self, state: State) -> Dict[str, Any]:
        """The worker node that interacts with the LLM and tools"""
        logger.info("🤖 WORKER NODE: Generating response with GPT-5...")

        # Determine if we need full context
        last_user_message = ""
        for msg in reversed(state["messages"]):
            if isinstance(msg, HumanMessage):
                last_user_message = msg.content
                break

        use_full = self._should_use_full_context(last_user_message)
        logger.info(f"Using {'FULL' if use_full else 'CONCISE'} context")

        system_prompt = self._get_system_prompt(state, use_full_context=use_full)

        if state.get("feedback_on_work"):
            logger.info(f"⚠️  Worker received feedback from evaluator, retrying...")
            system_prompt += f"\n\nEVALUATOR FEEDBACK:\n{state['feedback_on_work']}\n\nPlease improve your response based on this feedback."

        # CRITICAL: Trim message history before sending
        trimmed_messages = self._trim_messages(state["messages"], max_tokens=6000)

        messages = [SystemMessage(content=system_prompt)] + trimmed_messages

        logger.info(
            f"   Message count: {len(messages)} (trimmed from {len(state['messages'])})"
        )
        logger.info(f"   System prompt length: ~{len(system_prompt)} chars")

        response = self.worker_llm_with_tools.invoke(messages)

        logger.info(f"✓ Worker response generated")
        logger.info(f"   Response length: {len(response.content)} chars")
        logger.info(f"   Has tool calls: {bool(response.tool_calls)}")

        return {"messages": [response]}

    def worker_router(self, state: State):
        """Route from worker to either tools or evaluator"""
        last_message = state["messages"][-1]

        if hasattr(last_message, "tool_calls") and last_message.tool_calls:
            logger.info("→ Routing to TOOLS")
            return "tools"

        logger.info("→ Routing to EVALUATOR")
        return "evaluator"

    def format_conversation(self, messages: List[Any]) -> str:
        """Format conversation for evaluator - KEEP CONCISE"""
        formatted = []

        # Only show last 5 exchanges to evaluator
        recent_messages = messages[-10:]  # Last 10 messages = ~5 exchanges

        for msg in recent_messages:
            if isinstance(msg, HumanMessage):
                formatted.append(f"USER: {msg.content}")
            elif isinstance(msg, AIMessage):
                if msg.content:
                    formatted.append(f"ASSISTANT: {msg.content}")
            elif isinstance(msg, ToolMessage):
                # Don't show full tool output, just that it was called
                formatted.append(f"[Tool executed: {msg.name}]")

        return "\n".join(formatted)

    def evaluator(self, state: State) -> Dict[str, Any]:
        """The evaluator node that checks if the response meets criteria"""
        logger.info("🔍 EVALUATOR NODE: Reviewing response with Claude Haiku...")

        last_message = state["messages"][-1]

        if not isinstance(last_message, AIMessage):
            logger.warning("⚠️  Last message is not an AI message, skipping evaluation")
            return {
                "success_criteria_met": True,
                "user_input_needed": False,
            }

        last_response = last_message.content

        # CONCISE system message for evaluator
        system_message = """You evaluate if the Assistant's response meets the success criteria.
        
Respond with:
- feedback: Brief assessment
- success_criteria_met: true/false
- user_input_needed: true if user must respond

Be lenient - approve responses unless clearly inadequate."""

        # CONCISE user message - only recent context
        user_message = f"""Recent conversation:
{self.format_conversation(state['messages'])}

Success criteria: {state['success_criteria']}

Final response to evaluate:
{last_response}

Does this meet the criteria?"""

        if state.get("feedback_on_work"):
            user_message += f"\n\nPrior feedback: {state['feedback_on_work'][:200]}"

        evaluator_messages = [
            SystemMessage(content=system_message),
            HumanMessage(content=user_message),
        ]

        logger.info("   Calling Claude Haiku 4.5 for evaluation...")
        eval_result = self.evaluator_llm_with_output.invoke(evaluator_messages)

        # Log evaluation results
        logger.info(f"✓ Evaluator decision:")
        logger.info(f"   - Success criteria met: {eval_result.success_criteria_met}")
        logger.info(f"   - User input needed: {eval_result.user_input_needed}")
        logger.info(f"   - Feedback: {eval_result.feedback[:100]}...")

        if not eval_result.success_criteria_met and not eval_result.user_input_needed:
            logger.warning("⚠️  Evaluator rejected response - worker will retry!")

        return {
            "feedback_on_work": eval_result.feedback,
            "success_criteria_met": eval_result.success_criteria_met,
            "user_input_needed": eval_result.user_input_needed,
        }

    def route_based_on_evaluation(self, state: State) -> str:
        """Route based on evaluation results"""
        if state["success_criteria_met"] or state["user_input_needed"]:
            logger.info("→ Routing to END (response approved)")
            return "END"

        logger.info("→ Routing back to WORKER (needs improvement)")
        return "worker"

    async def build_graph(self):
        """Build the LangGraph workflow"""
        logger.info("Building LangGraph workflow...")

        try:

            def worker_traced(state: State) -> Dict[str, Any]:
                return self.worker(state)

            def evaluator_traced(state: State) -> Dict[str, Any]:
                return self.evaluator(state)

            logger.info("✓ Langfuse tracing enabled for worker and evaluator")
            use_tracing = True
        except Exception as e:
            logger.warning(f"Could not enable Langfuse tracing: {e}")
            # Fallback to non-traced versions
            worker_traced = self.worker
            evaluator_traced = self.evaluator
            use_tracing = False

        try:
            graph_builder = StateGraph(State)

            # Add nodes
            logger.info("Adding nodes to graph...")
            graph_builder.add_node("worker", worker_traced)
            graph_builder.add_node("tools", ToolNode(tools=self.tools))
            graph_builder.add_node("evaluator", evaluator_traced)
            logger.info("✓ Nodes added")

            # Add edges
            logger.info("Adding edges to graph...")
            graph_builder.add_conditional_edges(
                "worker",
                self.worker_router,
                {"tools": "tools", "evaluator": "evaluator"},
            )
            graph_builder.add_edge("tools", "worker")
            graph_builder.add_conditional_edges(
                "evaluator",
                self.route_based_on_evaluation,
                {"worker": "worker", "END": END},
            )
            graph_builder.add_edge(START, "worker")
            logger.info("✓ Edges added")

            # Compile the graph
            logger.info("Compiling graph...")
            self.graph = graph_builder.compile(checkpointer=self.memory)

            if self.graph is None:
                raise RuntimeError("Graph compilation returned None")

            logger.info(
                f"✓ Graph compiled successfully (tracing={'enabled' if use_tracing else 'disabled'})"
            )

        except Exception as e:
            logger.error(f"Failed to build graph: {e}", exc_info=True)
            raise

    async def run_streaming(
        self, message: str, thread_id: str, success_criteria: str = None
    ) -> Dict[str, Any]:
        """Run the graph with streaming support for WebSocket"""
        logger.info(f"▶️  Starting graph execution for thread: {thread_id}")
        logger.info(f"   User message: {message[:100]}...")

        config = {"configurable": {"thread_id": thread_id}}

        sc = success_criteria
        if not sc:
            sc = "Provide a clear and accurate response to the user's question"

        state = {
            "messages": [HumanMessage(content=message)],
            "success_criteria": sc,
            "feedback_on_work": None,
            "success_criteria_met": False,
            "user_input_needed": False,
            "resume_name": "William Hubenschmidt",
            "resume_context": self.resume_context,
        }

        # Guard: no websocket sender available → non-streaming path
        if not self.send_ws:
            result = await self.graph.ainvoke(state, config=config)
            logger.info("✅ Graph execution completed")
            return result

        # Streaming path
        await self.send_ws(json.dumps({"type": "start"}))

        last_content = ""
        iteration = 0

        async for event in self.graph.astream(
            state, config=config, stream_mode="values"
        ):
            iteration += 1
            logger.info(f"   Graph iteration {iteration}")

            messages = event.get("messages")
            last_msg = None
            try:
                last_msg = (messages or [])[-1]
            except Exception:
                last_msg = None

            current_content = getattr(last_msg, "content", None)
            is_ai = isinstance(last_msg, AIMessage)
            changed = (
                isinstance(current_content, str)
                and current_content
                and current_content != last_content
            )
            should_emit = is_ai and changed

            if should_emit:
                last_content = current_content
                await self.send_ws(
                    json.dumps(
                        {
                            "type": "content",
                            "content": current_content,
                            "is_final": False,
                        }
                    )
                )

        await self.send_ws(
            json.dumps({"type": "content", "content": last_content, "is_final": True})
        )
        await self.send_ws(json.dumps({"type": "end"}))
        logger.info(f"✅ Graph execution completed in {iteration} iterations")

        return state

    def free_resources(self):
        """Clean up resources"""
        logger.info("Freeing Sidekick resources")
        self.send_ws = None
