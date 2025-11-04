import logging
from typing import Dict, Any
from datetime import datetime
from langchain_core.messages import SystemMessage
from langchain_openai import ChatOpenAI
from langfuse.langchain import CallbackHandler

logger = logging.getLogger(__name__)


class CoverLetterWriterNode:
    """Cover letter writer node for the Sidekick graph"""

    def __init__(self):
        self.llm = None

    async def setup(self):
        """Initialize the LLM for cover letter writing"""
        logger.info("Setting up Cover Letter Writer LLM...")
        langfuse_handler = CallbackHandler()

        self.llm = ChatOpenAI(
            model="gpt-5-chat-latest",
            temperature=0.7,
            max_completion_tokens=1500,
            max_retries=3,
            callbacks=[langfuse_handler],
        )
        logger.info("✓ Cover Letter Writer LLM configured")

    def _get_system_prompt(self, resume_name: str, resume_context: str) -> str:
        """Generate system prompt for cover letter writing"""
        return f"""You are a professional cover letter writer for {resume_name}.

AVAILABLE CONTEXT:
{resume_context}

YOUR TASK:
Write compelling, professional cover letters in Russian that:
1. Are properly formatted (greeting, 2-4 body paragraphs, closing)
2. Reference specific job details from the conversation
3. Use actual experience and skills from the resume context above
4. Match the professional tone from sample cover letters
5. Are 250-400 words in length
6. Use plain text formatting (no markdown)
7. Are tailored to the specific job and company
8. Only output the contents of the cover letter

STYLE GUIDELINES:
- Professional but personable tone
- Specific examples over generic claims
- Show enthusiasm for the role
- Connect resume experience to job requirements
- Strong opening and closing

Date: {datetime.now().strftime("%Y-%m-%d")}
"""

    def execute(self, state: Dict[str, Any], trim_messages_fn) -> Dict[str, Any]:
        """Execute the cover letter writer node"""
        logger.info("📝 COVER LETTER WRITER NODE: Generating cover letter...")

        # Build system prompt for cover letter writing
        system_prompt = self._get_system_prompt(
            state["resume_name"], state["resume_context"]
        )

        if state.get("feedback_on_work"):
            logger.info(f"⚠️  Cover Letter Writer received feedback, retrying...")
            system_prompt += f"\n\nEVALUATOR FEEDBACK:\n{state['feedback_on_work']}\n\nPlease improve the cover letter based on this feedback."

        # Get recent conversation for context
        trimmed_messages = trim_messages_fn(state["messages"], max_tokens=6000)
        messages = [SystemMessage(content=system_prompt)] + trimmed_messages

        logger.info(f"   Message count: {len(messages)}")
        logger.info(f"   System prompt length: ~{len(system_prompt)} chars")

        response = self.llm.invoke(messages)

        logger.info(f"✓ Cover letter generated")
        logger.info(f"   Response length: {len(response.content)} chars")

        return {"messages": [response]}
