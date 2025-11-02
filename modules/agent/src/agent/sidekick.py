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
from langchain_core.messages import AIMessage, HumanMessage, SystemMessage, ToolMessage
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

    def _build_resume_context(self) -> str:
        """Build the resume context string from loaded documents"""
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

    async def setup(self):
        """Initialize the LLMs and build the graph"""
        logger.info("=== SIDEKICK SETUP ===")

        self.tools = await all_tools()
        langfuse_handler = CallbackHandler()

        # Worker LLM setup
        logger.info("Setting up Worker LLM: OpenAI GPT-5 (temperature=0.7)")
        worker_llm = ChatOpenAI(
            model="gpt-5-chat-latest", temperature=0.7, callbacks=[langfuse_handler]
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

    def worker(self, state: State) -> Dict[str, Any]:
        """The worker node that interacts with the LLM and tools"""
        logger.info("🤖 WORKER NODE: Generating response with GPT-5...")

        system_message = f"""
ROLE
You are {state['resume_name']} on his personal website. Answer questions about his career, background, skills, and experience. Speak as {state['resume_name']}, professionally and succinctly.

You have access to detailed information about {state['resume_name']}'s background below. Use this information to answer questions accurately.

DATA ACCESS
{state['resume_context']}

CURRENT TASK
{state['success_criteria']}

STYLE (MUST FOLLOW)
- Output must be plain text only (no Markdown or other formatting).
- Do not use asterisks, underscores, backticks, tildes, hashes, brackets, angle brackets, emojis, or decorative characters.
- Do not bold, italicize, add headings, lists, tables, code fences, links, or inline formatting.
- Keep responses concise; short paragraphs separated by newlines only.
- Do not re-greet mid-conversation. Do not say "Hi there!" again after the first turn.
- Do not use em dashes (--) when responding or writing cover letters.

INSTRUCTIONS
- Use the resume data above to answer questions
- Be conversational but professional
- Keep responses focused and concise
- If the user provides contact information, use the record_user_details tool
- If you cannot answer a question with the available information, use record_unknown_question tool

JOB SEARCH
- When prompted to job search, search for roles matching your resume and experience. Focus on full-stack engineering roles or AI roles.
- CRITICAL: Always return DIRECT LINKS to specific job postings, not general career pages. Each URL must go directly to the individual job listing page where the user can apply.
- Use the search tool to find current job postings and extract the exact posting URL (e.g., "https://company.com/careers")
- If you cannot find the direct posting URL, use the search tool again with more specific queries like "[company name] [job title] apply link"
- Format as: "Company - Job Title: [FULL_URL]" where FULL_URL is the complete, clickable link to that specific posting
- Prioritize recent job postings (within the last 7 days)
- Search for roles in New York City.
- Prompt the user if you wish to expand the search to different criteria, but always remain within New York City.
- If the user asks you to conduct a job search, do not prompt them for their email or contact information.

CONTACT CAPTURE (ASK ONCE, RECORD ONCE) 
- After your third substantive answer, briefly ask once for the user's name and email for follow-up. 
- If an email is provided (with or without a name), immediately call record_user_details with provided fields. 
- If a name is provided but no email, ask once for the email. If declined, acknowledge and continue. Do not ask again once asked or once recorded. 

UNKNOWN ANSWERS 
- If you do not know an answer, call record_unknown_question with the question, then (only if not already collected) ask once for name/email as above. 

SCOPE CONTROL 
- If a request is not about {state['resume_name']}'s career, background, skills, or experience, briefly steer back to those topics. 

COVER LETTER WORKFLOW 
- If asked to write a cover letter, ask for a job-posting URL, read it, then write the letter directly using COVER_LETTERS and SAMPLE_ANSWERS for style and content. 
- Do not ask for confirmation.
- Do not use em dashes.

The current date and time is {datetime.now().strftime("%Y-%m-%d %H:%M:%S")}
"""

        if state.get("feedback_on_work"):
            logger.info(f"⚠️  Worker received feedback from evaluator, retrying...")
            logger.info(f"   Feedback: {state['feedback_on_work'][:100]}...")
            system_message += f"""

FEEDBACK
Previously you thought you answered the question, but your reply was rejected because the success criteria was not met.
Here is the feedback on why this was rejected:
{state['feedback_on_work']}

With this feedback, please continue to try to answer the question, ensuring that you meet the success criteria or have a question for the user.
"""

        # Update or add system message
        messages = state["messages"].copy()

        sys_idx = next(
            (i for i, m in enumerate(messages) if isinstance(m, SystemMessage)), None
        )
        if sys_idx is not None:
            messages[sys_idx] = SystemMessage(content=system_message)

        if sys_idx is None:
            messages = [SystemMessage(content=system_message)] + messages

        # Invoke the LLM with tools
        logger.info(f"   Calling GPT-5 with {len(messages)} messages...")
        response = self.worker_llm_with_tools.invoke(messages)

        def tool_name(tc):
            # 1) attribute: name
            name_attr = getattr(tc, "name", None)
            if name_attr:
                return name_attr

            # 2) attribute: function.name (safe getattr chain; no ternary)
            fn_name_attr = getattr(getattr(tc, "function", None), "name", None)
            if fn_name_attr:
                return fn_name_attr

            # 3) dict: name
            if not isinstance(tc, dict):
                return "unknown"
            name_dict = tc.get("name")
            if name_dict:
                return name_dict

            # 4) dict: function.name
            fn_dict = tc.get("function")
            if not isinstance(fn_dict, dict):
                return "unknown"
            fn_name_dict = fn_dict.get("name")
            if fn_name_dict:
                return fn_name_dict

            return "unknown"

        tool_names = [tool_name(tc) for tc in response.tool_calls]
        logger.info(f"✓ Worker called {len(response.tool_calls)} tool(s): {tool_names}")
        return {"messages": [response]}

    def worker_router(self, state: State) -> str:
        """Route based on whether tools need to be called"""
        last_message = state["messages"][-1]

        if hasattr(last_message, "tool_calls") and last_message.tool_calls:
            logger.info("→ Routing to TOOLS node")
            return "tools"

        logger.info("→ Routing to EVALUATOR node")
        return "evaluator"

    def format_conversation(self, messages: List[Any]) -> str:
        """Format conversation history for the evaluator"""
        conversation = "Conversation history:\n\n"
        for message in messages:
            if isinstance(message, HumanMessage):
                conversation += f"User: {message.content}\n"
            elif isinstance(message, AIMessage):
                text = message.content or "[Tool use]"
                conversation += f"Assistant: {text}\n"
            elif isinstance(message, ToolMessage):
                conversation += f"[Tool result: {message.content[:100]}...]\n"
        return conversation

    def evaluator(self, state: State) -> Dict[str, Any]:
        """Evaluate if the task is complete"""
        logger.info(
            "🎯 EVALUATOR NODE: Checking response quality with Claude Haiku 4.5..."
        )

        last_response = state["messages"][-1].content
        response_preview = last_response[:100] if last_response else "[empty]"
        logger.info(f"   Evaluating: {response_preview}...")

        system_message = """You are an evaluator that determines if a task has been completed successfully by an Assistant.
Assess the Assistant's last response based on the given criteria. Respond with your feedback, and with your decision on whether the success criteria has been met,
and whether more input is needed from the user."""

        user_message = f"""You are evaluating a conversation between the User and Assistant. You decide what action to take based on the last response from the Assistant.

The entire conversation with the assistant, with the user's original request and all replies, is:
{self.format_conversation(state['messages'])}

The success criteria for this assignment is:
{state['success_criteria']}

And the final response from the Assistant that you are evaluating is:
{last_response}

Respond with your feedback, and decide if the success criteria is met by this response.
Also, decide if more user input is required, either because the assistant has a question, needs clarification, or seems to be stuck and unable to answer without help.

The Assistant has access to tools to record user details and unknown questions. Give the Assistant the benefit of the doubt if they say they've done something.
"""

        if state.get("feedback_on_work"):
            user_message += f"\nNote: In a prior attempt, you provided this feedback: {state['feedback_on_work']}\n"
            user_message += "If you're seeing the Assistant repeating the same mistakes, consider responding that user input is required."

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
