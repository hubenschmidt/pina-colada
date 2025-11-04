import logging
from typing import Dict, Any
from datetime import datetime
from langchain_core.messages import SystemMessage, HumanMessage
from langchain_openai import ChatOpenAI
from langfuse.langchain import CallbackHandler

logger = logging.getLogger(__name__)


class WorkerNode:
    """Worker node that handles general queries with tool access"""

    def __init__(self, tools):
        self.llm_with_tools = None
        self.tools = tools
        self.resume_context_concise = None

    async def setup(self, resume_context_concise: str):
        """Initialize the LLM for worker"""
        logger.info("Setting up Worker LLM: OpenAI GPT-5 (temperature=0.7)")
        langfuse_handler = CallbackHandler()

        worker_llm = ChatOpenAI(
            model="gpt-5-chat-latest",
            temperature=0.7,
            max_completion_tokens=512,
            max_retries=3,
            callbacks=[langfuse_handler],
        )
        self.llm_with_tools = worker_llm.bind_tools(self.tools)
        self.resume_context_concise = resume_context_concise
        logger.info(f"✓ Worker LLM configured with {len(self.tools)} tools")

    def _should_use_full_context(self, message: str) -> bool:
        """Determine if we need the full context or if concise is sufficient"""
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

    def _get_system_prompt(
        self, state: Dict[str, Any], use_full_context: bool, resume_context_concise: str
    ) -> str:
        """Generate system prompt - concise by default, full only when needed"""

        # Choose context based on need
        context = (
            state["resume_context"] if use_full_context else resume_context_concise
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

    def execute(
        self,
        state: Dict[str, Any],
        trim_messages_fn,
    ) -> Dict[str, Any]:
        """Execute the worker node"""
        logger.info("🤖 WORKER NODE: Generating response with GPT-5...")

        # Determine if we need full context
        last_user_message = ""
        for msg in reversed(state["messages"]):
            if isinstance(msg, HumanMessage):
                last_user_message = msg.content
                break

        use_full = self._should_use_full_context(last_user_message)
        logger.info(f"Using {'FULL' if use_full else 'CONCISE'} context")

        system_prompt = self._get_system_prompt(
            state,
            use_full_context=use_full,
            resume_context_concise=self.resume_context_concise,
        )

        if state.get("feedback_on_work"):
            logger.info(f"⚠️  Worker received feedback from evaluator, retrying...")
            system_prompt += f"\n\nEVALUATOR FEEDBACK:\n{state['feedback_on_work']}\n\nPlease improve your response based on this feedback."

        # CRITICAL: Trim message history before sending
        trimmed_messages = trim_messages_fn(state["messages"], max_tokens=6000)

        messages = [SystemMessage(content=system_prompt)] + trimmed_messages

        logger.info(
            f"   Message count: {len(messages)} (trimmed from {len(state['messages'])})"
        )
        logger.info(f"   System prompt length: ~{len(system_prompt)} chars")

        response = self.llm_with_tools.invoke(messages)

        logger.info(f"✓ Worker response generated")
        logger.info(f"   Response length: {len(response.content)} chars")
        logger.info(f"   Has tool calls: {bool(response.tool_calls)}")

        return {"messages": [response]}

    def route_from_worker(self, state: Dict[str, Any]) -> str:
        """Route from worker to either tools or evaluator"""
        last_message = state["messages"][-1]

        if hasattr(last_message, "tool_calls") and last_message.tool_calls:
            logger.info("→ Routing to TOOLS")
            return "tools"

        logger.info("→ Routing to EVALUATOR")
        return "evaluator"
