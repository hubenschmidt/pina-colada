"""
Cover letter writer node - functional implementation with closure
"""

import logging
from typing import Dict, Any, Callable
from datetime import datetime
from langchain_core.messages import SystemMessage
from langchain_openai import ChatOpenAI

logger = logging.getLogger(__name__)

from langfuse.langchain import CallbackHandler


def _build_system_prompt(resume_name: str, resume_context: str) -> str:
    """Pure function to build cover letter writer system prompt"""
    SIGNATURE = (
        "Warm regards,\n\n"
        "William Hubenschmidt\n\n"
        "whubenschmidt@gmail.com\n"
        "pinacolada.co\n"
        "Brooklyn, NY"
    )
    return f"""You are a professional cover letter writer for {resume_name}.

AVAILABLE CONTEXT:
{resume_context}


YOUR TASK:
Write compelling, professional cover letters that:
1. Are properly formatted (greeting, 2-4 body paragraphs, closing)
2. Reference specific job details from the job description. Do not forget to ask for a job description.
3. Use actual experience and skills from the resume context above
4. Match the professional tone from sample cover letters
5. Are 200-300 words in length
6. ALWAYS Use plain text formatting in every response. No markdown, or illegal characters, or bolded text.
7. Are tailored to the specific job and company
8. Only output the contents of the cover letter

STYLE GUIDELINES:
- Professional but personable tone
- Specific examples over generic claims
- Show enthusiasm for the role
- Connect resume experience to job requirements
- Strong opening and closing

SIGNATURE BLOCK:
- MANDATORY: should always contain {SIGNATURE}
- MANDATORY: insert a blank line between the sign off ("Warm regards" and your name "William Hubenschmidt")

Warm regards,

William Hubenschmidt

whubenschmidt@gmail.com
pinacolada.co
Brooklyn, NY


Date: {datetime.now().strftime("%Y-%m-%d")}
"""


async def create_cover_letter_writer_node(trim_messages_fn: Callable):
    """
    Factory function that creates a cover letter writer node

    Returns a pure function that takes state and returns cover letter
    """
    logger.info("Setting up Cover Letter Writer LLM...")
    langfuse_handler = CallbackHandler()

    llm = ChatOpenAI(
        model="gpt-5-chat-latest",
        temperature=0.7,
        max_completion_tokens=1500,
        max_retries=3,
        callbacks=[langfuse_handler],
    )
    logger.info("✓ Cover Letter Writer LLM configured")

    # Return the actual node function with closed-over LLM
    def cover_letter_writer_node(state: Dict[str, Any]) -> Dict[str, Any]:
        """Pure function: state in -> cover letter out"""
        logger.info("📝 COVER LETTER WRITER NODE: Generating cover letter...")

        # Build system prompt
        system_prompt = _build_system_prompt(
            state["resume_name"], state["resume_context"]
        )

        if state.get("feedback_on_work"):
            logger.info("⚠️  Cover Letter Writer received feedback, retrying...")
            system_prompt += (
                f"\n\nEVALUATOR FEEDBACK:\n{state['feedback_on_work']}\n\n"
                "Please improve the cover letter based on this feedback."
            )

        # Get recent conversation for context
        trimmed_messages = trim_messages_fn(state["messages"], max_tokens=6000)
        messages = [SystemMessage(content=system_prompt)] + trimmed_messages

        logger.info(f"   Message count: {len(messages)}")
        logger.info(f"   System prompt length: ~{len(system_prompt)} chars")

        # Generate cover letter
        response = llm.invoke(messages)

        logger.info("✓ Cover letter generated")
        logger.info(f"   Response length: {len(response.content)} chars")

        return {"messages": [response]}

    return cover_letter_writer_node
