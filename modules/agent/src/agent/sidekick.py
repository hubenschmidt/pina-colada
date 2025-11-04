from langgraph.checkpoint.memory import MemorySaver
import uuid
import json
import logging
from typing import List, Any, Optional, Dict, Annotated, Callable, Awaitable
from typing_extensions import TypedDict
from agent.tools.sidekick_tools import sidekick_tools
from agent.workers.worker import WorkerNode
from agent.workers.evaluator import EvaluatorNode
from agent.workers.cover_letter_writer import CoverLetterWriterNode
from agent.routers.agent_router import route_to_agent, route_from_router_edge
from langgraph.graph import StateGraph, START, END
from dotenv import load_dotenv
from langgraph.graph.message import add_messages
from langgraph.prebuilt import ToolNode
from langchain_core.messages import (
    AIMessage,
    HumanMessage,
    trim_messages,
)

logger = logging.getLogger(__name__)
load_dotenv(override=True)


class State(TypedDict):
    messages: Annotated[List[Any], add_messages]
    success_criteria: str
    feedback_on_work: Optional[str]
    success_criteria_met: bool
    user_input_needed: bool
    resume_name: str
    resume_context: str
    route_to_agent: Optional[str]


class Sidekick:
    def __init__(
        self,
        resume_text: str = "",
        summary: str = "",
        sample_answers: str = "",
        cover_letters: list = None,
    ):
        self.worker_node = None
        self.evaluator_node = None
        self.cover_letter_writer_node = None
        self.tools = None
        self.graph = None
        self.agent_id = str(uuid.uuid4())
        self.memory = MemorySaver()
        self.send_ws: Optional[Callable[[str], Awaitable[None]]] = None

        # Store document context
        self.resume_text = resume_text
        self.summary = summary
        self.sample_answers = sample_answers
        self.cover_letters = cover_letters or []

        # Build resume context string
        self.resume_context = self._build_resume_context()
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

        if self.resume_text:
            preview = (
                self.resume_text[:500] + "..."
                if len(self.resume_text) > 500
                else self.resume_text
            )
            context_parts.append(f"RESUME (excerpt)\n{preview}")
            context_parts.append("\n[Full resume available via file tools if needed]")

        return "\n\n".join(context_parts)

    async def setup(self):
        """Initialize the nodes and build the graph"""
        logger.info("=== AGENT SETUP ===")

        self.tools = await sidekick_tools()

        # Initialize all nodes
        self.worker_node = WorkerNode(self.tools)
        await self.worker_node.setup(self.resume_context_concise)

        self.evaluator_node = EvaluatorNode()
        await self.evaluator_node.setup()

        self.cover_letter_writer_node = CoverLetterWriterNode()
        await self.cover_letter_writer_node.setup()

        await self.build_graph()
        logger.info(f"✓ Agent initialized with {len(self.tools)} tools")
        logger.info("===================")

    def set_websocket_sender(self, send_ws: Callable[[str], Awaitable[None]]):
        """Set the WebSocket sender for streaming responses"""
        self.send_ws = send_ws

    async def _stream_if_available(self, content: str, is_final: bool = False):
        """Stream content to WebSocket if available"""
        if self.send_ws:
            payload = {"type": "content", "content": content, "is_final": is_final}
            await self.send_ws(json.dumps(payload))

    def _trim_messages(self, messages: List[Any], max_tokens: int = 8000) -> List[Any]:
        """Trim message history to stay under token limit"""
        try:
            trimmed = trim_messages(
                messages,
                max_tokens=max_tokens,
                strategy="last",
                token_counter=len,
                include_system=True,
                allow_partial=False,
            )

            if len(trimmed) < len(messages):
                logger.info(f"Trimmed messages from {len(messages)} to {len(trimmed)}")

            return trimmed
        except Exception as e:
            logger.warning(f"Message trimming failed: {e}, using last 10 messages")
            return messages[-10:]

    # Wrapper methods for nodes
    def router(self, state: State) -> Dict[str, Any]:
        return route_to_agent(state)

    def worker(self, state: State) -> Dict[str, Any]:
        return self.worker_node.execute(state, self._trim_messages)

    def worker_router(self, state: State):
        return self.worker_node.route_from_worker(state)

    def cover_letter_writer(self, state: State) -> Dict[str, Any]:
        return self.cover_letter_writer_node.execute(state, self._trim_messages)

    def evaluator(self, state: State) -> Dict[str, Any]:
        return self.evaluator_node.execute(state)

    def route_based_on_evaluation(self, state: State) -> str:
        return self.evaluator_node.route_from_evaluator(state)

    async def build_graph(self):
        """Build the LangGraph workflow"""
        logger.info("Building LangGraph workflow...")

        try:
            graph_builder = StateGraph(State)

            # Add nodes
            logger.info("Adding nodes to graph...")
            graph_builder.add_node("router", self.router)
            graph_builder.add_node("worker", self.worker)
            graph_builder.add_node("cover_letter_writer", self.cover_letter_writer)
            graph_builder.add_node("tools", ToolNode(tools=self.tools))
            graph_builder.add_node("evaluator", self.evaluator)
            logger.info("✓ Nodes added")

            # Add edges
            logger.info("Adding edges to graph...")

            # Start with router
            graph_builder.add_edge(START, "router")

            # Router decides which agent to use
            graph_builder.add_conditional_edges(
                "router",
                route_from_router_edge,
                {"worker": "worker", "cover_letter_writer": "cover_letter_writer"},
            )

            # Worker can go to tools or evaluator
            graph_builder.add_conditional_edges(
                "worker",
                self.worker_router,
                {"tools": "tools", "evaluator": "evaluator"},
            )

            # Tools go back to worker
            graph_builder.add_edge("tools", "worker")

            # Cover letter writer goes to evaluator
            graph_builder.add_edge("cover_letter_writer", "evaluator")

            # Evaluator can route back to worker, cover_letter_writer, or END
            graph_builder.add_conditional_edges(
                "evaluator",
                self.route_based_on_evaluation,
                {
                    "worker": "worker",
                    "cover_letter_writer": "cover_letter_writer",
                    "END": END,
                },
            )

            logger.info("✓ Edges added")

            # Compile the graph
            logger.info("Compiling graph...")
            self.graph = graph_builder.compile(checkpointer=self.memory)

            if self.graph is None:
                raise RuntimeError("Graph compilation returned None")

            logger.info("✓ Graph compiled successfully")

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
            "route_to_agent": None,
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
        logger.info("Freeing Agent resources")
        self.send_ws = None
