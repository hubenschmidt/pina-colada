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
    )
    EvaluatorOutput.__doc__ = "Structured output for evaluator"
    return EvaluatorOutput


def format_conversation(messages) -> str:
    """Format conversation history for evaluation"""
    formatted = []
    relevant_messages = messages[-6:] if len(messages) > 6 else messages

    for msg in relevant_messages:
        if isinstance(msg, HumanMessage):
            formatted.append(f"USER: {msg.content}")
            continue
        if isinstance(msg, AIMessage) and msg.content:
            formatted.append(f"ASSISTANT: {msg.content}")
            continue
        if isinstance(msg, ToolMessage):
            formatted.append(f"[Tool executed: {msg.name}]")

    return "\n".join(formatted)


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
            return {
                "feedback_on_work": "No AI response to evaluate.",
                "success_criteria_met": True,
                "user_input_needed": False,
            }

        # Detect retry loops
        messages = state.get("messages", [])
        ai_message_count = sum(1 for msg in messages if isinstance(msg, AIMessage))
        has_feedback = bool(state.get("feedback_on_work"))
        is_retry_loop = has_feedback and ai_message_count >= 3
        retry_count = (ai_message_count - 1) if is_retry_loop else 0

        # Build prompts
        resume_context = state.get("resume_context", "")
        system_message = build_system_prompt(resume_context)

        user_message = (
            "Recent conversation (condensed):\n"
            f"{format_conversation(state['messages'])}\n\n"
            f"Success criteria: {state.get('success_criteria','')}\n\n"
            "Final response to evaluate (from Assistant):\n"
            f"{last_message.content}\n\n"
        )

        if is_retry_loop:
            user_message += (
                "NOTE: This response has been retried multiple times. Be more lenient - "
                "approve unless there are critical errors that make the response unusable.\n\n"
            )

        user_message += "Does this meet the criteria?"

        if state.get("feedback_on_work"):
            user_message += f"\n\nPrior feedback:\n{state['feedback_on_work'][:200]}"

        evaluator_messages = [
            SystemMessage(content=system_message),
            HumanMessage(content=user_message),
        ]

        # Get evaluation
        logger.info(f"   Calling Claude Haiku 4.5 for {evaluator_name} evaluation...")
        if is_retry_loop:
            logger.warning(f"   ⚠️  Retry loop detected ({retry_count} retries)")

        try:
            eval_result = llm_with_output.invoke(evaluator_messages)

            # Force approval to break retry loops
            should_force_approval = (
                is_retry_loop
                and not eval_result.success_criteria_met
                and not eval_result.user_input_needed
            )
            if should_force_approval:
                logger.warning("   ⚠️  Forcing approval to break retry loop")
                eval_result.success_criteria_met = True
                eval_result.feedback = f"{eval_result.feedback} (Approved after {retry_count} retries)"
        except Exception as e:
            logger.error(f"⚠️  Evaluation failed: {e}")
            return {
                "feedback_on_work": "Evaluation error occurred, defaulting to approval.",
                "success_criteria_met": True,
                "user_input_needed": False,
            }

        # Log results
        logger.info("✓ Evaluator decision:")
        logger.info(f"   - Success criteria met: {eval_result.success_criteria_met}")
        logger.info(f"   - User input needed: {eval_result.user_input_needed}")

        width = shutil.get_terminal_size(fallback=(100, 24)).columns
        wrapped = textwrap.fill(
            eval_result.feedback,
            width=width,
            subsequent_indent=" " * 4,
        )
        logger.info("   - Feedback: %s", wrapped)

        if not eval_result.success_criteria_met and not eval_result.user_input_needed:
            logger.warning("⚠️  Evaluator rejected response - agent will retry!")

        return {
            "feedback_on_work": eval_result.feedback,
            "success_criteria_met": eval_result.success_criteria_met,
            "user_input_needed": eval_result.user_input_needed,
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
