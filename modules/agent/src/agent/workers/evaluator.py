"""
Evaluator node - functional implementation with closure
"""

import logging
from typing import Dict, Any, Type, Annotated
from langchain_core.messages import SystemMessage, HumanMessage, AIMessage, ToolMessage
from langchain_anthropic import ChatAnthropic
from pydantic import BaseModel, Field, create_model
import shutil
import textwrap

logger = logging.getLogger(__name__)

from agent.util.langfuse_helper import get_langfuse_handler


def make_evaluator_output_model() -> Type[BaseModel]:
    """Closure/factory that returns a Pydantic model class for the evaluator output."""
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


def _build_system_prompt(resume_context: str) -> str:
    """Pure function to build evaluator system prompt"""
    base_prompt = (
        "You evaluate whether the Assistant's response meets the success criteria.\n\n"
    )

    # Only include resume-specific instructions if we have resume context
    if resume_context:
        base_prompt += (
            "RESUME_CONTEXT\n"
            f"{resume_context}\n\n"
            "CRITICAL DISTINCTION:\n"
            "- JOB SEARCH RESULTS: When the user asks to find jobs, the assistant will list EXTERNAL job postings "
            "from companies. These are NOT the user's own work history. Job titles like 'Senior AI Engineer at DataFabric' "
            "are VALID if they are job postings the user could apply to, even if they don't appear in the resume.\n"
            "- RESUME DATA: The user's own work history (e.g., 'Principal Engineer at PinaColada.co') must match the resume exactly.\n"
            "- DO NOT confuse job search results (external postings) with the user's resume/work history.\n\n"
            "SPECIAL CHECKS:\n"
            "- For job search responses: Ensure all job links are direct posting URLs (not job board links)\n"
            "- Links should be accessible and relevant to the job listings\n"
            "- Job search results showing external postings are VALID even if those companies/titles aren't in the resume\n\n"
        )

    base_prompt += (
        "Respond with a strict JSON object matching the schema (handled by the tool):\n"
        "- feedback: Brief assessment\n"
        "- success_criteria_met: true/false\n"
        "- user_input_needed: true if user must respond\n\n"
        "Be slightly lenient—approve unless clearly inadequate."
    )

    return base_prompt


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

        # Detect retry loops by counting AIMessages (each retry adds a new AIMessage)
        # If we have feedback_on_work set, we're already retrying
        messages = state.get("messages", [])
        ai_message_count = sum(1 for msg in messages if isinstance(msg, AIMessage))

        # If we have feedback and multiple AI messages, we're in a retry loop
        has_feedback = bool(state.get("feedback_on_work"))
        is_retry_loop = has_feedback and ai_message_count >= 3
        retry_count = (ai_message_count - 1) if is_retry_loop else 0

        # Build prompts
        resume_context = state.get("resume_context", "")
        has_resume = bool(resume_context)
        logger.info(f"📋 Evaluator using prompt {'WITH' if has_resume else 'WITHOUT'} resume context")
        system_message = _build_system_prompt(resume_context)

        user_message = (
            "Recent conversation (condensed):\n"
            f"{_format_conversation(state['messages'])}\n\n"
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
        logger.info("   Calling Claude Haiku 4.5 for evaluation...")
        if is_retry_loop:
            logger.warning(
                f"   ⚠️  Retry loop detected ({retry_count} retries) - being more lenient"
            )

        try:
            eval_result = llm_with_output.invoke(evaluator_messages)

            # If we're in a retry loop and still rejecting, force approval to break the loop
            if (
                is_retry_loop
                and not eval_result.success_criteria_met
                and not eval_result.user_input_needed
            ):
                logger.warning("   ⚠️  Forcing approval to break retry loop")
                eval_result.success_criteria_met = True
                eval_result.feedback = f"{eval_result.feedback} (Approved after {retry_count} retries to prevent infinite loop)"
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
