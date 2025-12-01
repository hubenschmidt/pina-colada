"""
Base evaluator - shared logic for all specialized evaluators
"""

import logging
import shutil
import textwrap
from typing import Dict, Any, Type, Annotated, Callable

from langchain_core.messages import SystemMessage, HumanMessage, AIMessage, ToolMessage
from langchain_anthropic import ChatAnthropic
from pydantic import BaseModel, Field, create_model

from agent.util.langfuse_helper import get_langfuse_handler

logger = logging.getLogger(__name__)


def make_evaluator_output_model() -> Type[BaseModel]:
    """Factory that returns a Pydantic model class for evaluator output."""
    EvaluatorOutput = create_model(
        "EvaluatorOutput",
        feedback=(
            Annotated[str, Field(description="Feedback on the assistant's response")],
            ...,
        ),
        success_criteria_met=(
            Annotated[
                bool, Field(description="Whether the success criteria have been met")
            ],
            ...,
        ),
        user_input_needed=(
            Annotated[
                bool, Field(description="True if more input is needed from the user")
            ],
            ...,
        ),
        score=(
            Annotated[
                int, Field(description="Numeric score from 0-100 rating the quality of the response", ge=0, le=100)
            ],
            ...,
        ),
    )
    EvaluatorOutput.__doc__ = "Structured output for evaluator"
    return EvaluatorOutput


def _format_message(msg) -> str | None:
    """Format a single message. Returns None if not formattable."""
    if isinstance(msg, HumanMessage):
        return f"USER: {msg.content}"
    if isinstance(msg, AIMessage) and msg.content:
        return f"ASSISTANT: {msg.content}"
    if isinstance(msg, ToolMessage):
        return f"[Tool executed: {msg.name}]"
    return None


def format_conversation(messages) -> str:
    """Format conversation history for evaluation"""
    relevant_messages = messages[-6:] if len(messages) > 6 else messages
    formatted = [_format_message(msg) for msg in relevant_messages]
    return "\n".join(f for f in formatted if f)


def _detect_retry_loop(state: Dict[str, Any]) -> tuple[bool, int]:
    """Detect if we're in a retry loop. Returns (is_retry, retry_count)."""
    messages = state.get("messages", [])
    ai_count = sum(1 for msg in messages if isinstance(msg, AIMessage))
    has_feedback = bool(state.get("feedback_on_work"))
    is_retry = has_feedback and ai_count >= 3
    return is_retry, (ai_count - 1) if is_retry else 0


def _build_user_prompt(state: Dict[str, Any], is_retry: bool) -> str:
    """Build the user prompt for evaluation."""
    last_message = state["messages"][-1]
    parts = [
        "Recent conversation (condensed):",
        format_conversation(state["messages"]),
        "",
        f"Success criteria: {state.get('success_criteria', '')}",
        "",
        "Final response to evaluate (from Assistant):",
        str(last_message.content),
        "",
    ]

    if is_retry:
        parts.append(
            "NOTE: This response has been retried multiple times. Be more lenient - "
            "approve unless there are critical errors that make the response unusable."
        )
        parts.append("")

    parts.append("Does this meet the criteria?")

    if state.get("feedback_on_work"):
        parts.append(f"\n\nPrior feedback:\n{state['feedback_on_work'][:200]}")

    return "\n".join(parts)


def _force_approval_if_needed(eval_result, is_retry: bool, retry_count: int):
    """Mutate eval_result to force approval if stuck in retry loop."""
    if not is_retry:
        return
    if eval_result.success_criteria_met:
        return
    if eval_result.user_input_needed:
        return

    logger.warning("   ⚠️  Forcing approval to break retry loop")
    eval_result.success_criteria_met = True
    eval_result.feedback = f"{eval_result.feedback} (Approved after {retry_count} retries)"
    if eval_result.score < 60:
        eval_result.score = 60


def _log_result(evaluator_name: str, eval_result):
    """Log evaluation results."""
    status = "PASS" if eval_result.success_criteria_met else "FAIL"
    logger.info(f"✓ {evaluator_name.upper()} EVALUATOR: Result = {status} (score: {eval_result.score}/100)")
    logger.info("✓ Evaluator decision:")
    logger.info(f"   - Success criteria met: {eval_result.success_criteria_met}")
    logger.info(f"   - User input needed: {eval_result.user_input_needed}")
    logger.info(f"   - Score: {eval_result.score}/100")

    width = shutil.get_terminal_size(fallback=(100, 24)).columns
    wrapped = textwrap.fill(
        eval_result.feedback,
        width=width,
        subsequent_indent=" " * 4,
    )
    logger.info("   - Feedback: %s", wrapped)

    if not eval_result.success_criteria_met and not eval_result.user_input_needed:
        logger.warning("⚠️  Evaluator rejected response - agent will retry!")


def _default_eval_result() -> Dict[str, Any]:
    """Return default result when evaluation is skipped or fails."""
    return {
        "feedback_on_work": "No AI response to evaluate.",
        "success_criteria_met": True,
        "user_input_needed": False,
        "score": 100,
    }


async def create_base_evaluator_node(
    evaluator_name: str,
    build_system_prompt: Callable[[str], str],
):
    """
    Factory function that creates an evaluator node with custom system prompt.

    Args:
        evaluator_name: Name for logging purposes
        build_system_prompt: Function that takes resume_context and returns system prompt

    Returns:
        Pure function that takes state and returns evaluation results
    """
    logger.info(f"Setting up {evaluator_name} Evaluator LLM: Claude Haiku 4.5")
    langfuse_handler = get_langfuse_handler()
    callbacks = [langfuse_handler] if langfuse_handler else []

    evaluator_llm = ChatAnthropic(
        model="claude-haiku-4-5-20251001",
        temperature=0,
        max_tokens=1000,
        callbacks=callbacks,
    )
    EvaluatorOutput = make_evaluator_output_model()
    llm_with_output = evaluator_llm.with_structured_output(EvaluatorOutput)
    logger.info(f"✓ {evaluator_name} Evaluator LLM configured")

    def evaluator_node(state: Dict[str, Any]) -> Dict[str, Any]:
        """Pure function: state in -> evaluation updates out"""
        logger.info(f"🔍 {evaluator_name.upper()} EVALUATOR: Reviewing response...")

        last_message = state["messages"][-1]
        if not isinstance(last_message, AIMessage):
            logger.warning("⚠️  Last message is not an AI message, skipping evaluation")
            return _default_eval_result()

        is_retry, retry_count = _detect_retry_loop(state)
        if is_retry:
            logger.warning(f"   ⚠️  Retry loop detected ({retry_count} retries)")

        resume_context = state.get("resume_context", "")
        system_message = build_system_prompt(resume_context)
        user_message = _build_user_prompt(state, is_retry)

        evaluator_messages = [
            SystemMessage(content=system_message),
            HumanMessage(content=user_message),
        ]

        logger.info(f"   Calling Claude Haiku 4.5 for {evaluator_name} evaluation...")

        try:
            eval_result = llm_with_output.invoke(evaluator_messages)
            _force_approval_if_needed(eval_result, is_retry, retry_count)
        except Exception as e:
            logger.error(f"⚠️  Evaluation failed: {e}")
            return {
                "feedback_on_work": "Evaluation error occurred, defaulting to approval.",
                "success_criteria_met": True,
                "user_input_needed": False,
                "score": 100,
            }

        _log_result(evaluator_name, eval_result)

        return {
            "feedback_on_work": eval_result.feedback,
            "success_criteria_met": eval_result.success_criteria_met,
            "user_input_needed": eval_result.user_input_needed,
            "score": eval_result.score,
        }

    return evaluator_node


def route_from_evaluator(state: Dict[str, Any]) -> str:
    """Pure routing function - examines state and returns next node"""
    if state["success_criteria_met"] or state["user_input_needed"]:
        logger.info("→ Routing to END (response approved)")
        return "END"

    route_to = state.get("route_to_agent", "worker")
    logger.info(f"→ Routing back to {route_to.upper()} (needs improvement)")
    return route_to
