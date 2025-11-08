"""
Evaluator node - functional implementation with closure
"""

import logging
from typing import Dict, Any, Type, Annotated
from langchain_core.messages import SystemMessage, HumanMessage, AIMessage, ToolMessage
from langchain_anthropic import ChatAnthropic
from pydantic import BaseModel, Field, create_model
from langfuse.langchain import CallbackHandler
import shutil
import textwrap

logger = logging.getLogger(__name__)


def make_evaluator_output_model() -> Type[BaseModel]:
    """Closure/factory that returns a Pydantic model class for the evaluator output."""
    EvaluatorOutput = create_model(
        "EvaluatorOutput",
        feedback=(
            Annotated[str, Field(description="Feedback on the assistant's response")],
            ...,
        ),
        success_criteria_met=(
            Annotated[bool, Field(description="Whether the success criteria have been met")],
            ...,
        ),
        user_input_needed=(
            Annotated[bool, Field(description="True if more input is needed from the user")],
            ...,
        ),
    )
    EvaluatorOutput.__doc__ = "Structured output for evaluator"
    return EvaluatorOutput


def _build_system_prompt(resume_context: str) -> str:
    """Pure function to build evaluator system prompt"""
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


def _format_conversation(messages) -> str:
    """Pure function to format conversation history"""
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


async def create_evaluator_node():
    """
    Factory function that creates an evaluator node

    Returns a pure function that takes state and returns evaluation results
    """
    logger.info("Setting up Evaluator LLM: Claude Haiku 4.5 (temperature=0)")
    langfuse_handler = CallbackHandler()

    evaluator_llm = ChatAnthropic(
        model="claude-haiku-4-5-20251001",
        temperature=0,
        max_tokens=1000,
        callbacks=[langfuse_handler],
    )
    EvaluatorOutput = make_evaluator_output_model()
    llm_with_output = evaluator_llm.with_structured_output(EvaluatorOutput)
    logger.info("✓ Evaluator LLM configured")

    # Return the actual node function with closed-over LLM
    def evaluator_node(state: Dict[str, Any]) -> Dict[str, Any]:
        """Pure function: state in -> evaluation updates out"""
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
        system_message = _build_system_prompt(state.get("resume_context", ""))

        user_message = (
            "Recent conversation (condensed):\n"
            f"{_format_conversation(state['messages'])}\n\n"
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

        # Get evaluation
        logger.info("   Calling Claude Haiku 4.5 for evaluation...")
        try:
            eval_result = llm_with_output.invoke(evaluator_messages)
        except Exception as e:
            logger.error(f"⚠️  Evaluation failed: {e}")
            logger.warning("   Falling back to default evaluation (approving response)")
            # Return a default evaluation that approves the response
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

    # Route back to whichever agent was working on this
    route_to = state.get("route_to_agent", "worker")
    logger.info(f"→ Routing back to {route_to.upper()} (needs improvement)")
    return route_to
