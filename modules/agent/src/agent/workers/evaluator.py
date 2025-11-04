import logging
from typing import Dict, Any
from langchain_core.messages import SystemMessage, HumanMessage, AIMessage, ToolMessage
from langchain_anthropic import ChatAnthropic
from pydantic import BaseModel, Field
from langfuse.langchain import CallbackHandler
from agent.util.logging_config import log_wrapped
import shutil, textwrap

logger = logging.getLogger(__name__)


class EvaluatorOutput(BaseModel):
    feedback: str = Field(description="Feedback on the assistant's response")
    success_criteria_met: bool = Field(
        description="Whether the success criteria have been met"
    )
    user_input_needed: bool = Field(
        description="True if more input is needed from the user, or clarifications, or the assistant is stuck"
    )


class EvaluatorNode:
    """Evaluator node that checks if responses meet success criteria"""

    def __init__(self):
        self.llm_with_output = None

    async def setup(self):
        """Initialize the LLM for evaluator"""
        logger.info("Setting up Evaluator LLM: Claude Haiku 4.5 (temperature=0)")
        langfuse_handler = CallbackHandler()

        evaluator_llm = ChatAnthropic(
            model="claude-haiku-4-5-20251001",
            temperature=0,
            max_tokens=1000,
            callbacks=[langfuse_handler],
        )
        self.llm_with_output = evaluator_llm.with_structured_output(EvaluatorOutput)
        logger.info("✓ Evaluator LLM configured")

    def _get_system_prompt(self, resume_context: str) -> str:
        """Generate system prompt for the evaluator"""
        return (
            "You evaluate whether the Assistant's response meets the success criteria.\n\n"
            "RESUME_CONTEXT\n"
            f"{resume_context}\n\n"
            "Respond with a strict JSON object matching the schema (handled by the tool):\n"
            "- feedback: Brief assessment\n"
            "- success_criteria_met: true/false\n"
            "- user_input_needed: true if user must respond\n\n"
            "Be slightly lenient—approve unless clearly inadequate."
        )

    def _format_conversation(self, messages) -> str:
        """Format conversation for evaluator - KEEP CONCISE"""
        formatted = []
        relevant_messages = messages[-6:] if len(messages) > 6 else messages

        for msg in relevant_messages:
            if isinstance(msg, HumanMessage):
                formatted.append(f"USER: {msg.content}")
            elif isinstance(msg, AIMessage):
                if msg.content:
                    formatted.append(f"ASSISTANT: {msg.content}")
            elif isinstance(msg, ToolMessage):
                formatted.append(f"[Tool executed: {msg.name}]")

        return "\n".join(formatted)

    def execute(self, state: Dict[str, Any]) -> Dict[str, Any]:
        """Execute the evaluator node"""
        logger.info("🔍 EVALUATOR NODE: Reviewing response with Claude Haiku...")

        last_message = state["messages"][-1]
        if not isinstance(last_message, AIMessage):
            logger.warning("⚠️  Last message is not an AI message, skipping evaluation")
            return {
                "feedback_on_work": "No AI response to evaluate.",
                "success_criteria_met": True,
                "user_input_needed": False,
            }

        # Build prompts
        system_message = self._get_system_prompt(state.get("resume_context", ""))

        user_message = (
            "Recent conversation (condensed):\n"
            f"{self._format_conversation(state['messages'])}\n\n"
            f"Success criteria: {state.get('success_criteria','')}\n\n"
            "Final response to evaluate (from Assistant):\n"
            f"{last_message.content}\n\n"
            "Does this meet the criteria?"
        )

        if state.get("feedback_on_work"):
            user_message += f"\n\nPrior feedback:\n{state['feedback_on_work'][:200]}"

        evaluator_messages = [
            SystemMessage(content=system_message),
            HumanMessage(content=user_message),
        ]

        logger.info("   Calling Claude Haiku 4.5 for evaluation...")
        eval_result = self.llm_with_output.invoke(evaluator_messages)

        # Log evaluation results
        logger.info("✓ Evaluator decision:")
        logger.info(f"   - Success criteria met: {eval_result.success_criteria_met}")
        logger.info(f"   - User input needed: {eval_result.user_input_needed}")
        width = shutil.get_terminal_size(
            fallback=(100, 24)
        ).columns  # docker often needs a fallback
        wrapped = textwrap.fill(
            eval_result.feedback,
            width=width,
            subsequent_indent=" " * 4,  # lines up under your "   - Feedback: "
        )

        logger.info("   - Feedback: %s", wrapped)

        if not eval_result.success_criteria_met and not eval_result.user_input_needed:
            logger.warning("⚠️  Evaluator rejected response - agent will retry!")

        return {
            "feedback_on_work": eval_result.feedback,
            "success_criteria_met": eval_result.success_criteria_met,
            "user_input_needed": eval_result.user_input_needed,
        }

    def route_from_evaluator(self, state: Dict[str, Any]) -> str:
        """Route based on evaluation results - back to the agent that needs improvement"""
        if state["success_criteria_met"] or state["user_input_needed"]:
            logger.info("→ Routing to END (response approved)")
            return "END"

        # Route back to whichever agent was working on this
        route_to = state.get("route_to_agent", "worker")
        logger.info(f"→ Routing back to {route_to.upper()} (needs improvement)")
        return route_to
